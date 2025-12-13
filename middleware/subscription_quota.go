package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func SubscriptionQuotaForClaudeBetaMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("beta") != "true" {
			c.Next()
			return
		}
		enforceSubscriptionQuotaOrAbort(c, true)
	}
}

func SubscriptionQuotaForOpenAIResponses() gin.HandlerFunc {
	return func(c *gin.Context) {
		enforceSubscriptionQuotaOrAbort(c, false)
	}
}

func enforceSubscriptionQuotaOrAbort(c *gin.Context, isClaude bool) {
	userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	if userGroup != "subscription" {
		c.Next()
		return
	}

	userID := common.GetContextKeyInt(c, constant.ContextKeyUserId)
	err := model.ConsumeUserSubscriptionQuota(userID, time.Now().Unix(), 1)
	if err == nil {
		c.Next()
		return
	}

	var apiErr *types.NewAPIError
	switch {
	case errors.Is(err, model.ErrSubscriptionQuotaExhausted):
		if isClaude {
			apiErr = types.WithClaudeError(
				types.ClaudeError{Type: string(types.ErrorCodeSubscriptionQuotaExhausted), Message: err.Error()},
				http.StatusForbidden,
				types.ErrOptionWithSkipRetry(),
				types.ErrOptionWithNoRecordErrorLog(),
			)
		} else {
			apiErr = types.NewErrorWithStatusCode(
				err,
				types.ErrorCodeSubscriptionQuotaExhausted,
				http.StatusForbidden,
				types.ErrOptionWithSkipRetry(),
				types.ErrOptionWithNoRecordErrorLog(),
			)
		}
	default:
		apiErr = types.NewErrorWithStatusCode(
			err,
			types.ErrorCodeUpdateDataError,
			http.StatusInternalServerError,
			types.ErrOptionWithSkipRetry(),
		)
	}

	if isClaude {
		c.JSON(apiErr.StatusCode, gin.H{
			"type":  "error",
			"error": apiErr.ToClaudeError(),
		})
	} else {
		c.JSON(apiErr.StatusCode, gin.H{
			"error": apiErr.ToOpenAIError(),
		})
	}

	c.Abort()
}
