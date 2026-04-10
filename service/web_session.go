package service

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-contrib/sessions"
)

const (
	DefaultWebSessionMaxAgeSeconds = 30 * 24 * 60 * 60
	AccessLinkSessionMaxAgeSeconds = 24 * 60 * 60
)

func WebSessionOptions(maxAge int) sessions.Options {
	return sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}

func ApplyDefaultWebSessionOptions(session sessions.Session) {
	if session == nil {
		return
	}
	session.Options(WebSessionOptions(DefaultWebSessionMaxAgeSeconds))
}

func ApplyAccessLinkWebSessionOptions(session sessions.Session) {
	if session == nil {
		return
	}
	session.Options(WebSessionOptions(AccessLinkSessionMaxAgeSeconds))
}

func SetAuthenticatedUserSession(session sessions.Session, user *model.User) {
	if session == nil || user == nil {
		return
	}
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("cah_id", user.CAHID)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	session.Set("group", user.Group)
	session.Set("session_version", user.WebSessionVersion)
	session.Set("global_session_version", common.GlobalWebSessionVersion)
}
