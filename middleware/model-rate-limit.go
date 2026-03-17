package middleware

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

// ModelRequestRateLimit 用户分组使用限制中间件
func ModelRequestRateLimit() func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := service.ReserveUsageRequest(c); err != nil {
			if service.IsUsageLimitExceededError(err) {
				abortWithOpenAiMessage(c, http.StatusTooManyRequests, err.Error(), "group_usage_limit_exceeded")
				return
			}
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "usage_limit_check_failed")
			return
		}

		c.Next()

		if c.Writer.Status() >= http.StatusBadRequest {
			_ = service.ReleaseUsageReservation(&relaycommon.RelayInfo{
				UserId:             common.GetContextKeyInt(c, constant.ContextKeyUserId),
				UsageReservationID: common.GetContextKeyString(c, constant.ContextKeyUsageReservationID),
			})
		}
	}
}
