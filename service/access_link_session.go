package service

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/model"

	"github.com/gin-contrib/sessions"
)

const (
	sessionKeyAccessLinkGrantID   = "access_link_grant_id"
	sessionKeyAccessLinkExpiresAt = "access_link_expires_at"
)

type AccessLinkSessionState struct {
	Active    bool
	GrantID   uint
	ExpiresAt int64
}

func GetAccessLinkSessionState(session sessions.Session) AccessLinkSessionState {
	if session == nil {
		return AccessLinkSessionState{}
	}
	state := AccessLinkSessionState{
		GrantID:   uint(sessionAnyToInt(session.Get(sessionKeyAccessLinkGrantID))),
		ExpiresAt: sessionAnyToInt64(session.Get(sessionKeyAccessLinkExpiresAt)),
	}
	state.Active = state.GrantID != 0
	return state
}

func (state AccessLinkSessionState) HasSessionExpired(now time.Time) bool {
	if !state.Active || state.ExpiresAt <= 0 {
		return false
	}
	return now.Unix() >= state.ExpiresAt
}

func ClearAccessLinkSession(session sessions.Session) {
	if session == nil {
		return
	}
	session.Delete(sessionKeyAccessLinkGrantID)
	session.Delete(sessionKeyAccessLinkExpiresAt)
}

func BeginAccessLinkSession(session sessions.Session, targetUser *model.User, grant *model.ImpersonationGrant) error {
	if session == nil || targetUser == nil || grant == nil {
		return fmt.Errorf("invalid access link session payload")
	}

	ClearAccessLinkSession(session)
	ApplyAccessLinkWebSessionOptions(session)
	session.Set(sessionKeyAccessLinkGrantID, int(grant.Id))
	if grant.SessionExpiresAt != nil {
		session.Set(sessionKeyAccessLinkExpiresAt, grant.SessionExpiresAt.Unix())
	} else {
		session.Delete(sessionKeyAccessLinkExpiresAt)
	}
	SetAuthenticatedUserSession(session, targetUser)
	return session.Save()
}

func CompleteCurrentAccessLinkGrant(session sessions.Session) (*model.ImpersonationGrant, error) {
	state := GetAccessLinkSessionState(session)
	if !state.Active || state.GrantID == 0 {
		return nil, nil
	}

	grant, err := model.GetImpersonationGrantByID(state.GrantID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if grant.State == model.ImpersonationStateActive {
		grant.State = model.ImpersonationStateCompleted
		grant.EndedAt = &now
		if err = model.CompleteImpersonationGrant(grant.Id, now); err != nil {
			return nil, err
		}
	}

	return grant, nil
}

func StopCurrentAccessLinkSession(session sessions.Session, clearLogin bool) (*model.ImpersonationGrant, error) {
	if session == nil {
		return nil, nil
	}

	grant, err := CompleteCurrentAccessLinkGrant(session)
	if err != nil {
		return grant, err
	}

	ClearAccessLinkSession(session)
	if clearLogin {
		session.Clear()
	}
	ApplyDefaultWebSessionOptions(session)
	if err = session.Save(); err != nil {
		return grant, err
	}
	return grant, nil
}
