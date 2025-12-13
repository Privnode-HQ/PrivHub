package model

import (
	"strings"
	"testing"
)

func TestConsumeSubscriptionQuotaFromFirstUsableItem_SelectsFirstMatchOnly(t *testing.T) {
	now := int64(1_700_000_000)
	items := []SubscriptionItem{
		{
			Status: "deployed",
			Limit5H: SubscriptionLimit{
				Total:     10,
				Available: 1,
				ResetAt:   now + 100,
			},
			Limit7D: SubscriptionLimit{
				Total:     10,
				Available: 0,
				ResetAt:   now + 100,
			},
			Duration: SubscriptionDuration{
				StartAt: now - 10,
				EndAt:   now + 10,
			},
		},
		{
			Status: "deployed",
			Limit5H: SubscriptionLimit{
				Total:     10,
				Available: 2,
				ResetAt:   now + 100,
			},
			Limit7D: SubscriptionLimit{
				Total:     10,
				Available: 2,
				ResetAt:   now + 100,
			},
			Duration: SubscriptionDuration{
				StartAt: now - 10,
				EndAt:   now + 10,
			},
		},
	}

	updated, token, ok := consumeSubscriptionQuotaFromFirstUsableItem(items, now, 2)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if !strings.HasPrefix(token, "fp:") {
		t.Fatalf("expected fingerprint token, got %q", token)
	}
	if updated[0].Limit5H.Available != 1 || updated[0].Limit7D.Available != 0 {
		t.Fatalf("unexpected change in first item: %+v", updated[0])
	}
	if updated[1].Limit5H.Available != 0 || updated[1].Limit7D.Available != 0 {
		t.Fatalf("expected second item decremented by 2, got %+v", updated[1])
	}
}

func TestDurationContains_RequiresStartAndEnd(t *testing.T) {
	now := int64(1_700_000_000)
	if durationContains(SubscriptionDuration{StartAt: 0, EndAt: now + 10}, now) {
		t.Fatalf("expected false when StartAt is 0")
	}
	if durationContains(SubscriptionDuration{StartAt: now - 10, EndAt: 0}, now) {
		t.Fatalf("expected false when EndAt is 0")
	}
}

func TestDurationContains_MillisTimestamp(t *testing.T) {
	nowSec := int64(1_700_000_000)
	startMs := nowSec*1000 - 1000
	endMs := nowSec*1000 + 1000
	if !durationContains(SubscriptionDuration{StartAt: startMs, EndAt: endMs}, nowSec) {
		t.Fatalf("expected durationContains to be true for millisecond timestamps")
	}
}

func TestResetAndPruneSubscriptionData_PruneNonDeployedAndReset(t *testing.T) {
	now := int64(1_700_000_000)
	items := []SubscriptionItem{
		{Status: "deploying"},
		{Status: "disabled"},
		{
			Status:  "deployed",
			Limit5H: SubscriptionLimit{Total: 10, Available: 3, ResetAt: now - 1},
			Limit7D: SubscriptionLimit{Total: 20, Available: 4, ResetAt: now - 1},
		},
		{
			Status:  "deployed",
			Limit5H: SubscriptionLimit{Total: 10, Available: 0, ResetAt: 0},
			Limit7D: SubscriptionLimit{Total: 20, Available: 0, ResetAt: 0},
		},
	}

	updated, changed := resetAndPruneSubscriptionData(items, now)
	if !changed {
		t.Fatalf("expected changed=true")
	}
	if len(updated) != 2 {
		t.Fatalf("expected only deployed items kept, got len=%d", len(updated))
	}
	if updated[0].Limit5H.Available != 10 || updated[0].Limit7D.Available != 20 {
		t.Fatalf("expected first deployed item reset to total, got %+v", updated[0])
	}
	if updated[0].Limit5H.ResetAt <= now || updated[0].Limit7D.ResetAt <= now {
		t.Fatalf("expected reset_at to advance beyond now, got %+v", updated[0])
	}
	if updated[1].Limit5H.Available != 10 || updated[1].Limit7D.Available != 20 {
		t.Fatalf("expected reset even when reset_at=0, got %+v", updated[1])
	}
	if updated[1].Limit5H.ResetAt != now+5*60*60 {
		t.Fatalf("expected 5h reset_at=%d, got %d", now+5*60*60, updated[1].Limit5H.ResetAt)
	}
	if updated[1].Limit7D.ResetAt != now+7*24*60*60 {
		t.Fatalf("expected 7d reset_at=%d, got %d", now+7*24*60*60, updated[1].Limit7D.ResetAt)
	}
}

func TestIsUsableSubscriptionItem_RequiresEnoughForAmount(t *testing.T) {
	now := int64(1_700_000_000)
	item := SubscriptionItem{
		Status:  "deployed",
		Limit5H: SubscriptionLimit{Total: 10, Available: 2},
		Limit7D: SubscriptionLimit{Total: 10, Available: 2},
		Duration: SubscriptionDuration{
			StartAt: now - 10,
			EndAt:   now + 10,
		},
	}
	if isUsableSubscriptionItem(item, now, 3) {
		t.Fatalf("expected unusable when amount exceeds available")
	}
	if !isUsableSubscriptionItem(item, now, 2) {
		t.Fatalf("expected usable when amount fits")
	}
}

func TestParseSubscriptionData_AllowsStringTimestampsAndQuotas(t *testing.T) {
	raw := `[
		{
			"status": "deployed",
			"5h_limit": {"total": "10", "available": "9", "reset_at": "1700000000"},
			"7d_limit": {"total": 20, "available": 19, "reset_at": 1700000000},
			"duration": {"start_at": 1700000000, "end_at": "1700003600", "auto_renew_enabled": true}
		}
	]`

	items, err := parseSubscriptionData(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Duration.EndAt != 1700003600 {
		t.Fatalf("expected EndAt=1700003600, got %d", items[0].Duration.EndAt)
	}
	if items[0].Limit5H.Total != 10 || items[0].Limit5H.Available != 9 || items[0].Limit5H.ResetAt != 1700000000 {
		t.Fatalf("unexpected 5h_limit: %+v", items[0].Limit5H)
	}
}
