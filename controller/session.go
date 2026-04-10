package controller

import (
	"errors"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func sessionValueToInt(value any) int {
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

func getValidatedSessionUserID(c *gin.Context) (int, error) {
	session := sessions.Default(c)
	idRaw := session.Get("id")
	if idRaw == nil {
		return 0, errors.New("未登录或登录已过期")
	}
	userId := sessionValueToInt(idRaw)
	if userId == 0 {
		return 0, errors.New("未登录或登录已过期")
	}

	userCache, err := model.GetUserCache(userId)
	if err != nil {
		return 0, err
	}
	if accessLinkState := service.GetAccessLinkSessionState(session); accessLinkState.Active && accessLinkState.HasSessionExpired(time.Now()) {
		_, _ = service.CompleteCurrentAccessLinkGrant(session)
		session.Clear()
		service.ApplyDefaultWebSessionOptions(session)
		_ = session.Save()
		return 0, errors.New("登录已失效，请重新登录")
	}

	if sessionValueToInt(session.Get("session_version")) != userCache.WebSessionVersion ||
		sessionValueToInt(session.Get("global_session_version")) != common.GlobalWebSessionVersion {
		session.Clear()
		_ = session.Save()
		return 0, errors.New("登录已失效，请重新登录")
	}

	return userId, nil
}
