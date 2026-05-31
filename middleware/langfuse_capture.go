package middleware

import (
	"bytes"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

const langfuseResponseCaptureLimit = 512 * 1024

type langfuseCaptureWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
	max  int
}

func (w *langfuseCaptureWriter) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *langfuseCaptureWriter) WriteString(data string) (int, error) {
	w.capture([]byte(data))
	return w.ResponseWriter.WriteString(data)
}

func (w *langfuseCaptureWriter) capture(data []byte) {
	if w.max <= 0 || len(data) == 0 || w.body.Len() >= w.max {
		return
	}
	remaining := w.max - w.body.Len()
	if len(data) > remaining {
		data = data[:remaining]
	}
	_, _ = w.body.Write(data)
}

func LangfuseCapture() func(c *gin.Context) {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		startedAt := time.Now()
		writer := &langfuseCaptureWriter{
			ResponseWriter: c.Writer,
			max:            langfuseResponseCaptureLimit,
		}
		c.Writer = writer
		c.Next()

		group := common.GetContextKeyString(c, constant.ContextKeyUsingGroup)
		if autoGroupAny, exists := c.Get("auto_group"); exists {
			if autoGroup, ok := autoGroupAny.(string); ok && autoGroup != "" {
				group = autoGroup
			}
		}
		userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
		captureRate := service.GetGroupCaptureRateForSelection(userGroup, group)
		if captureRate <= 0 {
			return
		}

		var requestBody []byte
		if cachedBody, ok := c.Get(common.KeyRequestBody); ok {
			if body, ok := cachedBody.([]byte); ok {
				requestBody = body
			}
		}
		statusCode := c.Writer.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		service.CaptureRelayToLangfuse(service.LangfuseCaptureRequest{
			UserID:              common.GetContextKeyInt(c, constant.ContextKeyUserId),
			TokenID:             common.GetContextKeyInt(c, constant.ContextKeyTokenId),
			TokenName:           c.GetString("token_name"),
			Group:               group,
			Model:               common.GetContextKeyString(c, constant.ContextKeyOriginalModel),
			Method:              c.Request.Method,
			Path:                c.Request.URL.Path,
			RequestID:           c.GetString(common.RequestIdKey),
			CaptureRate:         captureRate,
			StatusCode:          statusCode,
			StartedAt:           startedAt,
			EndedAt:             time.Now(),
			RequestBody:         requestBody,
			ResponseBody:        writer.body.Bytes(),
			RequestContentType:  c.GetHeader("Content-Type"),
			ResponseContentType: c.Writer.Header().Get("Content-Type"),
		})
	}
}
