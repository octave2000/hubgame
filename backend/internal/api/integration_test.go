package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"hubgame/backend/internal/api"
	"hubgame/backend/internal/controller"
	"hubgame/backend/internal/database"
)

func TestGatewayIntegrationAuthIfMatchAndWebsocket(t *testing.T) {
	controllerSrv, _, gatewaySrv, cleanup := setupIntegrationStack(t)
	defer cleanup()

	token := issueToken(t, controllerSrv.URL, adminToken, "u-1", "t-1", "developer")

	createResp := requestJSON(t, http.MethodPost, gatewaySrv.URL+"/v1/entities", token, map[string]any{
		"id":   "m-1",
		"kind": "match",
		"data": map[string]any{"mode": "pvp", "status": "created"},
	}, nil)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for create, got %d", createResp.StatusCode)
	}
	createResp.Body.Close()

	conflictHeaders := map[string]string{"If-Match": "99"}
	conflictResp := requestJSON(t, http.MethodPatch, gatewaySrv.URL+"/v1/entities/m-1", token, map[string]any{
		"data": map[string]any{"mode": "pvp", "status": "running"},
	}, conflictHeaders)
	if conflictResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", conflictResp.StatusCode)
	}
	conflictResp.Body.Close()

	okHeaders := map[string]string{"If-Match": "1"}
	updateResp := requestJSON(t, http.MethodPatch, gatewaySrv.URL+"/v1/entities/m-1", token, map[string]any{
		"data": map[string]any{"mode": "pvp", "status": "running"},
	}, okHeaders)
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 update, got %d", updateResp.StatusCode)
	}
	updateResp.Body.Close()

	wsURL := strings.Replace(gatewaySrv.URL, "http://", "ws://", 1) + "/v1/events/stream?topic=room.chat"
	wsHeaders := http.Header{}
	wsHeaders.Set("Authorization", "Bearer "+token)
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, wsHeaders)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer wsConn.Close()
	_ = wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))

	eventResp := requestJSON(t, http.MethodPost, gatewaySrv.URL+"/v1/events", token, map[string]any{
		"topic":   "room.chat",
		"key":     "room-1",
		"type":    "chat.send",
		"payload": map[string]any{"room_id": "room-1", "message": "hello"},
	}, nil)
	if eventResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 event create, got %d", eventResp.StatusCode)
	}
	eventResp.Body.Close()

	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	var ev database.Event
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal websocket event: %v", err)
	}
	if ev.Type != "chat.send" {
		t.Fatalf("expected chat.send event type, got %q", ev.Type)
	}
	if ev.TenantID != "t-1" {
		t.Fatalf("expected tenant t-1, got %q", ev.TenantID)
	}
}

func TestGatewayIntegrationRBACAndUnauthorizedWebsocket(t *testing.T) {
	controllerSrv, _, gatewaySrv, cleanup := setupIntegrationStack(t)
	defer cleanup()

	playerToken := issueToken(t, controllerSrv.URL, adminToken, "u-2", "t-1", "player")
	devToken := issueToken(t, controllerSrv.URL, adminToken, "u-3", "t-1", "developer")

	createResp := requestJSON(t, http.MethodPost, gatewaySrv.URL+"/v1/entities", devToken, map[string]any{
		"id":   "room-1",
		"kind": "room",
		"data": map[string]any{"name": "Main Room"},
	}, nil)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 create for developer, got %d", createResp.StatusCode)
	}
	createResp.Body.Close()

	deleteResp := requestJSON(t, http.MethodDelete, gatewaySrv.URL+"/v1/entities/room-1", playerToken, nil, nil)
	if deleteResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for player delete, got %d", deleteResp.StatusCode)
	}
	deleteResp.Body.Close()

	wsURL := strings.Replace(gatewaySrv.URL, "http://", "ws://", 1) + "/v1/events/stream?topic=room.chat"

	_, unauthorizedResp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatalf("expected websocket dial failure without auth")
	}
	if unauthorizedResp == nil {
		t.Fatalf("expected HTTP response for unauthorized websocket handshake")
	}
	if unauthorizedResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 websocket handshake without auth, got %d", unauthorizedResp.StatusCode)
	}

	wsHeaders := http.Header{}
	wsHeaders.Set("Authorization", "Bearer "+playerToken)
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, wsHeaders)
	if err != nil {
		t.Fatalf("expected player websocket stream access, got dial error: %v", err)
	}
	_ = wsConn.Close()
}

