package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	migrationDisableReason = "已迁移"
	migrationExportVersion = 1
)

type UserMigrationCreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Expression  string `json:"expression"`
	SendEmail   bool   `json:"send_email"`
}

type UserMigrationTargetInput struct {
	UserID int    `json:"user_id"`
	CAHID  string `json:"cah_id"`
	Email  string `json:"email"`
}

type UserMigrationAddTargetsInput struct {
	Expression string                     `json:"expression"`
	Targets    []UserMigrationTargetInput `json:"targets"`
	SendEmail  bool                       `json:"send_email"`
}

type UserMigrationImportInput struct {
	CAHID string          `json:"cah_id"`
	Email string          `json:"email"`
	Data  json.RawMessage `json:"data"`
}

type UserMigrationConfirmInput struct {
	TargetIDs []uint `json:"target_ids"`
}

type UserMigrationSetPasswordInput struct {
	SetupToken  string `json:"setup_token"`
	AccessToken string `json:"access_token"`
	Password    string `json:"password"`
}

type UserMigrationPreviewResult struct {
	Count int              `json:"count"`
	Users []map[string]any `json:"users"`
}

type UserMigrationTargetResult struct {
	Target *model.UserMigrationTarget `json:"target"`
	Link   string                     `json:"link,omitempty"`
	Error  string                     `json:"error,omitempty"`
}

type UserMigrationAddTargetsResult struct {
	Created       int                         `json:"created"`
	Duplicated    int                         `json:"duplicated"`
	EmailSent     int                         `json:"email_sent"`
	EmailFailed   int                         `json:"email_failed"`
	Invalid       int                         `json:"invalid"`
	TargetResults []UserMigrationTargetResult `json:"target_results"`
}

