package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-contrib/sessions"
)

const (
	ImpersonationGrantWindow   = 24 * time.Hour
	ImpersonationSessionWindow = 24 * time.Hour

	sessionKeyImpersonationGrantID            = "impersonation_grant_id"
	sessionKeyImpersonationOriginalID         = "impersonation_original_id"
	sessionKeyImpersonationOriginalUsername   = "impersonation_original_username"
	sessionKeyImpersonationOriginalCAHID      = "impersonation_original_cah_id"
	sessionKeyImpersonationOriginalRole       = "impersonation_original_role"
	sessionKeyImpersonationOriginalStatus     = "impersonation_original_status"
	sessionKeyImpersonationOriginalGroup      = "impersonation_original_group"
	sessionKeyImpersonationOriginalSessionVer = "impersonation_original_session_version"
	sessionKeyImpersonationOriginalGlobalVer  = "impersonation_original_global_session_version"
	sessionKeyImpersonationReadOnly           = "impersonation_read_only"
	sessionKeyImpersonationBreakGlass         = "impersonation_break_glass"
	sessionKeyImpersonationStartedAt          = "impersonation_started_at"
	sessionKeyImpersonationExpiresAt          = "impersonation_expires_at"
	sessionKeyImpersonationHeaderAliasID      = "impersonation_header_alias_id"
	sessionKeyImpersonationHeaderAliasCAHID   = "impersonation_header_alias_cah_id"
)

type ImpersonationSessionState struct {
	Active             bool
	GrantID            uint
	OriginalID         int
	OriginalUsername   string
	OriginalCAHID      string
	OriginalRole       int
	OriginalStatus     int
	OriginalGroup      string
	OriginalSessionVer int
	OriginalGlobalVer  int
	ReadOnly           bool
	BreakGlass         bool
	StartedAt          int64
	ExpiresAt          int64
	HeaderAliasID      int
	HeaderAliasCAHID   string
}

func sessionAnyToInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		intValue, _ := strconv.Atoi(v)
		return intValue
	default:
		return 0
	}
}

func sessionAnyToInt64(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		intValue, _ := strconv.ParseInt(v, 10, 64)
		return intValue
	default:
		return 0
	}
}

func sessionAnyToBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