func TestGatewayIntegrationLeaderboardGlobalAndGame(t *testing.T) {
	controllerSrv, _, gatewaySrv, cleanup := setupIntegrationStack(t)
	defer cleanup()

	devToken := issueToken(t, controllerSrv.URL, adminToken, "dev-1", "t-1", "developer")

	u1 := requestJSON(t, http.MethodPost, gatewaySrv.URL+"/v1/leaderboard/users", devToken, map[string]any{
		"user_id":      "u-1",
		"display_name": "Ada",
		"rank_title":   "Bronze",
	}, nil)
	if u1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 create leaderboard user, got %d", u1.StatusCode)
	}
	u1.Body.Close()

	u2 := requestJSON(t, http.MethodPost, gatewaySrv.URL+"/v1/leaderboard/users", devToken, map[string]any{
		"user_id":      "u-2",
		"display_name": "Turing",
		"rank_title":   "Silver",
	}, nil)
	if u2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 create leaderboard user 2, got %d", u2.StatusCode)
	}
	u2.Body.Close()

	s1 := requestJSON(t, http.MethodPost, gatewaySrv.URL+"/v1/leaderboard/scores", devToken, map[string]any{
		"game_id":        "tik-toe",
		"user_id":        "u-1",
		"score_delta":    12,
		"hubcoins_delta": 30,
	}, nil)
	if s1.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 score submit, got %d", s1.StatusCode)
	}
	s1.Body.Close()

	s2 := requestJSON(t, http.MethodPost, gatewaySrv.URL+"/v1/leaderboard/scores", devToken, map[string]any{
		"game_id":        "tik-toe",
		"user_id":        "u-2",
		"score_delta":    20,
		"hubcoins_delta": 10,
	}, nil)
	if s2.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 score submit 2, got %d", s2.StatusCode)
	}
	s2.Body.Close()

	gameLB := requestJSON(t, http.MethodGet, gatewaySrv.URL+"/v1/leaderboard?scope=game&game_id=tik-toe&limit=5", devToken, nil, nil)
	if gameLB.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 game leaderboard, got %d", gameLB.StatusCode)
	}
	var gamePayload struct {
		Items []struct {
			Rank        int    `json:"rank"`
			UserID      string `json:"user_id"`
			DisplayName string `json:"display_name"`
			Score       int    `json:"score"`
		} `json:"items"`
	}
	if err := json.NewDecoder(gameLB.Body).Decode(&gamePayload); err != nil {
		t.Fatalf("decode game leaderboard: %v", err)
	}
	gameLB.Body.Close()
	if len(gamePayload.Items) < 2 {
		t.Fatalf("expected at least 2 leaderboard rows, got %d", len(gamePayload.Items))
	}
	if gamePayload.Items[0].UserID != "u-2" {
		t.Fatalf("expected u-2 top rank, got %s", gamePayload.Items[0].UserID)
	}

	globalLB := requestJSON(t, http.MethodGet, gatewaySrv.URL+"/v1/leaderboard?scope=global&limit=5", devToken, nil, nil)
	if globalLB.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 global leaderboard, got %d", globalLB.StatusCode)
	}
	var globalPayload struct {
		Items []struct {
			UserID      string `json:"user_id"`
			GamesPlayed int    `json:"games_played"`
			Hubcoins    int    `json:"hubcoins"`
		} `json:"items"`
	}
	if err := json.NewDecoder(globalLB.Body).Decode(&globalPayload); err != nil {
		t.Fatalf("decode global leaderboard: %v", err)
	}
	globalLB.Body.Close()
	if len(globalPayload.Items) == 0 {
		t.Fatalf("expected global leaderboard rows")
	}
	if globalPayload.Items[0].GamesPlayed == 0 {
		t.Fatalf("expected non-zero games played for top user")
	}
}

func setupIntegrationStack(t *testing.T) (controllerSrv, dbEngineSrv, gatewaySrv *httptest.Server, cleanup func()) {
	t.Helper()

	ctx := context.Background()
	store, err := database.OpenSQLite(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	store.RegisterController(controller.SchemaController{})

	auth := controller.NewAuthController(secret, issuer)
	controllerSrv = httptest.NewServer(api.NewControllerService(auth, adminToken).Router())
	dbEngineSrv = httptest.NewServer(api.NewDBEngineServer(store, internalToken).Router())
	gatewaySrv = httptest.NewServer(api.NewGatewayServer(controllerSrv.URL, dbEngineSrv.URL, internalToken, adminToken, true).Router())

	cleanup = func() {
		gatewaySrv.Close()
		dbEngineSrv.Close()
		controllerSrv.Close()
		_ = store.Close()
	}
	return controllerSrv, dbEngineSrv, gatewaySrv, cleanup
}

const (
	secret        = "test-secret"
	issuer        = "test-issuer"
	internalToken = "internal-token"
	adminToken    = "admin-token"
)

func issueToken(t *testing.T, baseURL, adminToken, userID, tenantID, role string) string {
	t.Helper()
	resp := requestJSON(t, http.MethodPost, baseURL+"/v1/auth/token", "", map[string]any{
		"user_id":     userID,
		"tenant_id":   tenantID,
		"role":        role,
		"ttl_seconds": 3600,
	}, map[string]string{"X-Controller-Admin": adminToken})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("issue token failed with %d", resp.StatusCode)
	}
	var out map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	return out["token"]
}

func requestJSON(t *testing.T, method, target, bearer string, body any, headers map[string]string) *http.Response {
	t.Helper()
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}
	req, err := http.NewRequest(method, target, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}
