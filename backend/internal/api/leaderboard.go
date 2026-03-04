package api

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"hubgame/backend/internal/database"
)

const (
	kindLeaderboardUser  = "leaderboard_user"
	kindLeaderboardScore = "leaderboard_score"
)

type leaderboardUser struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	RankTitle   string    `json:"rank_title,omitempty"`
	Hubcoins    int       `json:"hubcoins"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type leaderboardScore struct {
	GameID      string    `json:"game_id"`
	UserID      string    `json:"user_id"`
	Score       int       `json:"score"`
	BestScore   int       `json:"best_score"`
	Hubcoins    int       `json:"hubcoins"`
	Submissions int       `json:"submissions"`
	RankTitle   string    `json:"rank_title,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type leaderboardRow struct {
	Rank        int    `json:"rank"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	GameID      string `json:"game_id,omitempty"`
	Score       int    `json:"score"`
	Hubcoins    int    `json:"hubcoins"`
	RankTitle   string `json:"rank_title,omitempty"`
	GamesPlayed int    `json:"games_played,omitempty"`
}

type leaderboardUserRequest struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	RankTitle   string `json:"rank_title"`
	Hubcoins    *int   `json:"hubcoins"`
}

type leaderboardScoreRequest struct {
	GameID        string `json:"game_id"`
	UserID        string `json:"user_id"`
	Score         *int   `json:"score"`
	ScoreDelta    *int   `json:"score_delta"`
	Hubcoins      *int   `json:"hubcoins"`
	HubcoinsDelta *int   `json:"hubcoins_delta"`
	RankTitle     string `json:"rank_title"`
}

func upsertLeaderboardUser(ctx context.Context, store *database.Store, tenantID string, req leaderboardUserRequest) (leaderboardUser, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return leaderboardUser{}, errors.New("user_id is required")
	}
	if strings.TrimSpace(req.DisplayName) == "" {
		req.DisplayName = req.UserID
	}

	entityID := leaderboardUserEntityID(req.UserID)
	existing, err := store.GetEntity(ctx, tenantID, entityID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return leaderboardUser{}, err
	}

	out := leaderboardUser{
		UserID:      req.UserID,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		RankTitle:   req.RankTitle,
		Hubcoins:    0,
		UpdatedAt:   time.Now().UTC(),
	}

	if existing != nil {
		var prev leaderboardUser
		if json.Unmarshal(existing.Data, &prev) == nil {
			if req.DisplayName == "" {
				out.DisplayName = prev.DisplayName
			}
			if req.AvatarURL == "" {
				out.AvatarURL = prev.AvatarURL
			}
			if req.RankTitle == "" {
				out.RankTitle = prev.RankTitle
			}
			out.Hubcoins = prev.Hubcoins
		}
	}

	if req.Hubcoins != nil {
		out.Hubcoins = *req.Hubcoins
	}
	if out.DisplayName == "" {
		out.DisplayName = out.UserID
	}

	data, _ := json.Marshal(out)
	if existing == nil {
		err = store.InsertEntity(ctx, &database.Entity{ID: entityID, TenantID: tenantID, Kind: kindLeaderboardUser, Data: data})
	} else {
		err = store.UpdateEntity(ctx, &database.Entity{ID: entityID, TenantID: tenantID, Data: data})
	}
	if err != nil {
		return leaderboardUser{}, err
	}
	return out, nil
}

func submitLeaderboardScore(ctx context.Context, store *database.Store, tenantID string, req leaderboardScoreRequest) (leaderboardScore, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return leaderboardScore{}, errors.New("user_id is required")
	}
	if strings.TrimSpace(req.GameID) == "" {
		return leaderboardScore{}, errors.New("game_id is required")
	}

	if _, err := upsertLeaderboardUser(ctx, store, tenantID, leaderboardUserRequest{
		UserID:      req.UserID,
		DisplayName: req.UserID,
	}); err != nil {
		return leaderboardScore{}, err
	}

	entityID := leaderboardScoreEntityID(req.GameID, req.UserID)
	existing, err := store.GetEntity(ctx, tenantID, entityID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return leaderboardScore{}, err
	}

	out := leaderboardScore{
		GameID:      req.GameID,
		UserID:      req.UserID,
		Score:       0,
		BestScore:   0,
		Hubcoins:    0,
		Submissions: 0,
		RankTitle:   req.RankTitle,
		UpdatedAt:   time.Now().UTC(),
	}
	if existing != nil {
		_ = json.Unmarshal(existing.Data, &out)
		out.UpdatedAt = time.Now().UTC()
		if req.RankTitle != "" {
			out.RankTitle = req.RankTitle
		}
	}

	if req.Score != nil {
		out.Score = *req.Score
	} else if req.ScoreDelta != nil {
		out.Score += *req.ScoreDelta
	}
	if req.Hubcoins != nil {
		out.Hubcoins = *req.Hubcoins
	} else if req.HubcoinsDelta != nil {
		out.Hubcoins += *req.HubcoinsDelta
	}
	if out.Score > out.BestScore {
		out.BestScore = out.Score
	}
	out.Submissions += 1

	data, _ := json.Marshal(out)
	if existing == nil {
		err = store.InsertEntity(ctx, &database.Entity{ID: entityID, TenantID: tenantID, Kind: kindLeaderboardScore, Data: data})
	} else {
		err = store.UpdateEntity(ctx, &database.Entity{ID: entityID, TenantID: tenantID, Data: data})
	}
	if err != nil {
		return leaderboardScore{}, err
	}

	_, _ = store.AppendEvent(ctx, database.Event{
		TenantID: tenantID,
		Topic:    "leaderboard." + req.GameID,
		Key:      req.UserID,
		Type:     "leaderboard.score_submitted",
		Payload:  data,
	})

	return out, nil
}

func queryLeaderboard(ctx context.Context, store *database.Store, tenantID, scope, gameID string, limit int) ([]leaderboardRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	profiles, err := loadLeaderboardUsers(ctx, store, tenantID)
	if err != nil {
		return nil, err
	}

	scores, err := store.ListEntities(ctx, tenantID, kindLeaderboardScore, 5000)
	if err != nil {
		return nil, err
	}

	if scope == "game" {
		rows := make([]leaderboardRow, 0, limit)
		for _, entity := range scores {
			var score leaderboardScore
			if json.Unmarshal(entity.Data, &score) != nil {
				continue
			}
			if score.GameID != gameID {
				continue
			}
			profile := profiles[score.UserID]
			rows = append(rows, leaderboardRow{
				UserID:      score.UserID,
				DisplayName: pickDisplayName(profile.DisplayName, score.UserID),
				GameID:      score.GameID,
				Score:       score.Score,
				Hubcoins:    score.Hubcoins,
				RankTitle:   pickDisplayName(score.RankTitle, profile.RankTitle),
			})
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Score == rows[j].Score {
				return rows[i].Hubcoins > rows[j].Hubcoins
			}
			return rows[i].Score > rows[j].Score
		})
		if len(rows) > limit {
			rows = rows[:limit]
		}
		for i := range rows {
			rows[i].Rank = i + 1
		}
		return rows, nil
	}

	agg := map[string]*leaderboardRow{}
	gameCount := map[string]map[string]struct{}{}
	for _, entity := range scores {
		var score leaderboardScore
		if json.Unmarshal(entity.Data, &score) != nil {
			continue
		}
		if _, ok := agg[score.UserID]; !ok {
			profile := profiles[score.UserID]
			agg[score.UserID] = &leaderboardRow{
				UserID:      score.UserID,
				DisplayName: pickDisplayName(profile.DisplayName, score.UserID),
				Score:       0,
				Hubcoins:    0,
				RankTitle:   pickDisplayName(profile.RankTitle, score.RankTitle),
			}
			gameCount[score.UserID] = map[string]struct{}{}
		}
		row := agg[score.UserID]
		row.Score += score.Score
		row.Hubcoins += score.Hubcoins
		gameCount[score.UserID][score.GameID] = struct{}{}
	}

	rows := make([]leaderboardRow, 0, len(agg))
	for userID, row := range agg {
		row.GamesPlayed = len(gameCount[userID])
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Score == rows[j].Score {
			return rows[i].Hubcoins > rows[j].Hubcoins
		}
		return rows[i].Score > rows[j].Score
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	for i := range rows {
		rows[i].Rank = i + 1
	}
	return rows, nil
}

func loadLeaderboardUsers(ctx context.Context, store *database.Store, tenantID string) (map[string]leaderboardUser, error) {
	entities, err := store.ListEntities(ctx, tenantID, kindLeaderboardUser, 5000)
	if err != nil {
		return nil, err
	}
	out := make(map[string]leaderboardUser, len(entities))
	for _, entity := range entities {
		var u leaderboardUser
		if json.Unmarshal(entity.Data, &u) == nil && u.UserID != "" {
			out[u.UserID] = u
		}
	}
	return out, nil
}

func leaderboardUserEntityID(userID string) string {
	return "user:" + strings.TrimSpace(userID)
}

func leaderboardScoreEntityID(gameID, userID string) string {
	return "lb:" + strings.TrimSpace(gameID) + ":" + strings.TrimSpace(userID)
}

func pickDisplayName(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}