type UserMigrationVerifyResult struct {
	MigrationID string `json:"migration_id"`
	TargetID    uint   `json:"target_id"`
	CAHID       string `json:"cah_id"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	LoginOK     bool   `json:"login_ok"`
	Captured    bool   `json:"captured"`
}

type UserMigrationImportResult struct {
	ImportID string `json:"import_id"`
	UserID   int    `json:"user_id"`
	CAHID    string `json:"cah_id"`
	Email    string `json:"email"`
	Status   string `json:"status"`
	Link     string `json:"link,omitempty"`
}

type userMigrationTargetSecret struct {
	migrationToken string
	accessToken    string
	userToken      string
	link           string
}

type userMigrationEmailJob struct {
	target *model.UserMigrationTarget
	link   string
}

func CreateUserMigration(input UserMigrationCreateInput, operatorID int) (*model.UserMigration, *UserMigrationAddTargetsResult, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, nil, errors.New("迁移名称不能为空")
	}
	expression := strings.TrimSpace(input.Expression)
	if expression != "" {
		if err := ValidateUserMigrationExpression(expression); err != nil {
			return nil, nil, err
		}
	}
	migrateID, err := generateUniqueMigrationID("mig")
	if err != nil {
		return nil, nil, err
	}
	now := common.GetTimestamp()
	migration := &model.UserMigration{
		MigrateID:   migrateID,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		Expression:  expression,
		Status:      model.UserMigrationStatusActive,
		CreatedBy:   operatorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err = model.DB.Create(migration).Error; err != nil {
		return nil, nil, err
	}

	var addResult *UserMigrationAddTargetsResult
	if expression != "" {
		addResult, err = AddUserMigrationTargets(migrateID, UserMigrationAddTargetsInput{
			Expression: expression,
			SendEmail:  input.SendEmail,
		})
		if err != nil {
			return nil, nil, err
		}
	}
	if err = RefreshUserMigrationCounters(migrateID); err != nil {
		return nil, nil, err
	}
	_ = model.DB.First(migration, "migrate_id = ?", migrateID).Error
	return migration, addResult, nil
}

func PreviewUserMigrationExpression(expression string, limit int) (*UserMigrationPreviewResult, error) {
	users, err := MatchUsersForMigrationExpression(expression)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	result := &UserMigrationPreviewResult{Count: len(users), Users: make([]map[string]any, 0, minInt(len(users), limit))}
	for i, user := range users {
		if i >= limit {
			break
		}
		result.Users = append(result.Users, map[string]any{
			"user_id":      user.Id,
			"cah_id":       user.CAHID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"email":        user.Email,
			"group":        user.Group,
			"role":         user.Role,
			"status":       user.Status,
			"providers":    migrationExpressionProviders(user),
		})
	}
	return result, nil
}

func MatchUsersForMigrationExpression(expression string) ([]model.User, error) {
	if err := ValidateUserMigrationExpression(expression); err != nil {
		return nil, err
	}
	statusExplicit := MigrationExpressionMentionsField(expression, "status")
	matched := make([]model.User, 0)

	query := model.DB.Model(&model.User{})
	if !statusExplicit {
		query = query.Where("status = ?", common.UserStatusEnabled)
	}
	var batch []model.User
	err := query.FindInBatches(&batch, 1000, func(tx *gorm.DB, batchNumber int) error {
		for _, user := range batch {
			if user.Role == common.RoleRootUser {
				continue
			}
			ok, err := MatchUserMigrationExpression(expression, user)
			if err != nil {
				return err
			}
			if ok {
				matched = append(matched, user)
			}
		}
		return nil
	}).Error
	if err != nil {
		return nil, err
	}
	return matched, nil
}

func AddUserMigrationTargets(migrateID string, input UserMigrationAddTargetsInput) (*UserMigrationAddTargetsResult, error) {
	migration, err := GetUserMigration(migrateID)
	if err != nil {
		return nil, err
	}
	if migration.Status != model.UserMigrationStatusActive && migration.Status != model.UserMigrationStatusDraft {
		return nil, fmt.Errorf("迁移 %s 当前状态不允许追加用户", migrateID)
	}

	targetInputs := make([]UserMigrationTargetInput, 0)
	if strings.TrimSpace(input.Expression) != "" {
		users, err := MatchUsersForMigrationExpression(input.Expression)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			targetInputs = append(targetInputs, UserMigrationTargetInput{
				UserID: user.Id,
				CAHID:  user.CAHID,
				Email:  user.Email,
			})
		}
	}
	targetInputs = append(targetInputs, input.Targets...)
	if len(targetInputs) == 0 {
		return &UserMigrationAddTargetsResult{}, nil
	}

	result := &UserMigrationAddTargetsResult{TargetResults: make([]UserMigrationTargetResult, 0, len(targetInputs))}
	seen := make(map[int]bool)
	emailJobs := make([]userMigrationEmailJob, 0, len(targetInputs))
	for _, targetInput := range targetInputs {
		user, err := resolveMigrationTargetUser(targetInput)
		if err != nil {
			result.Invalid++
			result.TargetResults = append(result.TargetResults, UserMigrationTargetResult{Error: err.Error()})
			continue
		}
		if user.Role == common.RoleRootUser {
			result.Invalid++
			result.TargetResults = append(result.TargetResults, UserMigrationTargetResult{Error: "不能迁移超级管理员用户"})
			continue
		}
		if seen[user.Id] {
			result.Duplicated++
			continue
		}
		seen[user.Id] = true

		var existing int64
		if err = model.DB.Model(&model.UserMigrationTarget{}).
			Where("migrate_id = ? AND user_id = ?", migrateID, user.Id).
			Count(&existing).Error; err != nil {
			return nil, err
		}
		if existing > 0 {
			result.Duplicated++
			continue
		}

		secret, err := newMigrationTargetSecret(migrateID)
		if err != nil {
			return nil, err
		}
		now := common.GetTimestamp()
		target := &model.UserMigrationTarget{
			MigrateID:          migrateID,
			UserID:             user.Id,
			CAHID:              user.CAHID,
			Email:              strings.TrimSpace(targetInput.Email),
			Status:             model.UserMigrationTargetStatusPending,
			MigrationTokenHash: hashMigrationToken(secret.migrationToken),
			AccessTokenHash:    hashMigrationToken(secret.accessToken),
			UserTokenHash:      hashMigrationToken(secret.userToken),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if target.Email == "" {
			target.Email = user.Email
		}
		if err = model.DB.Create(target).Error; err != nil {
			return nil, err
		}
		result.Created++
		secret.link, err = BuildUserMigrationLink(migrateID, secret.migrationToken, secret.accessToken, secret.userToken)
		if err != nil {
			return nil, err
		}
		if input.SendEmail {
			emailJobs = append(emailJobs, userMigrationEmailJob{target: target, link: secret.link})
		}
		result.TargetResults = append(result.TargetResults, UserMigrationTargetResult{Target: target, Link: secret.link})
	}
	if input.SendEmail && len(emailJobs) > 0 {
		sent, failed, emailErrors := sendUserMigrationTargetEmailBatch(migration, emailJobs)
		result.EmailSent += sent
		result.EmailFailed += failed
		if len(emailErrors) > 0 {
			for i := range result.TargetResults {
				if result.TargetResults[i].Target == nil {
					continue
				}
				if errorMessage, ok := emailErrors[result.TargetResults[i].Target.ID]; ok {
					result.TargetResults[i].Error = errorMessage
				}
			}
		}
	}
	if err = RefreshUserMigrationCounters(migrateID); err != nil {
		return nil, err
	}
	return result, nil
}

func GetUserMigration(migrateID string) (*model.UserMigration, error) {
	migrateID = strings.TrimSpace(migrateID)
	if migrateID == "" {
		return nil, errors.New("迁移 ID 不能为空")
	}
	var migration model.UserMigration
	if err := model.DB.First(&migration, "migrate_id = ?", migrateID).Error; err != nil {
		return nil, err
	}
	return &migration, nil
}

func ListUserMigrations(pageInfo *common.PageInfo) ([]model.UserMigration, int64, error) {
	var migrations []model.UserMigration
	var total int64
	query := model.DB.Model(&model.UserMigration{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&migrations).Error; err != nil {
		return nil, 0, err
	}
	return migrations, total, nil
}

func ListUserMigrationTargets(migrateID string, pageInfo *common.PageInfo) ([]model.UserMigrationTarget, int64, error) {
	var targets []model.UserMigrationTarget
	var total int64
	query := model.DB.Model(&model.UserMigrationTarget{}).Where("migrate_id = ?", migrateID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&targets).Error; err != nil {
		return nil, 0, err
	}
	return targets, total, nil
}

func UpdateUserMigrationStatus(migrateID string, status string) (*model.UserMigration, error) {
	status = strings.TrimSpace(status)
	switch status {
	case model.UserMigrationStatusDraft, model.UserMigrationStatusActive, model.UserMigrationStatusClosed, model.UserMigrationStatusCancelled:
	default:
		return nil, fmt.Errorf("不支持的迁移状态：%s", status)
	}
	migration, err := GetUserMigration(migrateID)
	if err != nil {
		return nil, err
	}
	migration.Status = status
	migration.UpdatedAt = common.GetTimestamp()
	if status == model.UserMigrationStatusClosed || status == model.UserMigrationStatusCancelled {
		migration.ClosedAt = migration.UpdatedAt
	}
	if err = model.DB.Save(migration).Error; err != nil {
		return nil, err
	}
	return migration, nil
}

func VerifyUserMigrationTarget(migrateID, migrationToken, accessToken, userToken string, currentUserID int) (*UserMigrationVerifyResult, error) {
	target, err := findMigrationTargetByTokens(migrateID, migrationToken, accessToken, userToken)
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	updates := map[string]any{"opened_at": now, "updated_at": now}
	if target.Status == model.UserMigrationTargetStatusPending || target.Status == model.UserMigrationTargetStatusEmailSent {
		updates["status"] = model.UserMigrationTargetStatusOpened
		target.Status = model.UserMigrationTargetStatusOpened
	}
	if target.OpenedAt == 0 {
		target.OpenedAt = now
	}
	_ = model.DB.Model(&model.UserMigrationTarget{}).Where("id = ?", target.ID).Updates(updates).Error
	return &UserMigrationVerifyResult{
		MigrationID: migrateID,
		TargetID:    target.ID,
		CAHID:       target.CAHID,
		Email:       target.Email,
		Status:      target.Status,
		LoginOK:     currentUserID != 0 && currentUserID == target.UserID,
		Captured:    target.CapturedAt != 0,
	}, nil
}

func LoginAndCaptureUserMigrationTarget(migrateID, migrationToken, accessToken, userToken, username, password string) (*UserMigrationVerifyResult, error) {
	target, err := findMigrationTargetByTokens(migrateID, migrationToken, accessToken, userToken)
	if err != nil {
		return nil, err
	}
	user := model.User{Username: strings.TrimSpace(username), Password: password}
	if err = user.ValidateAndFill(); err != nil {
		return nil, err
	}
	if user.Id != target.UserID {
		return nil, errors.New("当前登录账户不是此迁移链接指定的账户")
	}
	if err = CaptureUserMigrationData(target.ID); err != nil {
		return nil, err
	}
	return VerifyUserMigrationTarget(migrateID, migrationToken, accessToken, userToken, user.Id)
}

func CaptureUserMigrationData(targetID uint) error {
	var target model.UserMigrationTarget
	if err := model.DB.First(&target, "id = ?", targetID).Error; err != nil {
		return err
	}
	if target.Status == model.UserMigrationTargetStatusMigrated {
		return errors.New("该迁移目标已标记为已迁移")
	}
	user, err := model.GetUserById(target.UserID, true)
	if err != nil {
		return err
	}
	data, err := BuildUserMigrationExportData(*user, target.Email)
	if err != nil {
		return err
	}
	now := common.GetTimestamp()
	return model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserMigrationTarget{}).Where("id = ?", target.ID).Updates(map[string]any{
			"status":      model.UserMigrationTargetStatusCaptured,
			"captured_at": now,
			"data_json":   string(data),
			"updated_at":  now,
		}).Error; err != nil {
			return err
		}
		return refreshUserMigrationCountersTx(tx, target.MigrateID)
	})
}

func BuildUserMigrationExportData(user model.User, migrationEmail string) ([]byte, error) {
	account := sanitizedJSONMap(user, "id", "username", "password", "original_password", "github_id", "discord_id", "oidc_id", "wechat_id", "telegram_id", "linux_do_id", "verification_code", "access_token", "deleted_at")
	account["email"] = strings.TrimSpace(migrationEmail)
	delete(account, "status")
	delete(account, "ban_reason")
	delete(account, "force_password_reset")
	delete(account, "force_email_bind")
	delete(account, "web_session_version")

	records, err := loadUserMigrationRecordSnapshot(user.Id)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"version":     migrationExportVersion,
		"captured_at": common.GetTimestamp(),
		"cah_id":      user.CAHID,
		"email":       strings.TrimSpace(migrationEmail),
		"account":     account,
		"records":     records,
		"excluded": []string{
			"users.id",
			"users.username",
			"users.password",
			"third_party_login_ids",
			"access_tokens",
			"api_key_secrets",
			"sessions",
			"passkeys",
			"two_factor_auth",
			"logs",
			"admin_audit_logs",
		},
	}
	return json.MarshalIndent(payload, "", "  ")
}

func ListUserMigrationExportTargets(migrateID string, limit int) ([]model.UserMigrationTarget, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	query := model.DB.Model(&model.UserMigrationTarget{}).
		Where("status = ? AND captured_at > 0", model.UserMigrationTargetStatusCaptured)
	if strings.TrimSpace(migrateID) != "" {
		query = query.Where("migrate_id = ?", strings.TrimSpace(migrateID))
	}
	var targets []model.UserMigrationTarget
	if err := query.Order("created_at asc, id asc").Limit(limit).Find(&targets).Error; err != nil {
		return nil, err
	}
	return targets, nil
}

func ConfirmUserMigrationTargetsMigrated(input UserMigrationConfirmInput) (int, error) {
	if len(input.TargetIDs) == 0 {
		return 0, errors.New("target_ids 不能为空")
	}
	targetIDs, err := normalizeMigrationTargetIDs(input.TargetIDs)
	if err != nil {
		return 0, err
	}
	now := common.GetTimestamp()
	updated := 0
	updatedUserIDs := make([]int, 0, len(input.TargetIDs))
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var targets []model.UserMigrationTarget
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", targetIDs).
			Find(&targets).Error; err != nil {
			return err
		}
		if len(targets) != len(targetIDs) {
			return errors.New("target_ids 包含不存在的迁移目标")
		}
		for _, target := range targets {
			if target.Status != model.UserMigrationTargetStatusCaptured {
				return fmt.Errorf("迁移目标 %d 当前状态为 %s，不能确认已迁移", target.ID, target.Status)
			}
			if err := tx.Model(&model.UserMigrationTarget{}).Where("id = ?", target.ID).Updates(map[string]any{
				"status":      model.UserMigrationTargetStatusMigrated,
				"migrated_at": now,
				"updated_at":  now,
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.User{}).Where("id = ?", target.UserID).Updates(map[string]any{
				"status":              common.UserStatusDisabled,
				"ban_reason":          migrationDisableReason,
				"web_session_version": gorm.Expr("web_session_version + ?", 1),
			}).Error; err != nil {
				return err
			}
			updated++
			updatedUserIDs = append(updatedUserIDs, target.UserID)
		}
		seenMigrations := map[string]bool{}
		for _, target := range targets {
			if seenMigrations[target.MigrateID] {
				continue
			}
			seenMigrations[target.MigrateID] = true
			if err := refreshUserMigrationCountersTx(tx, target.MigrateID); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		for _, userID := range updatedUserIDs {
			if refreshed, refreshErr := model.GetUserById(userID, false); refreshErr == nil {
				_ = refreshed.UpdateSelected("Status", "BanReason", "WebSessionVersion")
			}
			model.RecordLog(userID, model.LogTypeSystem, "账户已因用户迁移完成而禁用")
		}
	}
	return updated, err
}

func normalizeMigrationTargetIDs(ids []uint) ([]uint, error) {
	seen := make(map[uint]bool, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			return nil, errors.New("target_ids 包含无效 ID")
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	if len(result) == 0 {
		return nil, errors.New("target_ids 不能为空")
	}
	return result, nil
}

func ImportMigratedUser(input UserMigrationImportInput, operatorID int) (*UserMigrationImportResult, error) {
	cahID := model.NormalizeCAHID(input.CAHID)
	email := strings.TrimSpace(input.Email)
	if !model.IsValidCAHID(cahID) {
		return nil, errors.New("CAH 无效")
	}
	if err := validateEmailAddress(email); err != nil {
		return nil, err
	}
	if len(input.Data) == 0 || strings.TrimSpace(string(input.Data)) == "" {
		input.Data = json.RawMessage(`{}`)
	}
	if !json.Valid(input.Data) {
		return nil, errors.New("data 必须是合法 JSON")
	}
	var existing int64
	if err := model.DB.Unscoped().Model(&model.User{}).Where("cah_id = ? OR email = ?", cahID, email).Count(&existing).Error; err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, errors.New("CAH 或邮箱已存在，拒绝重复导入")
	}

	importID, err := generateUniqueMigrationID("imp")
	if err != nil {
		return nil, err
	}
	setupToken, accessToken, err := generateTokenPair()
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	var imported model.UserMigrationImport
	var user model.User
	if err = model.DB.Transaction(func(tx *gorm.DB) error {
		password, err := common.GenerateRandomCharsKey(32)
		if err != nil {
			return err
		}
		hashedPassword, err := common.Password2Hash(password)
		if err != nil {
			return err
		}
		user = buildImportedUserFromData(cahID, email, string(input.Data), hashedPassword)
		if err = tx.Create(&user).Error; err != nil {
			return err
		}
		if err = importUserMigrationRecordsTx(tx, user.Id, user.Username, input.Data); err != nil {
			return err
		}
		imported = model.UserMigrationImport{
			ImportID:        importID,
			CAHID:           cahID,
			Email:           email,
			UserID:          user.Id,
			Status:          model.UserMigrationImportStatusPendingSetup,
			SetupTokenHash:  hashMigrationToken(setupToken),
			AccessTokenHash: hashMigrationToken(accessToken),
			DataJSON:        string(input.Data),
			CreatedBy:       operatorID,
			CreatedAt:       now,
			UpdatedAt:       now,
			ImportedAt:      now,
		}
		return tx.Create(&imported).Error
	}); err != nil {
		return nil, err
	}

	link := BuildUserMigrationImportSetupLink(setupToken, accessToken)
	if err = sendUserMigrationImportEmail(&imported, &user, link); err != nil {
		imported.Status = model.UserMigrationImportStatusEmailFailed
		imported.EmailError = err.Error()
		imported.UpdatedAt = common.GetTimestamp()
		_ = model.DB.Save(&imported).Error
		return &UserMigrationImportResult{
			ImportID: importID,
			UserID:   user.Id,
			CAHID:    cahID,
			Email:    email,
			Status:   imported.Status,
			Link:     link,
		}, nil
	}
	return &UserMigrationImportResult{
		ImportID: importID,
		UserID:   user.Id,
		CAHID:    cahID,
		Email:    email,
		Status:   imported.Status,
		Link:     link,
	}, nil
}

func VerifyUserMigrationImportSetup(setupToken string, accessToken string) (*UserMigrationImportResult, error) {
	importRecord, err := findMigrationImportByTokens(setupToken, accessToken)
	if err != nil {
		return nil, err
	}
	return &UserMigrationImportResult{
		ImportID: importRecord.ImportID,
		UserID:   importRecord.UserID,
		CAHID:    importRecord.CAHID,
		Email:    importRecord.Email,
		Status:   importRecord.Status,
	}, nil
}

func CompleteUserMigrationImportPassword(input UserMigrationSetPasswordInput) (*UserMigrationImportResult, error) {
	password := strings.TrimSpace(input.Password)
	if len(password) < 8 || len(password) > 64 {
		return nil, errors.New("密码长度必须为 8 到 64 位")
	}
	importRecord, err := findMigrationImportByTokens(input.SetupToken, input.AccessToken)
	if err != nil {
		return nil, err
	}
	if importRecord.Status == model.UserMigrationImportStatusActive {
		return nil, errors.New("该导入账户已完成密码设置")
	}
	now := common.GetTimestamp()
	hashedPassword, err := common.Password2Hash(password)
	if err != nil {
		return nil, err
	}
	if err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", importRecord.UserID).Updates(map[string]any{
			"password":             hashedPassword,
			"status":               common.UserStatusEnabled,
			"ban_reason":           "",
			"force_password_reset": false,
			"web_session_version":  gorm.Expr("web_session_version + ?", 1),
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.UserMigrationImport{}).Where("id = ?", importRecord.ID).Updates(map[string]any{
			"status":             model.UserMigrationImportStatusActive,
			"setup_completed_at": now,
			"updated_at":         now,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &UserMigrationImportResult{
		ImportID: importRecord.ImportID,
		UserID:   importRecord.UserID,
		CAHID:    importRecord.CAHID,
		Email:    importRecord.Email,
		Status:   model.UserMigrationImportStatusActive,
	}, nil
}

func RefreshUserMigrationCounters(migrateID string) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		return refreshUserMigrationCountersTx(tx, migrateID)
	})
}

func refreshUserMigrationCountersTx(tx *gorm.DB, migrateID string) error {
	var total, sent, captured, migrated int64
	base := tx.Model(&model.UserMigrationTarget{}).Where("migrate_id = ?", migrateID)
	if err := base.Count(&total).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.UserMigrationTarget{}).Where("migrate_id = ? AND email_sent_at > 0", migrateID).Count(&sent).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.UserMigrationTarget{}).Where("migrate_id = ? AND captured_at > 0", migrateID).Count(&captured).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.UserMigrationTarget{}).Where("migrate_id = ? AND status = ?", migrateID, model.UserMigrationTargetStatusMigrated).Count(&migrated).Error; err != nil {
		return err
	}
	return tx.Model(&model.UserMigration{}).Where("migrate_id = ?", migrateID).Updates(map[string]any{
		"target_count":     int(total),
		"email_sent_count": int(sent),
		"captured_count":   int(captured),
		"migrated_count":   int(migrated),
		"updated_at":       common.GetTimestamp(),
	}).Error
}

func findMigrationTargetByTokens(migrateID, migrationToken, accessToken, userToken string) (*model.UserMigrationTarget, error) {
	if strings.TrimSpace(migrateID) == "" || strings.TrimSpace(migrationToken) == "" || strings.TrimSpace(accessToken) == "" || strings.TrimSpace(userToken) == "" {
		return nil, errors.New("迁移链接参数不完整")
	}
	var migration model.UserMigration
	if err := model.DB.First(&migration, "migrate_id = ?", strings.TrimSpace(migrateID)).Error; err != nil {
		return nil, err
	}
	if migration.Status != model.UserMigrationStatusActive {
		return nil, errors.New("迁移任务未启用")
	}
	query := model.DB.Where("migrate_id = ? AND migration_token_hash = ? AND access_token_hash = ?",
		strings.TrimSpace(migrateID),
		hashMigrationToken(migrationToken),
		hashMigrationToken(accessToken),
	).Where("user_token_hash = ?", hashMigrationToken(userToken))
	var target model.UserMigrationTarget
	if err := query.First(&target).Error; err != nil {
		return nil, errors.New("迁移链接非法或已失效")
	}
	if target.Status == model.UserMigrationTargetStatusMigrated {
		return nil, errors.New("该账户已完成迁移")
	}
	return &target, nil
}

func findMigrationImportByTokens(setupToken string, accessToken string) (*model.UserMigrationImport, error) {
	if strings.TrimSpace(setupToken) == "" || strings.TrimSpace(accessToken) == "" {
		return nil, errors.New("设置链接参数不完整")
	}
	var importRecord model.UserMigrationImport
	if err := model.DB.Where("setup_token_hash = ? AND access_token_hash = ?",
		hashMigrationToken(setupToken),
		hashMigrationToken(accessToken),
	).First(&importRecord).Error; err != nil {
		return nil, errors.New("设置链接非法或已失效")
	}
	return &importRecord, nil
}

func resolveMigrationTargetUser(input UserMigrationTargetInput) (*model.User, error) {
	if input.UserID != 0 {
		return model.GetUserById(input.UserID, true)
	}
	cahID := strings.TrimSpace(input.CAHID)
	if cahID == "" {
		return nil, errors.New("必须提供 user_id 或 cah_id")
	}
	return model.GetUserByCAHID(cahID, true)
}

func markMigrationTargetEmailFailed(target *model.UserMigrationTarget, err error) error {
	now := common.GetTimestamp()
	target.Status = model.UserMigrationTargetStatusEmailFailed
	target.EmailError = err.Error()
	target.UpdatedAt = now
	_ = model.DB.Model(&model.UserMigrationTarget{}).Where("id = ?", target.ID).Updates(map[string]any{
		"status":      target.Status,
		"email_error": target.EmailError,
		"updated_at":  now,
	}).Error
	return err
}

func sendUserMigrationTargetEmailBatch(migration *model.UserMigration, jobs []userMigrationEmailJob) (int, int, map[uint]string) {
	sent := 0
	failed := 0
	errorByTargetID := map[uint]string{}
	entries := make([]common.BatchEmailEntry, 0, len(jobs))
	entryJobs := make([]userMigrationEmailJob, 0, len(jobs))
	targetIDs := make([]string, 0, len(jobs))

	for _, job := range jobs {
		if job.target == nil {
			continue
		}
		targetIDs = append(targetIDs, strconv.Itoa(int(job.target.ID)))
		if err := validateEmailAddress(job.target.Email); err != nil {
			_ = markMigrationTargetEmailFailed(job.target, err)
			failed++
			errorByTargetID[job.target.ID] = err.Error()
			continue
		}
		entries = append(entries, common.BatchEmailEntry{
			Recipient: job.target.Email,
			Subject:   fmt.Sprintf("%s 用户迁移指引", common.SystemName),
			Content:   buildUserMigrationTargetEmailContent(migration, job.target, job.link),
			Context: common.EmailRecipientContext{
				CAHID: job.target.CAHID,
				Email: job.target.Email,
			},
		})
		entryJobs = append(entryJobs, job)
	}
	if len(entries) == 0 {
		return sent, failed, errorByTargetID
	}

	results, err := common.SendBatchEmailsWithIdempotencyKey(
		entries,
		common.GenerateEmailIdempotencyKey("user-migration-invite-batch", migration.MigrateID, strings.Join(targetIDs, ",")),
	)
	if err != nil {
		for _, job := range entryJobs {
			_ = markMigrationTargetEmailFailed(job.target, err)
			failed++
			errorByTargetID[job.target.ID] = err.Error()
		}
		return sent, failed, errorByTargetID
	}

	for i, result := range results {
		if i >= len(entryJobs) {
			break
		}
		job := entryJobs[i]
		if result.Success {
			now := common.GetTimestamp()
			job.target.Status = model.UserMigrationTargetStatusEmailSent
			job.target.EmailSentAt = now
			job.target.EmailError = ""
			job.target.UpdatedAt = now
			_ = model.DB.Model(&model.UserMigrationTarget{}).Where("id = ?", job.target.ID).Updates(map[string]any{
				"status":        job.target.Status,
				"email_sent_at": job.target.EmailSentAt,
				"email_error":   "",
				"updated_at":    job.target.UpdatedAt,
			}).Error
			sent++
			continue
		}
		errorMessage := strings.TrimSpace(result.Error)
		if errorMessage == "" {
			errorMessage = "邮件发送失败"
		}
		_ = markMigrationTargetEmailFailed(job.target, errors.New(errorMessage))
		failed++
		errorByTargetID[job.target.ID] = errorMessage
	}
	return sent, failed, errorByTargetID
}

func buildUserMigrationTargetEmailContent(migration *model.UserMigration, target *model.UserMigrationTarget, link string) string {
	return fmt.Sprintf("您收到此邮件，是因为管理员发起了 **%s** 用户迁移。\n\n请使用下方链接打开迁移流程，并登录 CAH 为 `%s` 的原 PrivHub 账户完成确认：\n\n[%s](%s)\n\n如果链接无法直接点击，请将以下地址复制到浏览器打开：\n\n`%s`\n\n整个流程会在 `/migrate` 路径下完成。完成后，系统会记录 CAH、此邮箱以及可迁移的个人数据，等待接收方确认导入。", migration.Name, target.CAHID, link, link, link)
}

func sendUserMigrationImportEmail(importRecord *model.UserMigrationImport, user *model.User, link string) error {
	subject := fmt.Sprintf("%s 迁移完成，请设置密码", common.SystemName)
	content := fmt.Sprintf("您的账户数据已迁移完成。\n\n请使用下方链接为 CAH `%s` 设置新密码并启用账户：\n\n[%s](%s)\n\n如果链接无法直接点击，请将以下地址复制到浏览器打开：\n\n`%s`\n\n设置密码前，账户不会开放登录。", importRecord.CAHID, link, link, link)
	err := common.SendEmailWithIdempotencyKeyAndContext(
		subject,
		importRecord.Email,
		content,
		common.GenerateEmailIdempotencyKey("user-migration-import-setup", importRecord.ImportID, importRecord.Email),
		common.EmailRecipientContext{
			Username:      user.Username,
			RecipientName: user.DisplayName,
			CAHID:         importRecord.CAHID,
			Email:         importRecord.Email,
		},
	)
	if err != nil {
		return err
	}
	now := common.GetTimestamp()
	importRecord.EmailSentAt = now
	importRecord.EmailError = ""
	importRecord.UpdatedAt = now
	return model.DB.Model(&model.UserMigrationImport{}).Where("id = ?", importRecord.ID).Updates(map[string]any{
		"email_sent_at": now,
		"email_error":   "",
		"updated_at":    now,
	}).Error
}

func BuildUserMigrationLink(migrateID string, migrationToken string, accessToken string, userToken string) (string, error) {
	if strings.TrimSpace(migrateID) == "" || strings.TrimSpace(migrationToken) == "" || strings.TrimSpace(accessToken) == "" || strings.TrimSpace(userToken) == "" {
		return "", errors.New("迁移链接参数不完整")
	}
	fragment := fmt.Sprintf("#~/t/%s/%s", url.PathEscape(migrationToken), url.PathEscape(accessToken))
	fragment += fmt.Sprintf("/~/u/%s", url.PathEscape(userToken))
	return absoluteMigrationURL(fmt.Sprintf("/migrate/%s/%s", url.PathEscape(migrateID), fragment)), nil
}

func BuildUserMigrationImportSetupLink(setupToken string, accessToken string) string {
	fragment := fmt.Sprintf("#~/i/%s/%s", url.PathEscape(setupToken), url.PathEscape(accessToken))
	return absoluteMigrationURL("/migrate/import/" + fragment)
}

func absoluteMigrationURL(path string) string {
	base := strings.TrimRight(strings.TrimSpace(system_setting.ServerAddress), "/")
	if base == "" {
		return path
	}
	return base + path
}

func newMigrationTargetSecret(migrateID string) (*userMigrationTargetSecret, error) {
	migrationToken, accessToken, err := generateTokenPair()
	if err != nil {
		return nil, err
	}
	userToken, err := common.GenerateRandomCharsKey(24)
	if err != nil {
		return nil, err
	}
	return &userMigrationTargetSecret{
		migrationToken: migrationToken,
		accessToken:    accessToken,
		userToken:      strings.ToLower(userToken),
	}, nil
}

func generateTokenPair() (string, string, error) {
	first, err := common.GenerateRandomCharsKey(32)
	if err != nil {
		return "", "", err
	}
	second, err := common.GenerateRandomCharsKey(32)
	if err != nil {
		return "", "", err
	}
	return strings.ToLower(first), strings.ToLower(second), nil
}

func generateUniqueMigrationID(prefix string) (string, error) {
	for i := 0; i < 32; i++ {
		token, err := common.GenerateRandomCharsKey(16)
		if err != nil {
			return "", err
		}
		id := fmt.Sprintf("%s_%s", prefix, strings.ToLower(token))
		var count int64
		if prefix == "imp" {
			if err = model.DB.Model(&model.UserMigrationImport{}).Where("import_id = ?", id).Count(&count).Error; err != nil {
				return "", err
			}
		} else {
			if err = model.DB.Model(&model.UserMigration{}).Where("migrate_id = ?", id).Count(&count).Error; err != nil {
				return "", err
			}
		}
		if count == 0 {
			return id, nil
		}
	}
	return "", errors.New("无法生成唯一迁移 ID")
}

func hashMigrationToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func validateEmailAddress(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("邮箱不能为空")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("邮箱格式不合法")
	}
	return nil
}

func sanitizedJSONMap(value any, removeKeys ...string) map[string]any {
	raw, _ := json.Marshal(value)
	result := map[string]any{}
	_ = json.Unmarshal(raw, &result)
	for _, key := range removeKeys {
		delete(result, key)
	}
	return result
}

func sanitizedJSONSlice[T any](values []T, removeKeys ...string) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		result = append(result, sanitizedJSONMap(value, removeKeys...))
	}
	return result
}

func loadUserMigrationRecordSnapshot(userID int) (map[string]any, error) {
	records := map[string]any{}

	var tokens []model.Token
	if err := model.DB.Unscoped().Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}
	records["tokens"] = sanitizedJSONSlice(tokens, "id", "user_id", "key", "deleted_at")

	var topUps []model.TopUp
	if err := model.DB.Where("user_id = ?", userID).Find(&topUps).Error; err != nil {
		return nil, err
	}
	records["topups"] = sanitizedJSONSlice(topUps, "id", "user_id", "stripe_coupon_id")

	var coupons []model.TopUpCoupon
	if err := model.DB.Where("bound_user_id = ?", userID).Find(&coupons).Error; err != nil {
		return nil, err
	}
	records["topup_coupons"] = sanitizedJSONSlice(coupons, "id", "bound_user_id", "bound_username", "issued_by_admin_id", "reserved_top_up_id", "used_top_up_id", "revoked_by_admin_id")

	var redemptions []model.Redemption
	if err := model.DB.Unscoped().Where("user_id = ? OR used_user_id = ?", userID, userID).Find(&redemptions).Error; err != nil {
		return nil, err
	}
	records["redemptions"] = sanitizedJSONSlice(redemptions, "id", "user_id", "used_user_id", "key", "deleted_at")

	var usageWindows []model.UserUsageWindow
	if err := model.DB.Where("user_id = ?", userID).Find(&usageWindows).Error; err != nil {
		return nil, err
	}
	records["usage_windows"] = sanitizedJSONSlice(usageWindows, "id", "user_id")

	var userMessages []map[string]any
	if err := model.DB.Table("user_messages").
		Select("messages.title, messages.content, messages.status, messages.source, messages.published_at, user_messages.read_at, user_messages.email_sent_at, user_messages.created_at").
		Joins("JOIN messages ON messages.id = user_messages.message_id").
		Where("user_messages.user_id = ?", userID).
		Order("user_messages.id asc").
		Find(&userMessages).Error; err != nil {
		return nil, err
	}
	records["messages"] = userMessages

	return records, nil
}

func buildImportedUserFromData(cahID string, email string, dataJSON string, hashedPassword string) model.User {
	account := migrationAccountMap(dataJSON)
	nowName := strings.ToLower("migrated_" + cahID)
	user := model.User{
		CAHID:              cahID,
		Username:           nowName,
		Password:           hashedPassword,
		DisplayName:        stringFromMap(account, "display_name"),
		Role:               common.RoleCommonUser,
		Status:             common.UserStatusDisabled,
		BanReason:          "等待迁移密码设置",
		Email:              email,
		ForcePasswordReset: true,
		Group:              stringOrDefault(stringFromMap(account, "group"), "default"),
		Quota:              intFromMap(account, "quota"),
		UsedQuota:          intFromMap(account, "used_quota"),
		RequestCount:       intFromMap(account, "request_count"),
		AffCode:            common.GetRandomString(4),
		AffQuota:           intFromMap(account, "aff_quota"),
		AffHistoryQuota:    intFromMap(account, "aff_history_quota"),
		Setting:            stringFromMap(account, "setting"),
		Remark:             stringFromMap(account, "remark"),
		StripeCustomer:     stringFromMap(account, "stripe_customer"),
		SubscriptionData:   stringFromMap(account, "subscription_data"),
	}
	if user.DisplayName == "" {
		user.DisplayName = "Migrated User"
	}
	return user
}

func importUserMigrationRecordsTx(tx *gorm.DB, userID int, username string, data json.RawMessage) error {
	records := migrationRecordsMap(data)
	if rawTopUps, ok := records["topups"].([]any); ok {
		for _, item := range rawTopUps {
			var topUp model.TopUp
			if !decodeMigrationRecord(item, &topUp) {
				continue
			}
			topUp.Id = 0
			topUp.UserId = userID
			topUp.StripeCouponId = ""
			if err := tx.Create(&topUp).Error; err != nil {
				return err
			}
		}
	}
	if rawCoupons, ok := records["topup_coupons"].([]any); ok {
		for _, item := range rawCoupons {
			var coupon model.TopUpCoupon
			if !decodeMigrationRecord(item, &coupon) {
				continue
			}
			coupon.Id = 0
			coupon.BoundUserId = userID
			coupon.BoundUsername = ""
			coupon.IssuedByAdminId = 0
			coupon.ReservedTopUpId = 0
			coupon.UsedTopUpId = 0
			coupon.RevokedByAdminId = 0
			if err := tx.Create(&coupon).Error; err != nil {
				return err
			}
		}
	}
	if rawWindows, ok := records["usage_windows"].([]any); ok {
		for _, item := range rawWindows {
			var window model.UserUsageWindow
			if !decodeMigrationRecord(item, &window) {
				continue
			}
			window.ID = 0
			window.UserID = userID
			if err := tx.Create(&window).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func migrationAccountMap(dataJSON string) map[string]any {
	root := map[string]any{}
	_ = json.Unmarshal([]byte(dataJSON), &root)
	if account, ok := root["account"].(map[string]any); ok {
		return account
	}
	return root
}

func migrationRecordsMap(data json.RawMessage) map[string]any {
	root := map[string]any{}
	_ = json.Unmarshal(data, &root)
	if records, ok := root["records"].(map[string]any); ok {
		return records
	}
	return map[string]any{}
}

func decodeMigrationRecord(input any, output any) bool {
	raw, err := json.Marshal(input)
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, output) == nil
}

func stringFromMap(values map[string]any, key string) string {
	if value, ok := values[key]; ok {
		switch v := value.(type) {
		case string:
			return strings.TrimSpace(v)
		case fmt.Stringer:
			return strings.TrimSpace(v.String())
		}
	}
	return ""
}

func stringOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func intFromMap(values map[string]any, key string) int {
	if value, ok := values[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case json.Number:
			i, _ := v.Int64()
			return int(i)
		case string:
			i, _ := strconv.Atoi(strings.TrimSpace(v))
			return i
		}
	}
	return 0
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