func sessionAnyToString(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func GetImpersonationSessionState(session sessions.Session) ImpersonationSessionState {
	state := ImpersonationSessionState{
		GrantID:            uint(sessionAnyToInt(session.Get(sessionKeyImpersonationGrantID))),
		OriginalID:         sessionAnyToInt(session.Get(sessionKeyImpersonationOriginalID)),
		OriginalUsername:   sessionAnyToString(session.Get(sessionKeyImpersonationOriginalUsername)),
		OriginalCAHID:      sessionAnyToString(session.Get(sessionKeyImpersonationOriginalCAHID)),
		OriginalRole:       sessionAnyToInt(session.Get(sessionKeyImpersonationOriginalRole)),
		OriginalStatus:     sessionAnyToInt(session.Get(sessionKeyImpersonationOriginalStatus)),
		OriginalGroup:      sessionAnyToString(session.Get(sessionKeyImpersonationOriginalGroup)),
		OriginalSessionVer: sessionAnyToInt(session.Get(sessionKeyImpersonationOriginalSessionVer)),
		OriginalGlobalVer:  sessionAnyToInt(session.Get(sessionKeyImpersonationOriginalGlobalVer)),
		ReadOnly:           sessionAnyToBool(session.Get(sessionKeyImpersonationReadOnly)),
		BreakGlass:         sessionAnyToBool(session.Get(sessionKeyImpersonationBreakGlass)),
		StartedAt:          sessionAnyToInt64(session.Get(sessionKeyImpersonationStartedAt)),
		ExpiresAt:          sessionAnyToInt64(session.Get(sessionKeyImpersonationExpiresAt)),
		HeaderAliasID:      sessionAnyToInt(session.Get(sessionKeyImpersonationHeaderAliasID)),
		HeaderAliasCAHID:   sessionAnyToString(session.Get(sessionKeyImpersonationHeaderAliasCAHID)),
	}
	state.Active = state.OriginalID != 0
	return state
}

func (state ImpersonationSessionState) HasSessionExpired(now time.Time) bool {
	if !state.Active || state.BreakGlass || state.ExpiresAt <= 0 {
		return false
	}
	return now.Unix() >= state.ExpiresAt
}

func ClearImpersonationSession(session sessions.Session) {
	keys := []string{
		sessionKeyImpersonationGrantID,
		sessionKeyImpersonationOriginalID,
		sessionKeyImpersonationOriginalUsername,
		sessionKeyImpersonationOriginalCAHID,
		sessionKeyImpersonationOriginalRole,
		sessionKeyImpersonationOriginalStatus,
		sessionKeyImpersonationOriginalGroup,
		sessionKeyImpersonationOriginalSessionVer,
		sessionKeyImpersonationOriginalGlobalVer,
		sessionKeyImpersonationReadOnly,
		sessionKeyImpersonationBreakGlass,
		sessionKeyImpersonationStartedAt,
		sessionKeyImpersonationExpiresAt,
	}
	for _, key := range keys {
		session.Delete(key)
	}
}

func ClearImpersonationHeaderAlias(session sessions.Session) {
	session.Delete(sessionKeyImpersonationHeaderAliasID)
	session.Delete(sessionKeyImpersonationHeaderAliasCAHID)
}

func BeginImpersonationSession(session sessions.Session, originalUser *model.User, targetUser *model.User, grant *model.ImpersonationGrant, readOnly bool) error {
	if session == nil || originalUser == nil || targetUser == nil || grant == nil {
		return fmt.Errorf("invalid impersonation session payload")
	}

	ClearImpersonationSession(session)
	ClearImpersonationHeaderAlias(session)

	session.Set(sessionKeyImpersonationGrantID, int(grant.Id))
	session.Set(sessionKeyImpersonationOriginalID, originalUser.Id)
	session.Set(sessionKeyImpersonationOriginalUsername, originalUser.Username)
	session.Set(sessionKeyImpersonationOriginalCAHID, originalUser.CAHID)
	session.Set(sessionKeyImpersonationOriginalRole, originalUser.Role)
	session.Set(sessionKeyImpersonationOriginalStatus, originalUser.Status)
	session.Set(sessionKeyImpersonationOriginalGroup, originalUser.Group)
	session.Set(sessionKeyImpersonationOriginalSessionVer, originalUser.WebSessionVersion)
	session.Set(sessionKeyImpersonationOriginalGlobalVer, common.GlobalWebSessionVersion)
	session.Set(sessionKeyImpersonationReadOnly, readOnly)
	session.Set(sessionKeyImpersonationBreakGlass, grant.IsBreakGlass())
	startedAt := time.Now().Unix()
	session.Set(sessionKeyImpersonationStartedAt, startedAt)
	if grant.SessionExpiresAt != nil {
		session.Set(sessionKeyImpersonationExpiresAt, grant.SessionExpiresAt.Unix())
	} else {
		session.Delete(sessionKeyImpersonationExpiresAt)
	}

	session.Set("id", targetUser.Id)
	session.Set("username", targetUser.Username)
	session.Set("cah_id", targetUser.CAHID)
	session.Set("role", targetUser.Role)
	session.Set("status", targetUser.Status)
	session.Set("group", targetUser.Group)
	session.Set("session_version", targetUser.WebSessionVersion)
	session.Set("global_session_version", common.GlobalWebSessionVersion)

	return session.Save()
}

func RestoreOriginalSession(session sessions.Session, preserveHeaderAlias bool) (bool, error) {
	state := GetImpersonationSessionState(session)
	if !state.Active || state.OriginalID == 0 {
		return false, nil
	}

	if preserveHeaderAlias {
		session.Set(sessionKeyImpersonationHeaderAliasID, sessionAnyToInt(session.Get("id")))
		session.Set(sessionKeyImpersonationHeaderAliasCAHID, sessionAnyToString(session.Get("cah_id")))
	} else {
		ClearImpersonationHeaderAlias(session)
	}

	session.Set("id", state.OriginalID)
	session.Set("username", state.OriginalUsername)
	session.Set("cah_id", state.OriginalCAHID)
	session.Set("role", state.OriginalRole)
	session.Set("status", state.OriginalStatus)
	session.Set("group", state.OriginalGroup)
	session.Set("session_version", state.OriginalSessionVer)
	session.Set("global_session_version", state.OriginalGlobalVer)

	ClearImpersonationSession(session)
	return true, session.Save()
}

func IsImpersonationStopPath(path string) bool {
	return path == "/api/user/impersonation/stop" || path == "/api/user/logout"
}

func MatchesImpersonationHeaderAlias(session sessions.Session, identifier string) bool {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return false
	}
	state := GetImpersonationSessionState(session)
	if state.HeaderAliasID != 0 && identifier == strconv.Itoa(state.HeaderAliasID) {
		return true
	}
	normalizedCAHID := model.NormalizeCAHID(identifier)
	return normalizedCAHID != "" && state.HeaderAliasCAHID != "" && normalizedCAHID == model.NormalizeCAHID(state.HeaderAliasCAHID)
}

func IsReadOnlyImpersonation(session sessions.Session) bool {
	state := GetImpersonationSessionState(session)
	return state.Active && state.ReadOnly
}
