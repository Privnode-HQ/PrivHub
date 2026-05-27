package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetMigrateRouter(router *gin.Engine) {
	migrateAPI := router.Group("/migrate/api")
	migrateAPI.Use(gzip.Gzip(gzip.DefaultCompression))
	migrateAPI.Use(middleware.GlobalAPIRateLimit())
	{
		migrationFlow := migrateAPI.Group("/migrations/:migrate_id")
		migrationFlow.POST("/verify", controller.VerifyUserMigrationLink)
		migrationFlow.POST("/login", middleware.CriticalRateLimit(), controller.LoginUserMigrationLink)
		migrationFlow.POST("/capture", middleware.CriticalRateLimit(), controller.CaptureCurrentUserMigrationLink)

		migrateAPI.POST("/imports/setup/verify", controller.VerifyUserMigrationImportSetup)
		migrateAPI.POST("/imports/setup/password", middleware.CriticalRateLimit(), controller.CompleteUserMigrationImportPassword)

		adminRoute := migrateAPI.Group("/admin")
		adminRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
		{
			adminRoute.GET("/expression-docs", controller.GetUserMigrationExpressionDocs)
			adminRoute.POST("/preview", controller.PreviewUserMigrationExpression)
			adminRoute.GET("/migrations", controller.ListUserMigrations)
			adminRoute.POST("/migrations", controller.CreateUserMigration)
			adminRoute.GET("/migrations/:migrate_id", controller.GetUserMigration)
			adminRoute.PUT("/migrations/:migrate_id/status", controller.UpdateUserMigrationStatus)
			adminRoute.GET("/migrations/:migrate_id/targets", controller.ListUserMigrationTargets)
			adminRoute.POST("/migrations/:migrate_id/targets", controller.AddUserMigrationTargets)
			adminRoute.GET("/exports", controller.ListUserMigrationExports)
			adminRoute.GET("/exports/:target_id", controller.DownloadUserMigrationExport)
			adminRoute.POST("/exports/confirm", controller.ConfirmUserMigrationExports)
			adminRoute.POST("/imports/users", controller.ImportMigratedUser)
		}
	}
}
