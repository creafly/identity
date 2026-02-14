package app

import (
	"context"
	"net/http"

	"github.com/creafly/featureflags"
	"github.com/creafly/identity/internal/i18n"
	"github.com/creafly/identity/internal/middleware"
	"github.com/creafly/identity/internal/validator"
	"github.com/creafly/logger"
	sharedmw "github.com/creafly/middleware"
	"github.com/creafly/tracing"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xlab/closer"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type App struct {
	ServiceProvider *serviceProvider
	HttpServer      *http.Server
}

func NewApp() *App {
	a := (&App{}).initBaseApp()
	a.initHttpServer()
	return a
}

func NewMigratorApp() *App {
	return (&App{}).initBaseApp()
}

func (a *App) StartApp(ctx context.Context) {
	cfg := a.ServiceProvider.GetConfig()

	tracingShutdown, err := tracing.Init(tracing.Config{
		ServiceName:    cfg.Tracing.ServiceName,
		ServiceVersion: cfg.Tracing.ServiceVersion,
		Environment:    cfg.Tracing.Environment,
		OTLPEndpoint:   cfg.Tracing.OTLPEndpoint,
		Enabled:        cfg.Tracing.Enabled,
	})
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize tracing")
	} else {
		closer.Bind(func() {
			if err := tracingShutdown(context.Background()); err != nil {
				logger.Error().Err(err).Msg("Error shutting down tracer provider")
			}
		})
	}

	i18n.PreloadLocales()

	outboxWorker := a.ServiceProvider.GetOutboxWorker()
	outboxWorker.Start(context.Background())

	invitationsConsumer := a.ServiceProvider.GetInvitationsConsumer()
	invitationsConsumer.Start(ctx)

	go func() {
		if err := a.getHttpServer().ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
}

func (a *App) StartMigrator(migrateUp, migrateDown bool) {
	migrator := a.ServiceProvider.GetMigrator()

	if migrateUp {
		if err := migrator.Up(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run migrations up")
		}
		logger.Info().Msg("Migrations completed successfully")
		return
	}

	if migrateDown {
		if err := migrator.Down(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run migrations down")
		}
		logger.Info().Msg("Migrations rolled back successfully")
		return
	}

	if a.ServiceProvider.GetConfig().Database.AutoMigrate {
		logger.Info().Msg("Running auto-migrations...")
		if err := migrator.Up(); err != nil {
			logger.Warn().Err(err).Msg("Auto-migration failed")
		}
	}
}

func (a *App) initBaseApp() *App {
	_ = godotenv.Load()
	validator.Init()
	logger.InitFromEnv("identity")
	closer.Init(closer.Config{
		ExitSignals: closer.DefaultSignalSet,
	})

	a.initServiceProvider()

	return a
}

func (a *App) getHttpServer() *http.Server {
	if a.HttpServer == nil {
		addr := a.ServiceProvider.GetConfig().Server.Host + ":" + a.ServiceProvider.GetConfig().Server.Port
		logger.Info().Str("addr", addr).Msg("Starting Identity Service")

		a.HttpServer = &http.Server{
			Addr:    addr,
			Handler: a.ServiceProvider.GetHttpEngine(),
		}

		closer.Bind(func() {
			if err := a.HttpServer.Shutdown(context.Background()); err != nil {
				logger.Error().Err(err).Msg("Server forced to shutdown")
			}
		})
	}

	return a.HttpServer
}

func (a *App) initHttpServer() {
	gin.SetMode(a.ServiceProvider.GetConfig().Server.GinMode)

	a.initHttpMiddleware()
	a.initHttpRouting()
}

func (a *App) initServiceProvider() {
	if a.ServiceProvider == nil {
		a.ServiceProvider = NewServiceProvider()
	}
}

