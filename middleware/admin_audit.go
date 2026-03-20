package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

const (
	adminAuditBodyLimit     = 4096
	adminAuditResponseLimit = 4096
	adminAuditValueLimit    = 512
)

type adminAuditResponseWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *adminAuditResponseWriter) Write(data []byte) (int, error) {
	w.writeBody(data)
	return w.ResponseWriter.Write(data)
}

func (w *adminAuditResponseWriter) WriteString(s string) (int, error) {
	w.writeBody([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *adminAuditResponseWriter) writeBody(data []byte) {
	remaining := adminAuditResponseLimit - w.body.Len()
	if remaining <= 0 || len(data) == 0 {
		return
	}
	if len(data) > remaining {
		w.body.Write(data[:remaining])
		return
	}
	w.body.Write(data)
}

func AdminAudit() func(c *gin.Context) {
	return func(c *gin.Context) {
		requestBody, _ := common.GetRequestBody(c)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		recorder := &adminAuditResponseWriter{ResponseWriter: c.Writer}
		c.Writer = recorder

		startTime := time.Now()
		c.Next()

		if common.GetContextKeyBool(c, constant.ContextKeyAdminAuditSkip) {
			return
		}

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		resource := deriveAuditResource(route)
		sanitizedBody, rawBodyMap := sanitizeAuditBody(requestBody, c.ContentType(), resource)
		query := sanitizeQuery(c.Request.URL.Query(), resource)

		meta := model.GetAdminAuditMeta(c)
		if meta.Resource == "" {
			meta.Resource = resource
		}
		if meta.Action == "" {
			meta.Action = deriveAuditAction(c.Request.Method, route)
		}
		if meta.TargetType == "" && meta.TargetId == 0 && meta.TargetName == "" {
			meta.TargetType, meta.TargetId, meta.TargetName = inferAuditTarget(meta.Resource, c, rawBodyMap)
		}

		success, responseMessage := parseAuditResponse(c.Writer.Status(), recorder.body.Bytes())
		details := mergeAuditDetails(meta.Details)
		if len(query) > 0 {
			details["query"] = query
		}
		if sanitizedBody != nil {
			details["body"] = sanitizedBody
		}
		if responseMessage != "" {
			details["response_message"] = responseMessage
		}
		details["elapsed_ms"] = time.Since(startTime).Milliseconds()

		content := strings.TrimSpace(meta.Content)
		if content == "" {
			content = buildAuditContent(c.GetInt("id"), c.GetString("username"), c.Request.Method, c.Request.URL.Path, meta.TargetType, meta.TargetId, meta.TargetName, success, responseMessage)
		}

		auditLog := &model.AdminAuditLog{
			CreatedAt:        common.GetTimestamp(),
			OperatorId:       c.GetInt("id"),
			OperatorUsername: c.GetString("username"),
			OperatorRole:     c.GetInt("role"),
			Method:           c.Request.Method,
			Path:             c.Request.URL.Path,
			Route:            route,
			Resource:         meta.Resource,
			Action:           meta.Action,
			TargetType:       meta.TargetType,
			TargetId:         meta.TargetId,
			TargetName:       truncateAuditString(meta.TargetName, adminAuditValueLimit),
			StatusCode:       c.Writer.Status(),
			Success:          success,
			Ip:               truncateAuditString(c.ClientIP(), 64),
			Content:          truncateAuditString(content, 1024),
			Details:          common.MapToJsonStr(details),
		}
		if err := model.RecordAdminAuditLog(auditLog); err != nil {
			common.SysLog("failed to record admin audit log: " + err.Error())
		}
	}
}

func mergeAuditDetails(details map[string]interface{}) map[string]interface{} {
	if len(details) == 0 {
		return make(map[string]interface{})
	}
	merged := make(map[string]interface{}, len(details))
	for key, value := range details {
		merged[key] = value
	}
	return merged
}

func sanitizeQuery(values url.Values, resource string) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(values))
	for key, items := range values {
		if len(items) == 0 {
			continue
		}
		if len(items) == 1 {
			result[key] = sanitizeAuditValue(items[0], key, resource)
			continue
		}
		sanitizedItems := make([]interface{}, 0, len(items))
		for _, item := range items {
			sanitizedItems = append(sanitizedItems, sanitizeAuditValue(item, key, resource))
		}
		result[key] = sanitizedItems
	}
	return result
}

