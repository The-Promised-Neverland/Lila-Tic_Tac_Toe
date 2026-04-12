package game

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

func GenerateInviteCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const codeLength = 6

	buf := make([]byte, codeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate invite code: %w", err)
	}

	out := make([]byte, codeLength)
	for i, b := range buf {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}

func NormalizeInviteCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func WriteInviteCode(ctx context.Context, nk runtime.NakamaModule, inviteCode, matchID string) error {
	value, err := MustJSON(map[string]string{"match_id": matchID})
	if err != nil {
		return err
	}
	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{{
		Collection:      RoomCodeCollection,
		Key:             NormalizeInviteCode(inviteCode),
		UserID:          "",
		Value:           value,
		PermissionRead:  2,
		PermissionWrite: 0,
	}})
	if err != nil {
		return fmt.Errorf("write invite code: %w", err)
	}
	return nil
}

func DeleteInviteCode(ctx context.Context, nk runtime.NakamaModule, inviteCode string) error {
	if strings.TrimSpace(inviteCode) == "" {
		return nil
	}

	return nk.StorageDelete(ctx, []*runtime.StorageDelete{{
		Collection: RoomCodeCollection,
		Key:        NormalizeInviteCode(inviteCode),
		UserID:     "",
	}})
}

func LookupMatchIDByInviteCode(ctx context.Context, nk runtime.NakamaModule, inviteCode string) (string, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{{
		Collection: RoomCodeCollection,
		Key:        NormalizeInviteCode(inviteCode),
		UserID:     "",
	}})
	if err != nil {
		return "", fmt.Errorf("lookup invite code: %w", err)
	}
	if len(objects) == 0 {
		return "", errors.New("invite code not found")
	}

	var payload struct {
		MatchID string `json:"match_id"`
	}
	if err := json.Unmarshal([]byte(objects[0].Value), &payload); err != nil {
		return "", errors.New("invalid invite code data")
	}
	if payload.MatchID == "" {
		return "", errors.New("invite code is not ready yet")
	}
	return payload.MatchID, nil
}

func ReadPlayerStats(ctx context.Context, nk runtime.NakamaModule, userID, username string) (PlayerStats, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{{
		Collection: StatsCollection,
		Key:        StatsKey,
		UserID:     userID,
	}})
	if err != nil {
		return PlayerStats{}, fmt.Errorf("read player stats: %w", err)
	}

	stats := PlayerStats{
		UserID:   userID,
		Username: username,
	}
	if len(objects) == 0 {
		return stats, nil
	}

	if err := json.Unmarshal([]byte(objects[0].Value), &stats); err != nil {
		return PlayerStats{}, fmt.Errorf("decode player stats: %w", err)
	}
	if stats.UserID == "" {
		stats.UserID = userID
	}
	if stats.Username == "" {
		stats.Username = username
	}
	return stats, nil
}

func WritePlayerStats(ctx context.Context, nk runtime.NakamaModule, stats PlayerStats) error {
	value, err := MustJSON(stats)
	if err != nil {
		return err
	}

	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{{
		Collection:      StatsCollection,
		Key:             StatsKey,
		UserID:          stats.UserID,
		Value:           value,
		PermissionRead:  2,
		PermissionWrite: 0,
	}})
	if err != nil {
		return fmt.Errorf("write player stats: %w", err)
	}
	return nil
}

func ReadMatchHistory(ctx context.Context, nk runtime.NakamaModule, userID string) ([]MatchHistoryEntry, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{{
		Collection: MatchHistoryCollection,
		Key:        MatchHistoryKey,
		UserID:     userID,
	}})
	if err != nil {
		return nil, fmt.Errorf("read match history: %w", err)
	}
	if len(objects) == 0 {
		return []MatchHistoryEntry{}, nil
	}

	var history []MatchHistoryEntry
	if err := json.Unmarshal([]byte(objects[0].Value), &history); err != nil {
		return nil, fmt.Errorf("decode match history: %w", err)
	}
	return history, nil
}

func AppendMatchHistory(ctx context.Context, nk runtime.NakamaModule, userID string, entry MatchHistoryEntry) error {
	history, err := ReadMatchHistory(ctx, nk, userID)
	if err != nil {
		return err
	}

	history = append([]MatchHistoryEntry{entry}, history...)
	if len(history) > 50 {
		history = history[:50]
	}

	value, err := MustJSON(history)
	if err != nil {
		return err
	}

	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{{
		Collection:      MatchHistoryCollection,
		Key:             MatchHistoryKey,
		UserID:          userID,
		Value:           value,
		PermissionRead:  2,
		PermissionWrite: 0,
	}})
	if err != nil {
		return fmt.Errorf("write match history: %w", err)
	}
	return nil
}