func (a *App) initHttpMiddleware() {
	r := a.ServiceProvider.GetHttpEngine()
	cfg := a.ServiceProvider.GetConfig()

	r.Use(gin.Recovery())
	r.Use(sharedmw.RequestID())
	r.Use(sharedmw.SecurityHeaders())
	r.Use(sharedmw.HSTS(cfg.Server.GinMode == "release"))
	r.Use(sharedmw.ContentTypeValidation())
	r.Use(sharedmw.RateLimit(sharedmw.RateLimitConfig{
		Enabled:           cfg.RateLimit.Enabled,
		RequestsPerSecond: cfg.RateLimit.RequestsPerSecond,
		BurstSize:         cfg.RateLimit.BurstSize,
	}))
	r.Use(otelgin.Middleware("identity"))
	r.Use(sharedmw.Logging())
	r.Use(middleware.PrometheusMiddleware())
	r.Use(sharedmw.Locale())
	r.Use(sharedmw.CORS(sharedmw.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))
	r.Use(sharedmw.Compression())

	if ffClient := a.ServiceProvider.GetFeatureFlags(); ffClient != nil {
		r.Use(featureflags.GinMiddleware(ffClient))
	}
}

func (a *App) initHttpRouting() {
	r := a.ServiceProvider.GetHttpEngine()

	tokenService := a.ServiceProvider.GetTokenSvc()
	tenantService := a.ServiceProvider.GetTenantSvc()
	claimService := a.ServiceProvider.GetClaimSvc()
	userRepo := a.ServiceProvider.GetUserRepo()
	ffClient := a.ServiceProvider.GetFeatureFlags()

	authHandler := a.ServiceProvider.GetAuthHnd()
	healthHandler := a.ServiceProvider.GetHealthHnd()
	tenantHandler := a.ServiceProvider.GetTenantHnd()
	tenantRoleHandler := a.ServiceProvider.GetTenantRoleHnd()
	roleHandler := a.ServiceProvider.GetRoleHnd()
	claimHandler := a.ServiceProvider.GetClaimHnd()
	totpHandler := a.ServiceProvider.GetTOTPHnd()
	userHandler := a.ServiceProvider.GetUserHnd()
	analyticsHandler := a.ServiceProvider.GetAnalyticsHnd()

	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			registerHandlers := []gin.HandlerFunc{authHandler.Register}
			if ffClient != nil {
				registerHandlers = append(
					[]gin.HandlerFunc{featureflags.BlockIfGlobalEnabled(ffClient, "disable_registration", "Registration is currently disabled")},
					registerHandlers...,
				)
			}
			auth.POST("/register", registerHandlers...)
			auth.POST("/login", authHandler.Login)
			auth.POST("/login/verify-totp", authHandler.LoginVerifyTOTP)
			auth.POST("/refresh", authHandler.Refresh)
			auth.GET("/verify", authHandler.Verify)
			auth.POST("/forgot-password", sharedmw.StrictRateLimit(sharedmw.RateLimitConfig{
				Enabled:           a.ServiceProvider.GetConfig().RateLimit.Enabled,
				RequestsPerSecond: a.ServiceProvider.GetConfig().RateLimit.RequestsPerSecond,
				BurstSize:         a.ServiceProvider.GetConfig().RateLimit.BurstSize,
			}), authHandler.ForgotPassword)
			auth.POST("/reset-password", sharedmw.StrictRateLimit(sharedmw.RateLimitConfig{
				Enabled:           a.ServiceProvider.GetConfig().RateLimit.Enabled,
				RequestsPerSecond: a.ServiceProvider.GetConfig().RateLimit.RequestsPerSecond,
				BurstSize:         a.ServiceProvider.GetConfig().RateLimit.BurstSize,
			}), authHandler.ResetPassword)
		}

		v1.GET("/tenants/resolve/:slug", tenantHandler.ResolveSlug)

		internal := v1.Group("/internal")
		internal.Use(sharedmw.InternalAPI(sharedmw.InternalAPIConfig{
			APIKey:          a.ServiceProvider.GetConfig().InternalAPI.APIKey,
			AllowedServices: a.ServiceProvider.GetConfig().InternalAPI.AllowedServices,
		}))
		{
			internal.POST("/tenants/members/callback", tenantHandler.AddMemberCallback)
			internal.GET("/tenants/resolve/:slug", tenantHandler.ResolveSlug)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(tokenService))
		protected.Use(middleware.BlockedUserMiddleware(userRepo))
		{
			protected.GET("/me", authHandler.Me)
			protected.PUT("/me", authHandler.UpdateProfile)
			protected.POST("/change-password", authHandler.ChangePassword)
			protected.POST("/logout", authHandler.Logout)
			protected.POST("/logout-all", authHandler.LogoutAll)
			protected.POST("/verify-email", authHandler.VerifyEmail)
			protected.POST("/resend-verification", authHandler.ResendVerificationEmail)
			protected.GET("/my-tenants", tenantHandler.GetMyTenants)
			protected.GET("/my-claims", claimHandler.GetMyClaims)

			totp := protected.Group("/2fa")
			{
				totp.GET("/status", totpHandler.Status)
				totp.POST("/setup", totpHandler.Setup)
				totp.POST("/enable", totpHandler.Enable)
				totp.POST("/disable", totpHandler.Disable)
				totp.POST("/validate", totpHandler.Validate)
			}

			tenants := protected.Group("/tenants")
			{
				tenants.POST("", tenantHandler.Create)
				tenants.GET("", tenantHandler.List)
				tenants.GET("/:id/validate", tenantHandler.ValidateTenantAccess)
			}

			tenantView := protected.Group("/tenants/:id")
			tenantView.Use(middleware.RequireAnyClaimWithTenant(claimService, tenantService, "tenant:view", "tenant:manage"))
			{
				tenantView.GET("", tenantHandler.GetByID)
				tenantView.GET("/members", tenantHandler.GetMembers)
				tenantView.GET("/users/:userId/roles", tenantHandler.GetUserRoles)
			}

			tenantManage := protected.Group("/tenants/:id")
			tenantManage.Use(middleware.RequireClaimsWithTenant(claimService, tenantService, "tenant:manage"))
			{
				tenantManage.PUT("", tenantHandler.Update)
				tenantManage.DELETE("", tenantHandler.Delete)
			}

			tenantRolesView := protected.Group("/tenants/:id/roles")
			tenantRolesView.Use(middleware.RequireAnyClaimWithTenant(claimService, tenantService, "tenant:roles:view", "tenant:roles:manage"))
			{
				tenantRolesView.GET("", tenantRoleHandler.List)
				tenantRolesView.GET("/:roleId", tenantRoleHandler.GetByID)
				tenantRolesView.GET("/:roleId/claims", tenantRoleHandler.GetRoleClaims)
			}

			tenantClaimsView := protected.Group("/tenants/:id")
			tenantClaimsView.Use(middleware.RequireAnyClaimWithTenant(claimService, tenantService, "tenant:roles:view", "tenant:roles:manage"))
			{
				tenantClaimsView.GET("/claims", tenantRoleHandler.GetAvailableClaims)
			}

			tenantRolesManage := protected.Group("/tenants/:id/roles")
			tenantRolesManage.Use(middleware.RequireClaimsWithTenant(claimService, tenantService, "tenant:roles:manage"))
			{
				tenantRolesManage.POST("", tenantRoleHandler.Create)
				tenantRolesManage.PUT("/:roleId", tenantRoleHandler.Update)
				tenantRolesManage.DELETE("/:roleId", tenantRoleHandler.Delete)
				tenantRolesManage.POST("/:roleId/restore", tenantRoleHandler.Restore)
				tenantRolesManage.POST("/:roleId/claims", tenantRoleHandler.AssignClaim)
				tenantRolesManage.DELETE("/:roleId/claims/:claimId", tenantRoleHandler.RemoveClaim)
				tenantRolesManage.PUT("/:roleId/claims", tenantRoleHandler.BatchUpdateClaims)
			}

			tenantMembersView := protected.Group("/tenants/:id")
			tenantMembersView.Use(middleware.RequireAnyClaimWithTenant(claimService, tenantService, "tenant:members:view", "tenant:members:manage"))
			{
			}

			tenantMembersManage := protected.Group("/tenants/:id")
			tenantMembersManage.Use(middleware.RequireClaimsWithTenant(claimService, tenantService, "tenant:members:manage"))
			{
				tenantMembersManage.POST("/invite", tenantHandler.InviteMember)
				tenantMembersManage.DELETE("/members/:userId", tenantHandler.RemoveMember)
				tenantMembersManage.POST("/users/:userId/roles", tenantHandler.AssignRoleToTenantUser)
				tenantMembersManage.DELETE("/users/:userId/roles/:roleId", tenantHandler.RemoveRoleFromTenantUser)
				tenantMembersManage.PUT("/users/:userId/roles", tenantHandler.BatchUpdateTenantUserRoles)
			}

			adminRoles := protected.Group("/roles")
			adminRoles.Use(middleware.RequireAnyClaim(claimService, "roles:view", "roles:manage"))
			{
				adminRoles.GET("", roleHandler.List)
				adminRoles.GET("/:id", roleHandler.GetByID)
				adminRoles.GET("/:id/claims", claimHandler.GetRoleClaims)
			}

			adminRolesManage := protected.Group("/roles")
			adminRolesManage.Use(middleware.RequireClaims(claimService, "roles:manage"))
			{
				adminRolesManage.POST("", roleHandler.Create)
				adminRolesManage.PUT("/:id", roleHandler.Update)
				adminRolesManage.DELETE("/:id", roleHandler.Delete)
				adminRolesManage.POST("/:id/restore", roleHandler.Restore)
				adminRolesManage.POST("/:id/claims", claimHandler.AssignToRole)
				adminRolesManage.DELETE("/:id/claims/:claimId", claimHandler.RemoveFromRole)
			}

			adminClaims := protected.Group("/claims")
			adminClaims.Use(middleware.RequireAnyClaim(claimService, "claims:view", "claims:manage"))
			{
				adminClaims.GET("", claimHandler.List)
				adminClaims.GET("/:id", claimHandler.GetByID)
			}

			adminClaimsManage := protected.Group("/claims")
			adminClaimsManage.Use(middleware.RequireClaims(claimService, "claims:manage"))
			{
				adminClaimsManage.POST("", claimHandler.Create)
				adminClaimsManage.DELETE("/:id", claimHandler.Delete)
			}

			adminUsers := protected.Group("/users")
			adminUsers.Use(middleware.RequireAnyClaim(claimService, "users:view", "users:manage"))
			{
				adminUsers.GET("", userHandler.List)
				adminUsers.GET("/:userId", userHandler.GetByID)
				adminUsers.GET("/:userId/roles", roleHandler.GetUserRoles)
				adminUsers.GET("/:userId/claims", claimHandler.GetUserClaims)
			}

			adminUsersManage := protected.Group("/users")
			adminUsersManage.Use(middleware.RequireClaims(claimService, "users:manage"))
			{
				adminUsersManage.POST("/:userId/block", userHandler.Block)
				adminUsersManage.POST("/:userId/unblock", userHandler.Unblock)
				adminUsersManage.POST("/:userId/roles", roleHandler.AssignToUser)
				adminUsersManage.DELETE("/:userId/roles/:roleId", roleHandler.RemoveFromUser)
				adminUsersManage.POST("/:userId/claims", claimHandler.AssignToUser)
				adminUsersManage.DELETE("/:userId/claims/:claimId", claimHandler.RemoveFromUser)
			}

			adminAnalytics := protected.Group("/admin")
			adminAnalytics.Use(middleware.RequireAnyClaim(claimService, "users:view", "users:manage"))
			{
				adminAnalytics.GET("/analytics", analyticsHandler.GetAnalytics)
			}

			adminTenants := protected.Group("/admin/tenants")
			adminTenants.Use(middleware.RequireClaims(claimService, "tenants:manage"))
			{
				adminTenants.POST("/:id/block", tenantHandler.Block)
				adminTenants.POST("/:id/unblock", tenantHandler.Unblock)
			}
		}
	}
}