func sanitizeAuditBody(raw []byte, contentType string, resource string) (interface{}, map[string]interface{}) {
	trimmedBody := bytes.TrimSpace(raw)
	if len(trimmedBody) == 0 {
		return nil, nil
	}

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		var payload interface{}
		if err := json.Unmarshal(trimmedBody, &payload); err == nil {
			sanitized := sanitizeAuditValue(payload, "", resource)
			bodyMap, _ := payload.(map[string]interface{})
			return applyAuditResourceRedaction(sanitized, resource, bodyMap), bodyMap
		}
		return truncateAuditString(string(trimmedBody), adminAuditBodyLimit), nil
	case strings.Contains(contentType, gin.MIMEPOSTForm):
		values, err := url.ParseQuery(string(trimmedBody))
		if err != nil {
			return truncateAuditString(string(trimmedBody), adminAuditBodyLimit), nil
		}
		rawMap := valuesToMap(values)
		return sanitizeQuery(values, resource), rawMap
	case strings.Contains(contentType, gin.MIMEMultipartPOSTForm):
		return "[multipart form omitted]", nil
	default:
		return truncateAuditString(string(trimmedBody), adminAuditBodyLimit), nil
	}
}

func valuesToMap(values url.Values) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(values))
	for key, items := range values {
		if len(items) == 1 {
			result[key] = items[0]
			continue
		}
		multi := make([]interface{}, 0, len(items))
		for _, item := range items {
			multi = append(multi, item)
		}
		result[key] = multi
	}
	return result
}

func sanitizeAuditValue(value interface{}, fieldName string, resource string) interface{} {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typedValue))
		for key, item := range typedValue {
			result[key] = sanitizeAuditValue(item, key, resource)
		}
		return result
	case []interface{}:
		items := make([]interface{}, 0, len(typedValue))
		for _, item := range typedValue {
			items = append(items, sanitizeAuditValue(item, fieldName, resource))
		}
		return items
	case string:
		if shouldMaskAuditField(fieldName, resource) {
			return maskAuditString(typedValue)
		}
		return truncateAuditString(typedValue, adminAuditValueLimit)
	default:
		if shouldMaskAuditField(fieldName, resource) {
			return "[REDACTED]"
		}
		return typedValue
	}
}

func shouldMaskAuditField(fieldName string, resource string) bool {
	name := strings.ToLower(strings.TrimSpace(fieldName))
	switch {
	case name == "":
		return false
	case strings.Contains(name, "password"),
		strings.Contains(name, "secret"),
		strings.Contains(name, "token"),
		strings.Contains(name, "authorization"),
		strings.Contains(name, "private_key"),
		strings.Contains(name, "access_token"),
		strings.Contains(name, "client_secret"),
		strings.Contains(name, "credential"):
		return true
	case resource == "channel" && (name == "key" || name == "header_override"):
		return true
	default:
		return false
	}
}

func applyAuditResourceRedaction(value interface{}, resource string, rawBodyMap map[string]interface{}) interface{} {
	bodyMap, ok := value.(map[string]interface{})
	if !ok {
		return value
	}
	if resource == "channel" {
		if _, exists := bodyMap["key"]; exists {
			bodyMap["key"] = "[REDACTED]"
		}
		if _, exists := bodyMap["header_override"]; exists {
			bodyMap["header_override"] = "[REDACTED]"
		}
	}
	if resource == "option" && rawBodyMap != nil {
		optionKey, _ := rawBodyMap["key"].(string)
		if isSensitiveOptionKey(optionKey) {
			if _, exists := bodyMap["value"]; exists {
				bodyMap["value"] = "[REDACTED]"
			}
		}
	}
	return bodyMap
}

func isSensitiveOptionKey(key string) bool {
	key = strings.TrimSpace(key)
	return strings.HasSuffix(key, "Token") || strings.HasSuffix(key, "Secret") || strings.HasSuffix(key, "Key")
}

