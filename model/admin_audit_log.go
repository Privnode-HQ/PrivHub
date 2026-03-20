package model

import (
	"context"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
)

type AdminAuditLog struct {
	Id               int    `json:"id" gorm:"index:idx_admin_audit_created_at_id,priority:1"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index:idx_admin_audit_created_at_id,priority:2;index:idx_admin_audit_created_at_operator"`
	OperatorId       int    `json:"operator_id" gorm:"column:operator_id;index;index:idx_admin_audit_created_at_operator,priority:2"`
	OperatorUsername string `json:"operator_username" gorm:"column:operator_username;type:varchar(64);index"`
	OperatorRole     int    `json:"operator_role" gorm:"column:operator_role;index"`
	Method           string `json:"method" gorm:"type:varchar(8);index"`
	Path             string `json:"path" gorm:"type:varchar(255);index"`
	Route            string `json:"route" gorm:"type:varchar(255);index"`
	Resource         string `json:"resource" gorm:"type:varchar(64);index"`
	Action           string `json:"action" gorm:"type:varchar(64);index"`
	TargetType       string `json:"target_type" gorm:"type:varchar(64);index"`
	TargetId         int    `json:"target_id" gorm:"index"`
	TargetName       string `json:"target_name" gorm:"type:varchar(255);default:''"`
	StatusCode       int    `json:"status_code" gorm:"index"`
	Success          bool   `json:"success" gorm:"index"`
	Ip               string `json:"ip" gorm:"type:varchar(64);default:''"`
	Content          string `json:"content" gorm:"type:text"`
	Details          string `json:"details" gorm:"type:text"`
}

type AdminAuditMeta struct {
	Resource   string                 `json:"resource,omitempty"`
	Action     string                 `json:"action,omitempty"`
	TargetType string                 `json:"target_type,omitempty"`
	TargetId   int                    `json:"target_id,omitempty"`
	TargetName string                 `json:"target_name,omitempty"`
	Content    string                 `json:"content,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func GetAdminAuditMeta(c *gin.Context) AdminAuditMeta {
	meta, ok := common.GetContextKeyType[AdminAuditMeta](c, constant.ContextKeyAdminAuditMeta)
	if !ok {
		return AdminAuditMeta{}
	}
	return meta
}

func SetAdminAuditMeta(c *gin.Context, meta AdminAuditMeta) {
	current := GetAdminAuditMeta(c)
	if meta.Resource != "" {
		current.Resource = meta.Resource
	}
	if meta.Action != "" {
		current.Action = meta.Action
	}
	if meta.TargetType != "" {
		current.TargetType = meta.TargetType
	}
	if meta.TargetId != 0 {
		current.TargetId = meta.TargetId
	}
	if meta.TargetName != "" {
		current.TargetName = meta.TargetName
	}
	if meta.Content != "" {
		current.Content = meta.Content
	}
	if len(meta.Details) > 0 {
		if current.Details == nil {
			current.Details = make(map[string]interface{}, len(meta.Details))
		}
		for key, value := range meta.Details {
			current.Details[key] = value
		}
	}
	common.SetContextKey(c, constant.ContextKeyAdminAuditMeta, current)
}

func SkipAdminAudit(c *gin.Context) {
	common.SetContextKey(c, constant.ContextKeyAdminAuditSkip, true)
}

func RecordAdminAuditLog(log *AdminAuditLog) error {
	if log == nil {
		return nil
	}
	if log.CreatedAt == 0 {
		log.CreatedAt = common.GetTimestamp()
	}
	return LOG_DB.Create(log).Error
}

func GetAllAdminAuditLogs(startTimestamp int64, endTimestamp int64, username string, startIdx int, num int) (logs []*AdminAuditLog, total int64, err error) {
	tx := LOG_DB.Model(&AdminAuditLog{})
	if username != "" {
		tx = tx.Where("operator_username = ?", username)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if err = tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func SearchAdminAuditLogs(keyword string) (logs []*AdminAuditLog, err error) {
	pattern := "%" + strings.TrimSpace(keyword) + "%"
	err = LOG_DB.Where(
		"operator_username LIKE ? OR content LIKE ? OR target_name LIKE ? OR path LIKE ? OR route LIKE ?",
		pattern, pattern, pattern, pattern, pattern,
	).Order("id desc").Limit(common.MaxRecentItems).Find(&logs).Error
	return logs, err
}

func DeleteOldAdminAuditLog(ctx context.Context, targetTimestamp int64, limit int) (int64, error) {
	var total int64 = 0

	for {
		if ctx.Err() != nil {
			return total, ctx.Err()
		}

		result := LOG_DB.Where("created_at < ?", targetTimestamp).Limit(limit).Delete(&AdminAuditLog{})
		if result.Error != nil {
			return total, result.Error
		}

		total += result.RowsAffected
		if result.RowsAffected < int64(limit) {
			break
		}
	}

	return total, nil
}

func ConvertAdminAuditLogsToLogs(auditLogs []*AdminAuditLog) []*Log {
	logs := make([]*Log, 0, len(auditLogs))
	for _, auditLog := range auditLogs {
		if auditLog == nil {
			continue
		}
		logs = append(logs, auditLog.ToLog())
	}
	return logs
}

func (log *AdminAuditLog) ToLog() *Log {
	if log == nil {
		return nil
	}
	return &Log{
		Id:        log.Id,
		UserId:    log.OperatorId,
		CreatedAt: log.CreatedAt,
		Type:      LogTypeManage,
		Content:   log.Content,
		Username:  log.OperatorUsername,
		Ip:        log.Ip,
		Other:     log.toLegacyOther(),
	}
}

func (log *AdminAuditLog) toLegacyOther() string {
	other := map[string]interface{}{
		"request_path": log.Path,
		"audit_info": map[string]interface{}{
			"operator_id":   log.OperatorId,
			"operator_role": log.OperatorRole,
			"route":         log.Route,
			"resource":      log.Resource,
			"action":        log.Action,
			"target_type":   log.TargetType,
			"target_id":     log.TargetId,
			"target_name":   log.TargetName,
			"status_code":   log.StatusCode,
			"success":       log.Success,
		},
	}
	if log.Details != "" {
		if detailMap, err := common.StrToMap(log.Details); err == nil && detailMap != nil {
			other["audit_details"] = detailMap
		} else {
			other["audit_details_raw"] = log.Details
		}
	}
	return common.MapToJsonStr(other)
}
