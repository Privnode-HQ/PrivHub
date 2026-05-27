package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetUserMigrationExpressionDocs(c *gin.Context) {
	common.ApiSuccess(c, service.UserMigrationExpressionDocs())
}

func PreviewUserMigrationExpression(c *gin.Context) {
	var req struct {
		Expression string `json:"expression"`
		Limit      int    `json:"limit"`
	}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	result, err := service.PreviewUserMigrationExpression(req.Expression, req.Limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func CreateUserMigration(c *gin.Context) {
	var req service.UserMigrationCreateInput
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	migration, addResult, err := service.CreateUserMigration(req, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"migration":   migration,
		"add_targets": addResult,
	})
}

func ListUserMigrations(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	migrations, total, err := service.ListUserMigrations(pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(migrations)
	common.ApiSuccess(c, pageInfo)
}

func GetUserMigration(c *gin.Context) {
	migration, err := service.GetUserMigration(c.Param("migrate_id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, migration)
}

func UpdateUserMigrationStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status"`
	}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	migration, err := service.UpdateUserMigrationStatus(c.Param("migrate_id"), req.Status)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, migration)
}

func AddUserMigrationTargets(c *gin.Context) {
	var req service.UserMigrationAddTargetsInput
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	result, err := service.AddUserMigrationTargets(c.Param("migrate_id"), req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func ListUserMigrationTargets(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	targets, total, err := service.ListUserMigrationTargets(c.Param("migrate_id"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(targets)
	common.ApiSuccess(c, pageInfo)
}

func VerifyUserMigrationLink(c *gin.Context) {
	var req struct {
		MigrationToken string `json:"migration_token"`
		AccessToken    string `json:"access_token"`
		UserToken      string `json:"user_token"`
	}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	currentUserID := getOptionalValidatedSessionUserID(c)
	result, err := service.VerifyUserMigrationTarget(c.Param("migrate_id"), req.MigrationToken, req.AccessToken, req.UserToken, currentUserID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func LoginUserMigrationLink(c *gin.Context) {
	var req struct {
		MigrationToken string `json:"migration_token"`
		AccessToken    string `json:"access_token"`
		UserToken      string `json:"user_token"`
		Username       string `json:"username"`
		Password       string `json:"password"`
	}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	result, err := service.LoginAndCaptureUserMigrationTarget(c.Param("migrate_id"), req.MigrationToken, req.AccessToken, req.UserToken, req.Username, req.Password)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user := model.User{Username: strings.TrimSpace(req.Username), Password: req.Password}
	if err = user.ValidateAndFill(); err == nil {
		session := sessions.Default(c)
		service.ApplyDefaultWebSessionOptions(session)
		service.SetAuthenticatedUserSession(session, &user)
		if saveErr := session.Save(); saveErr != nil {
			common.ApiError(c, saveErr)
			return
		}
		common.ApiSuccess(c, gin.H{
			"migration": result,
			"user":      buildAuthenticatedUserResponseData(&user),
		})
		return
	}
	common.ApiSuccess(c, result)
}

func CaptureCurrentUserMigrationLink(c *gin.Context) {
	var req struct {
		MigrationToken string `json:"migration_token"`
		AccessToken    string `json:"access_token"`
		UserToken      string `json:"user_token"`
	}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	currentUserID := getOptionalValidatedSessionUserID(c)
	verifyResult, err := service.VerifyUserMigrationTarget(c.Param("migrate_id"), req.MigrationToken, req.AccessToken, req.UserToken, currentUserID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !verifyResult.LoginOK {
		common.ApiErrorMsg(c, "请先登录迁移链接指定的账户")
		return
	}
	if err = service.CaptureUserMigrationData(verifyResult.TargetID); err != nil {
		common.ApiError(c, err)
		return
	}
	verifyResult.Captured = true
	verifyResult.Status = model.UserMigrationTargetStatusCaptured
	common.ApiSuccess(c, verifyResult)
}

func ListUserMigrationExports(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	targets, err := service.ListUserMigrationExportTargets(c.Query("migrate_id"), limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]map[string]any, 0, len(targets))
	for _, target := range targets {
		items = append(items, map[string]any{
			"target_id":        target.ID,
			"migrate_id":       target.MigrateID,
			"cah_id":           target.CAHID,
			"email":            target.Email,
			"captured_at":      target.CapturedAt,
			"last_exported_at": target.LastExportedAt,
			"data_size":        len(target.DataJSON),
			"download_url":     "/migrate/api/admin/exports/" + strconv.Itoa(int(target.ID)),
		})
	}
	common.ApiSuccess(c, items)
}

func DownloadUserMigrationExport(c *gin.Context) {
	targetID, _ := strconv.Atoi(c.Param("target_id"))
	var target model.UserMigrationTarget
	if err := model.DB.First(&target, "id = ?", targetID).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	if strings.TrimSpace(target.DataJSON) == "" {
		common.ApiErrorMsg(c, "该迁移目标尚未生成导出数据")
		return
	}
	fileName := target.MigrateID + "_" + target.CAHID + ".json"
	c.Header("Content-Disposition", `attachment; filename="`+fileName+`"`)
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(target.DataJSON))
}

func ConfirmUserMigrationExports(c *gin.Context) {
	var req service.UserMigrationConfirmInput
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	updated, err := service.ConfirmUserMigrationTargetsMigrated(req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"updated": updated})
}

func ImportMigratedUser(c *gin.Context) {
	var req service.UserMigrationImportInput
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	result, err := service.ImportMigratedUser(req, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func VerifyUserMigrationImportSetup(c *gin.Context) {
	var req struct {
		SetupToken  string `json:"setup_token"`
		AccessToken string `json:"access_token"`
	}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	result, err := service.VerifyUserMigrationImportSetup(req.SetupToken, req.AccessToken)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func CompleteUserMigrationImportPassword(c *gin.Context) {
	var req service.UserMigrationSetPasswordInput
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	result, err := service.CompleteUserMigrationImportPassword(req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}