func parseAuditResponse(statusCode int, responseBody []byte) (bool, string) {
	success := statusCode < 400
	responseBody = bytes.TrimSpace(responseBody)
	if len(responseBody) == 0 {
		return success, ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(responseBody, &payload); err == nil {
		if value, ok := payload["success"].(bool); ok {
			success = value
		}
		if message, ok := payload["message"]; ok {
			return success, truncateAuditString(strings.TrimSpace(common.Interface2String(message)), 255)
		}
		return success, ""
	}

	text := strings.TrimSpace(string(responseBody))
	if text == "success" {
		return true, ""
	}
	return success, truncateAuditString(text, 255)
}

func deriveAuditResource(route string) string {
	segments := splitAuditPath(route)
	if len(segments) >= 2 && segments[0] == "api" {
		return strings.ReplaceAll(segments[1], "-", "_")
	}
	if len(segments) > 0 {
		return strings.ReplaceAll(segments[0], "-", "_")
	}
	return ""
}

func deriveAuditAction(method string, route string) string {
	segments := splitAuditPath(route)
	if len(segments) == 0 {
		return auditActionByMethod(method)
	}
	if segments[0] == "api" {
		segments = segments[1:]
	}
	if len(segments) <= 1 {
		return auditActionByMethod(method)
	}

	staticSegments := make([]string, 0, len(segments)-1)
	for _, segment := range segments[1:] {
		if segment == "" || strings.HasPrefix(segment, ":") {
			continue
		}
		staticSegments = append(staticSegments, strings.ReplaceAll(segment, "-", "_"))
	}
	if len(staticSegments) == 0 {
		return auditActionByMethod(method)
	}
	return strings.Join(staticSegments, "_")
}

func auditActionByMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func splitAuditPath(path string) []string {
	trimmedPath := strings.TrimSpace(path)
	trimmedPath = strings.Trim(trimmedPath, "/")
	if trimmedPath == "" {
		return nil
	}
	return strings.Split(trimmedPath, "/")
}

func inferAuditTarget(resource string, c *gin.Context, body map[string]interface{}) (string, int, string) {
	targetType := resource
	if targetType == "" {
		targetType = "resource"
	}

	if pathID := strings.TrimSpace(c.Param("id")); pathID != "" {
		if id, err := strconv.Atoi(pathID); err == nil {
			return targetType, id, inferAuditTargetName(resource, body)
		}
	}
	if body == nil {
		return targetType, 0, ""
	}

	if resource == "option" {
		if key, ok := body["key"].(string); ok {
			return "option", 0, truncateAuditString(strings.TrimSpace(key), 255)
		}
	}

	if id, ok := readAuditInt(body["id"]); ok {
		return targetType, id, inferAuditTargetName(resource, body)
	}
	if id, ok := readAuditInt(body["user_id"]); ok {
		return "user", id, inferAuditTargetName("user", body)
	}
	if id, ok := readAuditInt(body["bound_user_id"]); ok {
		return "user", id, inferAuditTargetName("user", body)
	}
	if id, ok := readAuditInt(body["channel_id"]); ok {
		return "channel", id, inferAuditTargetName("channel", body)
	}
	if id, ok := readAuditInt(body["coupon_id"]); ok {
		return "topup_coupon", id, inferAuditTargetName("topup_coupon", body)
	}
	return targetType, 0, inferAuditTargetName(resource, body)
}

func inferAuditTargetName(resource string, body map[string]interface{}) string {
	if body == nil {
		return ""
	}
	if resource == "option" {
		if key, ok := body["key"].(string); ok {
			return truncateAuditString(strings.TrimSpace(key), 255)
		}
	}
	if username, ok := body["username"].(string); ok {
		return truncateAuditString(strings.TrimSpace(username), 255)
	}
	if name, ok := body["name"].(string); ok {
		return truncateAuditString(strings.TrimSpace(name), 255)
	}
	if productID, ok := body["product_id"].(string); ok {
		return truncateAuditString(strings.TrimSpace(productID), 255)
	}
	return ""
}

func readAuditInt(value interface{}) (int, bool) {
	switch typedValue := value.(type) {
	case float64:
		return int(typedValue), true
	case float32:
		return int(typedValue), true
	case int:
		return typedValue, true
	case int64:
		return int(typedValue), true
	case string:
		id, err := strconv.Atoi(strings.TrimSpace(typedValue))
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}

func buildAuditContent(operatorID int, username string, method string, path string, targetType string, targetID int, targetName string, success bool, responseMessage string) string {
	builder := strings.Builder{}
	builder.WriteString("管理员(ID:")
	builder.WriteString(strconv.Itoa(operatorID))
	builder.WriteString(")")
	if username != "" {
		builder.WriteString(" ")
		builder.WriteString(username)
	}
	builder.WriteString(" 执行 ")
	builder.WriteString(strings.ToUpper(method))
	builder.WriteString(" ")
	builder.WriteString(path)

	targetDesc := formatAuditTarget(targetType, targetID, targetName)
	if targetDesc != "" {
		builder.WriteString("，目标 ")
		builder.WriteString(targetDesc)
	}
	if success {
		builder.WriteString("，结果成功")
	} else if responseMessage != "" {
		builder.WriteString("，结果失败：")
		builder.WriteString(truncateAuditString(responseMessage, 120))
	} else {
		builder.WriteString("，结果失败")
	}
	return builder.String()
}

func formatAuditTarget(targetType string, targetID int, targetName string) string {
	parts := make([]string, 0, 3)
	if targetType != "" {
		parts = append(parts, targetType)
	}
	if targetID != 0 {
		parts = append(parts, "#"+strconv.Itoa(targetID))
	}
	if targetName != "" {
		parts = append(parts, targetName)
	}
	return strings.Join(parts, " ")
}

func maskAuditString(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "[REDACTED]"
}

func truncateAuditString(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
