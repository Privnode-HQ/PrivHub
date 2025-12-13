package model

import "testing"

func TestConsumeFirstUsableSubscriptionItem_SelectsFirstMatchOnly(t *testing.T) {
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

	updated, ok := consumeFirstUsableSubscriptionItem(items, now, 1)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if updated[0].Limit5H.Available != 1 || updated[0].Limit7D.Available != 0 {
		t.Fatalf("unexpected change in first item: %+v", updated[0])
	}
	if updated[1].Limit5H.Available != 1 || updated[1].Limit7D.Available != 1 {
		t.Fatalf("expected second item decremented by 1, got %+v", updated[1])
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
	if updated[1].Limit5H.Available != 10 || updated[1].Limit7D.Available != 20 {
		t.Fatalf("expected reset even when reset_at=0, got %+v", updated[1])
	}
}
