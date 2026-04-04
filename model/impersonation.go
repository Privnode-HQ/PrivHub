package model

import (
	"errors"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	ImpersonationSourceAdminRequest = "admin_request"
	ImpersonationSourceSelfService  = "self_service"

	ImpersonationModeStandard   = "standard"
	ImpersonationModeBreakGlass = "break_glass"

	ImpersonationStatePending   = "pending"
	ImpersonationStateApproved  = "approved"
	ImpersonationStateRejected  = "rejected"
	ImpersonationStateCancelled = "cancelled"
	ImpersonationStateExpired   = "expired"
	ImpersonationStateActive    = "active"
	ImpersonationStateCompleted = "completed"
)

type ImpersonationGrant struct {
	Id                uint       `json:"id"`
	ApprovalToken     string     `json:"approval_token,omitempty" gorm:"size:96;uniqueIndex"`
	Source            string     `json:"source" gorm:"size:24;index;not null;default:'admin_request'"`
	Mode              string     `json:"mode" gorm:"size:24;index;not null;default:'standard'"`
	State             string     `json:"state" gorm:"size:24;index;not null;default:'pending'"`
	OperatorId        int        `json:"operator_id" gorm:"index"`
	OperatorUsername  string     `json:"operator_username" gorm:"size:64;index"`
	OperatorCAHID     string     `json:"operator_cah_id" gorm:"column:operator_cah_id;size:6;index"`
	TargetUserId      int        `json:"target_user_id" gorm:"index"`
	TargetUsername    string     `json:"target_username" gorm:"size:64;index"`
	TargetDisplayName string     `json:"target_display_name" gorm:"size:128"`
	TargetCAHID       string     `json:"target_cah_id" gorm:"column:target_cah_id;size:6;index"`
	RequestedReadOnly bool       `json:"requested_read_only" gorm:"not null;default:false"`
	ActiveReadOnly    bool       `json:"active_read_only" gorm:"not null;default:false"`
	ApprovedByUserId  int        `json:"approved_by_user_id" gorm:"index"`
	ApprovedByMethod  string     `json:"approved_by_method" gorm:"size:24"`
	RequestedAt       time.Time  `json:"requested_at" gorm:"index"`
	ApprovedAt        *time.Time `json:"approved_at,omitempty" gorm:"index"`
	GrantedExpiresAt  *time.Time `json:"granted_expires_at,omitempty" gorm:"index"`
	ActivatedAt       *time.Time `json:"activated_at,omitempty" gorm:"index"`
	SessionExpiresAt  *time.Time `json:"session_expires_at,omitempty" gorm:"index"`
	EndedAt           *time.Time `json:"ended_at,omitempty" gorm:"index"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type ImpersonationActionLog struct {
	Id               uint      `json:"id"`
	GrantId          uint      `json:"grant_id" gorm:"index"`
	CreatedAt        time.Time `json:"created_at" gorm:"index"`
	OperatorId       int       `json:"operator_id" gorm:"index"`
	OperatorUsername string    `json:"operator_username" gorm:"size:64;index"`
	TargetUserId     int       `json:"target_user_id" gorm:"index"`
	Method           string    `json:"method" gorm:"size:8;index"`
	Path             string    `json:"path" gorm:"size:255;index"`
	Route            string    `json:"route" gorm:"size:255"`
	StatusCode       int       `json:"status_code" gorm:"index"`
	Success          bool      `json:"success" gorm:"index"`
}

func (grant *ImpersonationGrant) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(grant.ApprovalToken) != "" {
		return nil
	}
	token, err := common.GenerateURLSafeToken(24)
	if err != nil {
		return err
	}
	grant.ApprovalToken = token
	return nil
}

type BreakGlassIncident struct {
	Grant   *ImpersonationGrant
	Actions []*ImpersonationActionLog
}

func (grant *ImpersonationGrant) IsBreakGlass() bool {
	return grant != nil && grant.Mode == ImpersonationModeBreakGlass
}

func (grant *ImpersonationGrant) IsPending() bool {
	return grant != nil && grant.State == ImpersonationStatePending
}

func (grant *ImpersonationGrant) IsApproved() bool {
	return grant != nil && grant.State == ImpersonationStateApproved
}

func (grant *ImpersonationGrant) IsActive() bool {
	return grant != nil && grant.State == ImpersonationStateActive
}

func (grant *ImpersonationGrant) AllowsRequestedAccess(requestReadOnly bool) bool {
	if grant == nil {
		return false
	}
	if !grant.RequestedReadOnly {
		return true
	}
	return requestReadOnly
}

func (grant *ImpersonationGrant) HasGrantWindowExpired(now time.Time) bool {
	if grant == nil || grant.GrantedExpiresAt == nil {
		return false
	}
	return now.After(*grant.GrantedExpiresAt)
}

func (grant *ImpersonationGrant) HasSessionExpired(now time.Time) bool {
	if grant == nil || grant.SessionExpiresAt == nil {
		return false
	}
	return now.After(*grant.SessionExpiresAt)
}

func CreateImpersonationGrant(grant *ImpersonationGrant) error {
	if grant == nil {
		return nil
	}
	if grant.RequestedAt.IsZero() {
		grant.RequestedAt = time.Now()
	}
	return DB.Create(grant).Error
}

func SaveImpersonationGrant(grant *ImpersonationGrant) error {
	if grant == nil {
		return nil
	}
	return DB.Save(grant).Error
}

func GetImpersonationGrantByID(id uint) (*ImpersonationGrant, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var grant ImpersonationGrant
	if err := DB.First(&grant, id).Error; err != nil {
		return nil, err
	}
	return &grant, nil
}

func GetImpersonationGrantByToken(token string) (*ImpersonationGrant, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var grant ImpersonationGrant
	if err := DB.Where("approval_token = ?", token).First(&grant).Error; err != nil {
		return nil, err
	}
	return &grant, nil
}

func FindLatestPendingImpersonationGrant(operatorID int, targetUserID int, requestReadOnly bool) (*ImpersonationGrant, error) {
	var grants []*ImpersonationGrant
	if err := DB.Where("operator_id = ? AND target_user_id = ? AND state = ?", operatorID, targetUserID, ImpersonationStatePending).
		Order("id desc").
		Limit(10).
		Find(&grants).Error; err != nil {
		return nil, err
	}
	for _, grant := range grants {
		if grant != nil && grant.AllowsRequestedAccess(requestReadOnly) {
			return grant, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func FindLatestActivatableImpersonationGrant(operatorID int, targetUserID int, requestReadOnly bool) (*ImpersonationGrant, error) {
	var grants []*ImpersonationGrant
	if err := DB.Where("target_user_id = ? AND state = ?", targetUserID, ImpersonationStateApproved).
		Where("operator_id = ? OR operator_id = 0", operatorID).
		Order("id desc").
		Limit(20).
		Find(&grants).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	for _, grant := range grants {
		if grant == nil {
			continue
		}
		if !grant.AllowsRequestedAccess(requestReadOnly) {
			continue
		}
		if grant.HasGrantWindowExpired(now) {
			continue
		}
		return grant, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func FindLatestOpenSelfServiceGrant(targetUserID int) (*ImpersonationGrant, error) {
	var grant ImpersonationGrant
	err := DB.Where("target_user_id = ? AND source = ? AND state = ?", targetUserID, ImpersonationSourceSelfService, ImpersonationStateApproved).
		Order("id desc").
		First(&grant).Error
	if err != nil {
		return nil, err
	}
	if grant.HasGrantWindowExpired(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}
	return &grant, nil
}

func CreateImpersonationActionLog(log *ImpersonationActionLog) error {
	if log == nil {
		return nil
	}
	return DB.Create(log).Error
}

func ListBreakGlassIncidentsByUserID(userID int, limit int) ([]*ImpersonationGrant, error) {
	if limit <= 0 {
		limit = common.MaxRecentItems
	}
	var grants []*ImpersonationGrant
	err := DB.Where("target_user_id = ? AND mode = ?", userID, ImpersonationModeBreakGlass).
		Order("COALESCE(activated_at, requested_at) desc").
		Limit(limit).
		Find(&grants).Error
	return grants, err
}

func ListImpersonationActionLogsByGrantIDs(grantIDs []uint) (map[uint][]*ImpersonationActionLog, error) {
	result := make(map[uint][]*ImpersonationActionLog)
	if len(grantIDs) == 0 {
		return result, nil
	}
	var logs []*ImpersonationActionLog
	if err := DB.Where("grant_id IN ?", grantIDs).Order("created_at asc, id asc").Find(&logs).Error; err != nil {
		return nil, err
	}
	for _, log := range logs {
		if log == nil {
			continue
		}
		result[log.GrantId] = append(result[log.GrantId], log)
	}
	return result, nil
}

func GetLatestBreakGlassAlertGrant(userID int) (*ImpersonationGrant, error) {
	var grant ImpersonationGrant
	recentThreshold := time.Now().Add(-30 * 24 * time.Hour)
	err := DB.Where("target_user_id = ? AND mode = ?", userID, ImpersonationModeBreakGlass).
		Where("ended_at IS NULL OR ended_at >= ?", recentThreshold).
		Order("CASE WHEN ended_at IS NULL THEN 0 ELSE 1 END asc, COALESCE(ended_at, activated_at, requested_at) desc").
		First(&grant).Error
	if err != nil {
		return nil, err
	}
	return &grant, nil
}

func ExpireImpersonationGrantIfNeeded(grant *ImpersonationGrant, now time.Time) bool {
	if grant == nil {
		return false
	}
	if grant.State == ImpersonationStateExpired || grant.State == ImpersonationStateCompleted || grant.State == ImpersonationStateRejected || grant.State == ImpersonationStateCancelled {
		return false
	}
	changed := false
	if grant.State == ImpersonationStateApproved && grant.HasGrantWindowExpired(now) {
		grant.State = ImpersonationStateExpired
		changed = true
	}
	if grant.State == ImpersonationStateActive && grant.HasSessionExpired(now) {
		grant.State = ImpersonationStateCompleted
		if grant.EndedAt == nil {
			grant.EndedAt = &now
		}
		changed = true
	}
	return changed
}

func LoadBreakGlassIncidentsByUserID(userID int, limit int) ([]*BreakGlassIncident, error) {
	grants, err := ListBreakGlassIncidentsByUserID(userID, limit)
	if err != nil {
		return nil, err
	}
	grantIDs := make([]uint, 0, len(grants))
	for _, grant := range grants {
		if grant != nil {
			grantIDs = append(grantIDs, grant.Id)
		}
	}
	logMap, err := ListImpersonationActionLogsByGrantIDs(grantIDs)
	if err != nil {
		return nil, err
	}
	incidents := make([]*BreakGlassIncident, 0, len(grants))
	for _, grant := range grants {
		if grant == nil {
			continue
		}
		incidents = append(incidents, &BreakGlassIncident{
			Grant:   grant,
			Actions: logMap[grant.Id],
		})
	}
	return incidents, nil
}

func CompleteImpersonationGrant(grantID uint, endedAt time.Time) error {
	if grantID == 0 {
		return nil
	}
	updates := map[string]interface{}{
		"state":    ImpersonationStateCompleted,
		"ended_at": endedAt,
	}
	return DB.Model(&ImpersonationGrant{}).Where("id = ? AND ended_at IS NULL", grantID).Updates(updates).Error
}

func CancelPendingImpersonationGrant(grantID uint) error {
	if grantID == 0 {
		return nil
	}
	return DB.Model(&ImpersonationGrant{}).Where("id = ? AND state = ?", grantID, ImpersonationStatePending).
		Update("state", ImpersonationStateCancelled).Error
}

func RejectPendingImpersonationGrant(grantID uint, approvedByUserID int, method string, decidedAt time.Time) error {
	if grantID == 0 {
		return nil
	}
	updates := map[string]interface{}{
		"state":               ImpersonationStateRejected,
		"approved_by_user_id": approvedByUserID,
		"approved_by_method":  strings.TrimSpace(method),
		"approved_at":         decidedAt,
	}
	return DB.Model(&ImpersonationGrant{}).Where("id = ? AND state = ?", grantID, ImpersonationStatePending).Updates(updates).Error
}

func ApprovePendingImpersonationGrant(grantID uint, approvedByUserID int, method string, decidedAt time.Time, grantExpiresAt time.Time) error {
	if grantID == 0 {
		return nil
	}
	updates := map[string]interface{}{
		"state":               ImpersonationStateApproved,
		"approved_by_user_id": approvedByUserID,
		"approved_by_method":  strings.TrimSpace(method),
		"approved_at":         decidedAt,
		"granted_expires_at":  grantExpiresAt,
	}
	return DB.Model(&ImpersonationGrant{}).Where("id = ? AND state = ?", grantID, ImpersonationStatePending).Updates(updates).Error
}

func ActivateApprovedImpersonationGrant(grantID uint, operatorID int, operatorUsername string, operatorCAHID string, readOnly bool, activatedAt time.Time, sessionExpiresAt *time.Time) error {
	if grantID == 0 {
		return errors.New("grant id is required")
	}
	updates := map[string]interface{}{
		"state":              ImpersonationStateActive,
		"operator_id":        operatorID,
		"operator_username":  strings.TrimSpace(operatorUsername),
		"operator_cah_id":    strings.TrimSpace(operatorCAHID),
		"active_read_only":   readOnly,
		"activated_at":       activatedAt,
		"session_expires_at": sessionExpiresAt,
	}
	return DB.Model(&ImpersonationGrant{}).Where("id = ? AND state = ?", grantID, ImpersonationStateApproved).Updates(updates).Error
}
