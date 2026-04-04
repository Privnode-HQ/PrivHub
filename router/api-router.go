package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	{
		apiRouter.GET("/setup", controller.GetSetup)
		apiRouter.POST("/setup", controller.PostSetup)
		apiRouter.GET("/status", controller.GetStatus)
		apiRouter.GET("/uptime/status", controller.GetUptimeKumaStatus)
		apiRouter.GET("/models", middleware.UserAuth(), controller.DashboardListModels)
		apiRouter.GET("/status/test", middleware.AdminAuth(), middleware.AdminAudit(), controller.TestStatus)
		apiRouter.GET("/notice", controller.GetNotice)
		apiRouter.GET("/user-agreement", controller.GetUserAgreement)
		apiRouter.GET("/privacy-policy", controller.GetPrivacyPolicy)
		apiRouter.GET("/about", controller.GetAbout)
		//apiRouter.GET("/midjourney", controller.GetMidjourney)
		apiRouter.GET("/home_page_content", controller.GetHomePageContent)
		apiRouter.GET("/pricing", middleware.TryUserAuth(), controller.GetPricing)
		apiRouter.GET("/verification", middleware.EmailVerificationRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)
		apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), controller.GitHubOAuth)
		apiRouter.GET("/oauth/discord", middleware.CriticalRateLimit(), controller.DiscordOAuth)
		apiRouter.GET("/oauth/oidc", middleware.CriticalRateLimit(), controller.OidcAuth)
		apiRouter.GET("/oauth/linuxdo", middleware.CriticalRateLimit(), controller.LinuxdoOAuth)
		apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), controller.GenerateOAuthCode)
		apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), controller.WeChatAuth)
		apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), controller.WeChatBind)
		apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), controller.EmailBind)
		apiRouter.GET("/oauth/telegram/login", middleware.CriticalRateLimit(), controller.TelegramLogin)
		apiRouter.GET("/oauth/telegram/bind", middleware.CriticalRateLimit(), controller.TelegramBind)
		apiRouter.GET("/ratio_config", middleware.CriticalRateLimit(), controller.GetRatioConfig)

		apiRouter.POST("/stripe/webhook", controller.StripeWebhook)
		apiRouter.POST("/creem/webhook", controller.CreemWebhook)

		// Universal secure verification routes
		apiRouter.POST("/verify", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.UniversalVerify)
		apiRouter.GET("/verify/status", middleware.UserAuth(), controller.GetVerificationStatus)

		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
			userRoute.POST("/login", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Login)
			userRoute.POST("/login/2fa", middleware.CriticalRateLimit(), controller.Verify2FALogin)
			userRoute.POST("/passkey/login/begin", middleware.CriticalRateLimit(), controller.PasskeyLoginBegin)
			userRoute.POST("/passkey/login/finish", middleware.CriticalRateLimit(), controller.PasskeyLoginFinish)
			//userRoute.POST("/tokenlog", middleware.CriticalRateLimit(), controller.TokenLog)
			userRoute.GET("/logout", controller.Logout)
			userRoute.GET("/epay/notify", controller.EpayNotify)
			userRoute.GET("/groups", controller.GetUserGroups)

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth())
			{
				selfRoute.GET("/self/groups", controller.GetUserGroups)
				selfRoute.GET("/self", controller.GetSelf)
				selfRoute.GET("/models", controller.GetUserModels)
				selfRoute.POST("/self/back_to_payg", controller.BackToPayAsYouGo)
				selfRoute.PUT("/self", controller.UpdateSelf)
				selfRoute.DELETE("/self", controller.DeleteSelf)
				selfRoute.GET("/token", controller.GenerateAccessToken)
				selfRoute.POST("/impersonation/stop", controller.StopUserImpersonation)
				selfRoute.GET("/impersonation/request/:token", controller.GetImpersonationRequest)
				selfRoute.POST("/impersonation/request/:token/approve", controller.ApproveImpersonationRequest)
				selfRoute.POST("/impersonation/request/:token/reject", controller.RejectImpersonationRequest)
				selfRoute.GET("/impersonation/history", controller.GetImpersonationHistory)
				selfRoute.POST("/impersonation/open_access", controller.OpenSelfServiceSupportAccess)
				selfRoute.DELETE("/impersonation/open_access", controller.CloseSelfServiceSupportAccess)
				selfRoute.GET("/passkey", controller.PasskeyStatus)
				selfRoute.POST("/passkey/register/begin", controller.PasskeyRegisterBegin)
				selfRoute.POST("/passkey/register/finish", controller.PasskeyRegisterFinish)
				selfRoute.POST("/passkey/verify/begin", controller.PasskeyVerifyBegin)
				selfRoute.POST("/passkey/verify/finish", controller.PasskeyVerifyFinish)
				selfRoute.DELETE("/passkey", controller.PasskeyDelete)
				selfRoute.GET("/aff", controller.GetAffCode)
				selfRoute.GET("/topup/info", controller.GetTopUpInfo)
				selfRoute.POST("/topup/quote", controller.RequestTopUpQuote)
				selfRoute.GET("/topup/self", controller.GetUserTopUps)
				selfRoute.POST("/topup", middleware.TurnstileCheck(), controller.TopUp)
				selfRoute.POST("/pay", middleware.CriticalRateLimit(), controller.RequestEpay)
				selfRoute.POST("/amount", controller.RequestAmount)
				selfRoute.POST("/stripe/pay", middleware.CriticalRateLimit(), controller.RequestStripePay)
				selfRoute.POST("/stripe/amount", controller.RequestStripeAmount)
				selfRoute.POST("/creem/pay", middleware.CriticalRateLimit(), controller.RequestCreemPay)
				selfRoute.POST("/aff_transfer", controller.TransferAffQuota)
				selfRoute.PUT("/setting", controller.UpdateUserSetting)

				// 2FA routes
				selfRoute.GET("/2fa/status", controller.Get2FAStatus)
				selfRoute.POST("/2fa/setup", controller.Setup2FA)
				selfRoute.POST("/2fa/enable", controller.Enable2FA)
				selfRoute.POST("/2fa/disable", controller.Disable2FA)
				selfRoute.POST("/2fa/backup_codes", controller.RegenerateBackupCodes)
			}

			supportRoute := userRoute.Group("/")
			supportRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
			{
				supportRoute.GET("/", controller.GetAllUsers)
				supportRoute.GET("/topup", controller.GetAllTopUps)
				supportRoute.GET("/search", controller.SearchUsers)
				supportRoute.GET("/:id", controller.GetUser)
			}

			adminRoute := userRoute.Group("/")
			adminRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
			{
				adminRoute.POST("/topup/complete", controller.AdminCompleteTopUp)
				adminRoute.POST("/", controller.CreateUser)
				adminRoute.POST("/:id/impersonation", controller.StartUserImpersonation)
				adminRoute.POST("/manage", controller.ManageUser)
				adminRoute.POST("/logout_all", controller.LogoutAllUsers)
				adminRoute.PUT("/", controller.UpdateUser)
				adminRoute.DELETE("/:id", controller.DeleteUser)
				adminRoute.DELETE("/:id/reset_passkey", controller.AdminResetPasskey)
				adminRoute.GET("/2fa/stats", controller.Admin2FAStats)
				adminRoute.DELETE("/:id/2fa", controller.AdminDisable2FA)
			}
		}
		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth(), middleware.AdminAudit())
		{
			optionRoute.GET("/", controller.GetOptions)
			optionRoute.PUT("/", controller.UpdateOption)
			optionRoute.POST("/rest_model_ratio", controller.ResetModelRatio)
			optionRoute.POST("/migrate_console_setting", controller.MigrateConsoleSetting) // 用于迁移检测的旧键，下个版本会删除
		}
		messageRoute := apiRouter.Group("/message")
		{
			selfMessageRoute := messageRoute.Group("/self")
			selfMessageRoute.Use(middleware.UserAuth())
			{
				selfMessageRoute.GET("/", controller.GetMyMessages)
				selfMessageRoute.GET("/unread", controller.GetMyUnreadMessageCount)
				selfMessageRoute.POST("/:id/read", controller.ReadMyMessage)
				selfMessageRoute.POST("/read/batch", controller.BatchReadMyMessages)
			}

			supportMessageRoute := messageRoute.Group("/")
			supportMessageRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
			{
				supportMessageRoute.GET("/", controller.GetAdminMessages)
				supportMessageRoute.GET("/template", controller.GetMessageTemplate)
			}

			adminMessageRoute := messageRoute.Group("/")
			adminMessageRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
			{
				adminMessageRoute.PUT("/template", controller.UpdateMessageTemplate)
				adminMessageRoute.POST("/", controller.CreateMessage)
				adminMessageRoute.PUT("/:id", controller.UpdateMessage)
				adminMessageRoute.POST("/:id/publish", controller.PublishMessage)
				adminMessageRoute.POST("/:id/copy", controller.CopyMessage)
				adminMessageRoute.POST("/:id/retry", controller.RetryMessageDelivery)
				adminMessageRoute.DELETE("/:id", controller.DeleteMessage)
			}
		}
		ratioSyncRoute := apiRouter.Group("/ratio_sync")
		ratioSyncRoute.Use(middleware.RootAuth(), middleware.AdminAudit())
		{
			ratioSyncRoute.GET("/channels", controller.GetSyncableChannels)
			ratioSyncRoute.POST("/fetch", controller.FetchUpstreamRatios)
		}
		channelRoute := apiRouter.Group("/channel")
		channelRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
		{
			channelRoute.GET("/", controller.GetAllChannels)
			channelRoute.GET("/search", controller.SearchChannels)
			channelRoute.GET("/models", controller.ChannelListModels)
			channelRoute.GET("/models_enabled", controller.EnabledListModels)
			channelRoute.GET("/:id", controller.GetChannel)
			channelRoute.POST("/:id/key", middleware.RootAuth(), middleware.CriticalRateLimit(), middleware.DisableCache(), middleware.SecureVerificationRequired(), controller.GetChannelKey)
			channelRoute.GET("/test", controller.TestAllChannels)
			channelRoute.GET("/test/:id", controller.TestChannel)
			channelRoute.GET("/update_balance", controller.UpdateAllChannelsBalance)
			channelRoute.GET("/update_balance/:id", controller.UpdateChannelBalance)
			channelRoute.POST("/", controller.AddChannel)
			channelRoute.PUT("/", controller.UpdateChannel)
			channelRoute.DELETE("/disabled", controller.DeleteDisabledChannel)
			channelRoute.POST("/tag/disabled", controller.DisableTagChannels)
			channelRoute.POST("/tag/enabled", controller.EnableTagChannels)
			channelRoute.PUT("/tag", controller.EditTagChannels)
			channelRoute.DELETE("/:id", controller.DeleteChannel)
			channelRoute.POST("/batch", controller.DeleteChannelBatch)
			channelRoute.POST("/fix", controller.FixChannelsAbilities)
			channelRoute.GET("/fetch_models/:id", controller.FetchUpstreamModels)
			channelRoute.POST("/fetch_models", controller.FetchModels)
			channelRoute.POST("/batch/tag", controller.BatchSetChannelTag)
			channelRoute.GET("/tag/models", controller.GetTagModels)
			channelRoute.POST("/copy/:id", controller.CopyChannel)
			channelRoute.POST("/multi_key/manage", controller.ManageMultiKeys)
		}
		tokenRoute := apiRouter.Group("/token")
		tokenRoute.Use(middleware.UserAuth())
		{
			tokenRoute.GET("/", controller.GetAllTokens)
			tokenRoute.GET("/search", controller.SearchTokens)
			tokenRoute.GET("/:id", controller.GetToken)
			tokenRoute.POST("/", controller.AddToken)
			tokenRoute.PUT("/", controller.UpdateToken)
			tokenRoute.DELETE("/:id", controller.DeleteToken)
			tokenRoute.POST("/batch", controller.DeleteTokenBatch)
		}

		usageRoute := apiRouter.Group("/usage")
		usageRoute.Use(middleware.CriticalRateLimit())
		{
			usageRoute.GET("/self/limits", middleware.UserAuth(), controller.GetSelfUsageLimits)
			tokenUsageRoute := usageRoute.Group("/token")
			tokenUsageRoute.Use(middleware.TokenAuth())
			{
				tokenUsageRoute.GET("/", controller.GetTokenUsage)
			}
		}

		redemptionReadRoute := apiRouter.Group("/redemption")
		redemptionReadRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
		{
			redemptionReadRoute.GET("/", controller.GetAllRedemptions)
			redemptionReadRoute.GET("/search", controller.SearchRedemptions)
			redemptionReadRoute.GET("/:id", controller.GetRedemption)
		}
		redemptionRoute := apiRouter.Group("/redemption")
		redemptionRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
		{
			redemptionRoute.POST("/", controller.AddRedemption)
			redemptionRoute.PUT("/", controller.UpdateRedemption)
			redemptionRoute.DELETE("/invalid", controller.DeleteInvalidRedemption)
			redemptionRoute.DELETE("/:id", controller.DeleteRedemption)
		}
		topUpCouponReadRoute := apiRouter.Group("/topup-coupon")
		topUpCouponReadRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
		{
			topUpCouponReadRoute.GET("/", controller.GetAllTopUpCoupons)
			topUpCouponReadRoute.GET("/search", controller.SearchTopUpCoupons)
			topUpCouponReadRoute.GET("/:id", controller.GetTopUpCoupon)
		}
		topUpCouponRoute := apiRouter.Group("/topup-coupon")
		topUpCouponRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
		{
			topUpCouponRoute.POST("/", controller.AddTopUpCoupon)
			topUpCouponRoute.PUT("/", controller.UpdateTopUpCoupon)
		}
		logRoute := apiRouter.Group("/log")
		logRoute.GET("/", middleware.SupportAuth(), middleware.AdminAudit(), controller.GetAllLogs)
		logRoute.DELETE("/", middleware.AdminAuth(), middleware.AdminAudit(), controller.DeleteHistoryLogs)
		logRoute.GET("/stat", middleware.SupportAuth(), middleware.AdminAudit(), controller.GetLogsStat)
		logRoute.GET("/self/stat", middleware.UserAuth(), controller.GetLogsSelfStat)
		logRoute.GET("/search", middleware.SupportAuth(), middleware.AdminAudit(), controller.SearchAllLogs)
		logRoute.GET("/self", middleware.UserAuth(), controller.GetUserLogs)
		logRoute.GET("/self/search", middleware.UserAuth(), controller.SearchUserLogs)

		dataRoute := apiRouter.Group("/data")
		dataRoute.GET("/", middleware.SupportAuth(), middleware.AdminAudit(), controller.GetAllQuotaDates)
		dataRoute.GET("/self", middleware.UserAuth(), controller.GetUserQuotaDates)

		logRoute.Use(middleware.CORS())
		{
			logRoute.GET("/token", controller.GetLogByKey)
		}
		groupRoute := apiRouter.Group("/group")
		groupRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
		{
			groupRoute.GET("/", controller.GetGroups)
		}

		prefillGroupReadRoute := apiRouter.Group("/prefill_group")
		prefillGroupReadRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
		{
			prefillGroupReadRoute.GET("/", controller.GetPrefillGroups)
		}
		prefillGroupRoute := apiRouter.Group("/prefill_group")
		prefillGroupRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
		{
			prefillGroupRoute.POST("/", controller.CreatePrefillGroup)
			prefillGroupRoute.PUT("/", controller.UpdatePrefillGroup)
			prefillGroupRoute.DELETE("/:id", controller.DeletePrefillGroup)
		}

		mjRoute := apiRouter.Group("/mj")
		mjRoute.GET("/self", middleware.UserAuth(), controller.GetUserMidjourney)
		mjRoute.GET("/", middleware.SupportAuth(), middleware.AdminAudit(), controller.GetAllMidjourney)

		taskRoute := apiRouter.Group("/task")
		{
			taskRoute.GET("/self", middleware.UserAuth(), controller.GetUserTask)
			taskRoute.GET("/", middleware.SupportAuth(), middleware.AdminAudit(), controller.GetAllTask)
		}

		vendorReadRoute := apiRouter.Group("/vendors")
		vendorReadRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
		{
			vendorReadRoute.GET("/", controller.GetAllVendors)
			vendorReadRoute.GET("/search", controller.SearchVendors)
			vendorReadRoute.GET("/:id", controller.GetVendorMeta)
		}
		vendorRoute := apiRouter.Group("/vendors")
		vendorRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
		{
			vendorRoute.POST("/", controller.CreateVendorMeta)
			vendorRoute.PUT("/", controller.UpdateVendorMeta)
			vendorRoute.DELETE("/:id", controller.DeleteVendorMeta)
		}

		modelsReadRoute := apiRouter.Group("/models")
		modelsReadRoute.Use(middleware.SupportAuth(), middleware.AdminAudit())
		{
			modelsReadRoute.GET("/sync_upstream/preview", controller.SyncUpstreamPreview)
			modelsReadRoute.GET("/missing", controller.GetMissingModels)
			modelsReadRoute.GET("/", controller.GetAllModelsMeta)
			modelsReadRoute.GET("/search", controller.SearchModelsMeta)
			modelsReadRoute.GET("/:id", controller.GetModelMeta)
		}
		modelsRoute := apiRouter.Group("/models")
		modelsRoute.Use(middleware.AdminAuth(), middleware.AdminAudit())
		{
			modelsRoute.POST("/sync_upstream", controller.SyncUpstreamModels)
			modelsRoute.POST("/", controller.CreateModelMeta)
			modelsRoute.PUT("/", controller.UpdateModelMeta)
			modelsRoute.DELETE("/:id", controller.DeleteModelMeta)
		}

		// SSO API routes
		ssoRoute := apiRouter.Group("/sso-beta")
		{
			ssoRoute.POST("/approve", middleware.UserAuth(), controller.SSOApprove)
			ssoRoute.POST("/cancel", controller.SSOCancel)
		}

		// Beta API routes
		betaRoute := apiRouter.Group("/beta")
		betaRoute.Use(middleware.UserAuth())
		{
			betaRoute.GET("/remain_actual_paid_amount", controller.GetRemainActualPaidAmount)
			betaRoute.PUT("/self_quota_take_away", controller.SelfQuotaTakeAway)
		}
	}
}
