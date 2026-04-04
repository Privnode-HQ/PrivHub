package model

import (
	"testing"
	"time"
)

func TestImpersonationGrantAllowsRequestedAccess(t *testing.T) {
	fullGrant := &ImpersonationGrant{RequestedReadOnly: false}
	if !fullGrant.AllowsRequestedAccess(false) {
		t.Fatalf("expected full grant to allow full-access activation")
	}
	if !fullGrant.AllowsRequestedAccess(true) {
		t.Fatalf("expected full grant to allow read-only activation")
	}

	readOnlyGrant := &ImpersonationGrant{RequestedReadOnly: true}
	if !readOnlyGrant.AllowsRequestedAccess(true) {
		t.Fatalf("expected read-only grant to allow read-only activation")
	}
	if readOnlyGrant.AllowsRequestedAccess(false) {
		t.Fatalf("expected read-only grant to reject full-access activation")
	}
}

func TestExpireImpersonationGrantIfNeeded(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)

	approvedGrant := &ImpersonationGrant{
		State:            ImpersonationStateApproved,
		GrantedExpiresAt: ptrTime(now.Add(-time.Minute)),
	}
	if !ExpireImpersonationGrantIfNeeded(approvedGrant, now) {
		t.Fatalf("expected approved grant to expire when activation window is missed")
	}
	if approvedGrant.State != ImpersonationStateExpired {
		t.Fatalf("expected state=%s, got %s", ImpersonationStateExpired, approvedGrant.State)
	}

	activeGrant := &ImpersonationGrant{
		State:            ImpersonationStateActive,
		SessionExpiresAt: ptrTime(now.Add(-time.Minute)),
	}
	if !ExpireImpersonationGrantIfNeeded(activeGrant, now) {
		t.Fatalf("expected active grant to complete when session window is exceeded")
	}
	if activeGrant.State != ImpersonationStateCompleted {
		t.Fatalf("expected state=%s, got %s", ImpersonationStateCompleted, activeGrant.State)
	}
	if activeGrant.EndedAt == nil || !activeGrant.EndedAt.Equal(now) {
		t.Fatalf("expected ended_at to be set to now, got %+v", activeGrant.EndedAt)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
