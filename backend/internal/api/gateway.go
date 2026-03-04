package api

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"hubgame/backend/internal/controller"
	"hubgame/backend/internal/controllerclient"
)

type gatewayClaimsKey struct{}

type GatewayServer struct {
	controller  *controllerclient.Client
	authorizer  *controller.Authorizer
	dbEngineURL string
	internalKey string
	http        *http.Client
}

func NewGatewayServer(controllerURL, dbEngineURL, internalKey string) *GatewayServer {
	return &GatewayServer{
		controller:  controllerclient.New(controllerURL),
		authorizer:  controller.NewAuthorizer(),
		dbEngineURL: strings.TrimRight(dbEngineURL, "/"),
		internalKey: internalKey,
		http:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *GatewayServer) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "gateway"})
	})
	mux.Handle("/v1/events/stream", g.requireAuth(g.requireAction(controller.ActionStreamRead, http.HandlerFunc(g.streamProxy))))
	mux.Handle("/v1/entities", g.requireAuth(http.HandlerFunc(g.entitiesProxy)))
	mux.Handle("/v1/entities/", g.requireAuth(http.HandlerFunc(g.entityByIDProxy)))
	mux.Handle("/v1/events", g.requireAuth(http.HandlerFunc(g.eventsProxy)))
	return mux
}

func (g *GatewayServer) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		claims, err := g.controller.VerifyToken(r.Context(), parts[1])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(withGatewayClaims(r.Context(), claims))
		next.ServeHTTP(w, r)
	})
}

func (g *GatewayServer) requireAction(action string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := gatewayClaimsFromContext(r.Context())
		if err := g.authorizer.Enforce(claims, action); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (g *GatewayServer) entitiesProxy(w http.ResponseWriter, r *http.Request) {
	claims, _ := gatewayClaimsFromContext(r.Context())
	if r.Method == http.MethodGet {
		if err := g.authorizer.Enforce(claims, controller.ActionEntityRead); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	} else if r.Method == http.MethodPost {
		if err := g.authorizer.Enforce(claims, controller.ActionEntityWrite); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	g.proxyHTTP(w, r, claims.TenantID)
}

func (g *GatewayServer) entityByIDProxy(w http.ResponseWriter, r *http.Request) {
	claims, _ := gatewayClaimsFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		if err := g.authorizer.Enforce(claims, controller.ActionEntityRead); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	case http.MethodPatch:
		if err := g.authorizer.Enforce(claims, controller.ActionEntityWrite); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	case http.MethodDelete:
		if err := g.authorizer.Enforce(claims, controller.ActionEntityDelete); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	g.proxyHTTP(w, r, claims.TenantID)
}

func (g *GatewayServer) eventsProxy(w http.ResponseWriter, r *http.Request) {
	claims, _ := gatewayClaimsFromContext(r.Context())
	if r.Method == http.MethodGet {
		if err := g.authorizer.Enforce(claims, controller.ActionEventRead); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	} else if r.Method == http.MethodPost {
		if err := g.authorizer.Enforce(claims, controller.ActionEventWrite); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	g.proxyHTTP(w, r, claims.TenantID)
}

func (g *GatewayServer) proxyHTTP(w http.ResponseWriter, r *http.Request, tenantID string) {
	target := g.dbEngineURL + r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	req.Header.Set("If-Match", r.Header.Get("If-Match"))
	req.Header.Set("X-Internal-Token", g.internalKey)
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := g.http.Do(req)
	if err != nil {
		http.Error(w, "db-engine unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (g *GatewayServer) streamProxy(w http.ResponseWriter, r *http.Request) {
	claims, _ := gatewayClaimsFromContext(r.Context())
	clientConn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		return
	}
	defer clientConn.Close()

	upstreamURL, err := url.Parse(g.dbEngineURL)
	if err != nil {
		return
	}
	if upstreamURL.Scheme == "https" {
		upstreamURL.Scheme = "wss"
	} else {
		upstreamURL.Scheme = "ws"
	}
	upstreamURL.Path = "/v1/events/stream"
	upstreamURL.RawQuery = r.URL.RawQuery

	headers := http.Header{}
	headers.Set("X-Internal-Token", g.internalKey)
	headers.Set("X-Tenant-ID", claims.TenantID)
	upstreamConn, _, err := websocket.DefaultDialer.Dial(upstreamURL.String(), headers)
	if err != nil {
		return
	}
	defer upstreamConn.Close()

	errCh := make(chan error, 2)
	go bridgeWebsocket(clientConn, upstreamConn, errCh)
	go bridgeWebsocket(upstreamConn, clientConn, errCh)
	<-errCh
}

func bridgeWebsocket(src, dst *websocket.Conn, errCh chan<- error) {
	for {
		mt, msg, err := src.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}
		if err := dst.WriteMessage(mt, msg); err != nil {
			errCh <- err
			return
		}
	}
}

func withGatewayClaims(ctx context.Context, claims *controller.Claims) context.Context {
	return context.WithValue(ctx, gatewayClaimsKey{}, claims)
}

func gatewayClaimsFromContext(ctx context.Context) (*controller.Claims, bool) {
	claims, ok := ctx.Value(gatewayClaimsKey{}).(*controller.Claims)
	return claims, ok
}
