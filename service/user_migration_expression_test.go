package service

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMatchUserMigrationExpression(t *testing.T) {
	user := model.User{
		CAHID:     "ABCDE1",
		Username:  "team-alpha",
		Email:     "alpha@example.com",
		Group:     "paid",
		Role:      common.RoleCommonUser,
		Status:    common.UserStatusEnabled,
		GitHubId:  "gh_1",
		Password:  "hashed",
		DiscordId: "",
	}

	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{
			name:       "matches email and status",
			expression: `status == "enabled" and email matches ".*@example\.com$"`,
			want:       true,
		},
		{
			name:       "provider in set",
			expression: `provider in {"github", "oidc"} and group == "paid"`,
			want:       true,
		},
		{
			name:       "not contains",
			expression: `not email contains "+test" and role in {"common", "support"}`,
			want:       true,
		},
		{
			name:       "negative group",
			expression: `group == "free" or status == "disabled"`,
			want:       false,
		},
		{
			name:       "not equal true",
			expression: `email != "other@example.com"`,
			want:       true,
		},
		{
			name:       "not equal false",
			expression: `email != "alpha@example.com"`,
			want:       false,
		},
		{
			name:       "ne true",
			expression: `email ne "other@example.com"`,
			want:       true,
		},
		{
			name:       "provider not equal false for existing provider",
			expression: `provider != "github"`,
			want:       false,
		},
		{
			name:       "provider not equal true for missing provider",
			expression: `provider != "discord"`,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchUserMigrationExpression(tt.expression, user)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("MatchUserMigrationExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUserMigrationExpressionRejectsUnknownField(t *testing.T) {
	err := ValidateUserMigrationExpression(`token.key == "secret"`)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestMigrationExpressionRegexPreservesEscapedDot(t *testing.T) {
	user := model.User{
		Email:  "a@exampleXcom",
		Role:   common.RoleCommonUser,
		Status: common.UserStatusEnabled,
	}
	got, err := MatchUserMigrationExpression(`email matches ".*@example\.com$"`, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatal("escaped dot should not match exampleXcom")
	}

	user.Email = "a@example.com"
	got, err = MatchUserMigrationExpression(`email matches ".*@example\.com$"`, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("escaped dot should match example.com")
	}
}

func TestMigrationExpressionMentionsField(t *testing.T) {
	if !MigrationExpressionMentionsField(`cah_id == "ABCDE1" and status == "disabled"`, "status") {
		t.Fatal("expected status field to be detected")
	}
	if MigrationExpressionMentionsField(`email == "a@example.com"`, "status") {
		t.Fatal("did not expect status field to be detected")
	}
}

func TestMatchUsersForMigrationExpressionDefaultsToEnabledUsers(t *testing.T) {
	originalDB := model.DB
	db, err := gorm.Open(sqlite.Open("file:user-migration-expression-users?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	model.DB = db
	t.Cleanup(func() {
		model.DB = originalDB
	})

	users := []model.User{
		{CAHID: "ABCDE4", Username: "enabled", Email: "enabled@example.com", Password: "hashed-password", Status: common.UserStatusEnabled, Role: common.RoleCommonUser, AffCode: "AFF1"},
		{CAHID: "ABCDF2", Username: "disabled", Email: "disabled@example.com", Password: "hashed-password", Status: common.UserStatusDisabled, Role: common.RoleCommonUser, AffCode: "AFF2"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	matched, err := MatchUsersForMigrationExpression(`email contains "example.com"`)
	if err != nil {
		t.Fatalf("match enabled expression: %v", err)
	}
	if len(matched) != 1 || matched[0].Username != "enabled" {
		t.Fatalf("expected only enabled user by default, got %+v", matched)
	}

	matched, err = MatchUsersForMigrationExpression(`status == "disabled"`)
	if err != nil {
		t.Fatalf("match disabled expression: %v", err)
	}
	if len(matched) != 1 || matched[0].Username != "disabled" {
		t.Fatalf("expected disabled user when status is explicit, got %+v", matched)
	}
}

func TestUserMigrationDataJSONUsesLongText(t *testing.T) {
	targetField, ok := reflect.TypeOf(model.UserMigrationTarget{}).FieldByName("DataJSON")
	if !ok {
		t.Fatal("UserMigrationTarget.DataJSON field missing")
	}
	if !strings.Contains(strings.ToLower(string(targetField.Tag)), "longtext") {
		t.Fatalf("expected target data_json to use longtext, got tag %q", targetField.Tag)
	}

	importField, ok := reflect.TypeOf(model.UserMigrationImport{}).FieldByName("DataJSON")
	if !ok {
		t.Fatal("UserMigrationImport.DataJSON field missing")
	}
	if !strings.Contains(strings.ToLower(string(importField.Tag)), "longtext") {
		t.Fatalf("expected import data_json to use longtext, got tag %q", importField.Tag)
	}
}

func TestBuildUserMigrationLinkRequiresUserToken(t *testing.T) {
	if _, err := BuildUserMigrationLink("mig_test", "migration-token", "access-token", ""); err == nil {
		t.Fatal("expected empty user token to fail")
	}
	link, err := BuildUserMigrationLink("mig_test", "migration-token", "access-token", "user-token")
	if err != nil {
		t.Fatalf("expected complete link to be generated, got %v", err)
	}
	if !strings.Contains(link, "/~/u/user-token") {
		t.Fatalf("expected generated link to include user token, got %q", link)
	}
}

func TestConfirmUserMigrationTargetsMigratedRequiresAllTargetsCaptured(t *testing.T) {
	originalDB := model.DB
	originalLogDB := model.LOG_DB
	db, err := gorm.Open(sqlite.Open("file:user-migration-confirm?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserMigration{}, &model.UserMigrationTarget{}, &model.Log{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	model.DB = db
	model.LOG_DB = db
	t.Cleanup(func() {
		model.DB = originalDB
		model.LOG_DB = originalLogDB
	})

	user1 := model.User{Username: "captured-user", Email: "captured@example.com", Password: "hashed-password", Status: common.UserStatusEnabled, Role: common.RoleCommonUser, AffCode: "AFFC1"}
	user2 := model.User{Username: "pending-user", Email: "pending@example.com", Password: "hashed-password", Status: common.UserStatusEnabled, Role: common.RoleCommonUser, AffCode: "AFFC2"}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("create user1: %v", err)
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("create user2: %v", err)
	}
	migration := model.UserMigration{MigrateID: "mig_confirm", Name: "confirm", Status: model.UserMigrationStatusActive}
	if err := db.Create(&migration).Error; err != nil {
		t.Fatalf("create migration: %v", err)
	}
	capturedTarget := model.UserMigrationTarget{
		MigrateID: migration.MigrateID,
		UserID:    user1.Id,
		Status:    model.UserMigrationTargetStatusCaptured,
		DataJSON:  "{}",
	}
	pendingTarget := model.UserMigrationTarget{
		MigrateID: migration.MigrateID,
		UserID:    user2.Id,
		Status:    model.UserMigrationTargetStatusPending,
		DataJSON:  "{}",
	}
	if err := db.Create(&capturedTarget).Error; err != nil {
		t.Fatalf("create captured target: %v", err)
	}
	if err := db.Create(&pendingTarget).Error; err != nil {
		t.Fatalf("create pending target: %v", err)
	}

	updated, err := ConfirmUserMigrationTargetsMigrated(UserMigrationConfirmInput{TargetIDs: []uint{capturedTarget.ID, 99999}})
	if err == nil {
		t.Fatal("expected partial invalid target list to fail")
	}
	if updated != 0 {
		t.Fatalf("expected updated=0 on failure, got %d", updated)
	}
	var unchanged model.UserMigrationTarget
	if err := db.First(&unchanged, capturedTarget.ID).Error; err != nil {
		t.Fatalf("load captured target: %v", err)
	}
	if unchanged.Status != model.UserMigrationTargetStatusCaptured {
		t.Fatalf("expected captured target to remain captured, got %s", unchanged.Status)
	}

	updated, err = ConfirmUserMigrationTargetsMigrated(UserMigrationConfirmInput{TargetIDs: []uint{pendingTarget.ID}})
	if err == nil {
		t.Fatal("expected non-captured target to fail")
	}
	if updated != 0 {
		t.Fatalf("expected updated=0 for non-captured target, got %d", updated)
	}

	updated, err = ConfirmUserMigrationTargetsMigrated(UserMigrationConfirmInput{TargetIDs: []uint{capturedTarget.ID, capturedTarget.ID}})
	if err != nil {
		t.Fatalf("expected captured target to confirm, got %v", err)
	}
	if updated != 1 {
		t.Fatalf("expected one unique target to be updated, got %d", updated)
	}
	var migratedTarget model.UserMigrationTarget
	if err := db.First(&migratedTarget, capturedTarget.ID).Error; err != nil {
		t.Fatalf("load migrated target: %v", err)
	}
	if migratedTarget.Status != model.UserMigrationTargetStatusMigrated {
		t.Fatalf("expected target migrated, got %s", migratedTarget.Status)
	}
	var migratedUser model.User
	if err := db.First(&migratedUser, user1.Id).Error; err != nil {
		t.Fatalf("load migrated user: %v", err)
	}
	if migratedUser.Status != common.UserStatusDisabled || migratedUser.BanReason != migrationDisableReason {
		t.Fatalf("expected user disabled as migrated, got status=%d reason=%q", migratedUser.Status, migratedUser.BanReason)
	}
	var logCount int64
	if err := db.Model(&model.Log{}).Where("user_id = ?", user1.Id).Count(&logCount).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	if logCount != 1 {
		t.Fatalf("expected one post-commit system log, got %d", logCount)
	}
}

func TestImportMigratedUserMailFailureReturnsEmailFailedResult(t *testing.T) {
	originalDB := model.DB
	db, err := gorm.Open(sqlite.Open("file:user-migration-import-mail-fail?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserMigrationImport{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	model.DB = db
	t.Cleanup(func() {
		model.DB = originalDB
	})

	originalToken := common.PostmarkServerToken
	originalSender := common.PostmarkSenderEmail
	common.PostmarkServerToken = ""
	common.PostmarkSenderEmail = ""
	t.Cleanup(func() {
		common.PostmarkServerToken = originalToken
		common.PostmarkSenderEmail = originalSender
	})

	result, err := ImportMigratedUser(UserMigrationImportInput{
		CAHID: "ABCDE4",
		Email: "migrated@example.com",
		Data:  json.RawMessage(`{"account":{"group":"default","display_name":"Migrated"}}`),
	}, 7)
	if err != nil {
		t.Fatalf("expected mail failure to be returned as email_failed result, got error: %v", err)
	}
	if result == nil {
		t.Fatal("expected import result")
	}
	if result.Status != model.UserMigrationImportStatusEmailFailed {
		t.Fatalf("expected status=%s, got %s", model.UserMigrationImportStatusEmailFailed, result.Status)
	}
	if result.Link == "" {
		t.Fatal("expected setup link to be returned for recovery")
	}

	var importRecord model.UserMigrationImport
	if err := db.First(&importRecord, "import_id = ?", result.ImportID).Error; err != nil {
		t.Fatalf("load import record: %v", err)
	}
	if strings.TrimSpace(importRecord.EmailError) == "" {
		t.Fatal("expected email_error to be saved")
	}

	_, err = ImportMigratedUser(UserMigrationImportInput{
		CAHID: "ABCDE4",
		Email: "migrated@example.com",
		Data:  json.RawMessage(`{}`),
	}, 7)
	if err == nil {
		t.Fatal("expected duplicate import to be rejected")
	}
}

func TestFindMigrationTargetRequiresUserToken(t *testing.T) {
	originalDB := model.DB
	db, err := gorm.Open(sqlite.Open("file:user-migration-token-required?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.UserMigration{}, &model.UserMigrationTarget{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	model.DB = db
	t.Cleanup(func() {
		model.DB = originalDB
	})

	migration := model.UserMigration{
		MigrateID: "mig_test",
		Name:      "test",
		Status:    model.UserMigrationStatusActive,
	}
	if err := db.Create(&migration).Error; err != nil {
		t.Fatalf("create migration: %v", err)
	}
	target := model.UserMigrationTarget{
		MigrateID:          migration.MigrateID,
		UserID:             1,
		Status:             model.UserMigrationTargetStatusPending,
		MigrationTokenHash: hashMigrationToken("migration-token"),
		AccessTokenHash:    hashMigrationToken("access-token"),
		UserTokenHash:      hashMigrationToken("user-token"),
	}
	if err := db.Create(&target).Error; err != nil {
		t.Fatalf("create target: %v", err)
	}

	if _, err := findMigrationTargetByTokens(migration.MigrateID, "migration-token", "access-token", ""); err == nil {
		t.Fatal("expected empty user token to fail")
	}
	if _, err := findMigrationTargetByTokens(migration.MigrateID, "migration-token", "access-token", "user-token"); err != nil {
		t.Fatalf("expected complete token set to pass, got %v", err)
	}
}
