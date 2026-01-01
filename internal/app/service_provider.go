package app

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/creafly/identity/internal/config"
	"github.com/creafly/identity/internal/domain/repository"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/handler"
	"github.com/creafly/identity/internal/infra/database"
	"github.com/creafly/identity/internal/infra/kafka"
	"github.com/creafly/identity/internal/infra/outbox"
	"github.com/creafly/identity/internal/logger"
	"github.com/creafly/identity/pkg/featureflags"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/xlab/closer"
)

type serviceProvider struct {
	cfg *config.Config

	db       *sqlx.DB
	migrator *database.Migrator

	kafkaProducer sarama.SyncProducer

	outboxEventHandler outbox.EventHandler
	outboxWorker       *outbox.Worker

	userRepo              repository.UserRepository
	roleRepo              repository.RoleRepository
	claimRepo             repository.ClaimRepository
	passwordResetRepo     repository.PasswordResetRepository
	emailVerificationRepo repository.EmailVerificationRepository
	tenantRepo            repository.TenantRepository
	tenantRoleRepo        repository.TenantRoleRepository
	outboxRepo            repository.OutboxRepository
	analyticsRepo         repository.AnalyticsRepository

	invitationsConsumer *kafka.InvitationsConsumer

	userSvc              service.UserService
	roleSvc              service.RoleService
	claimSvc             service.ClaimService
	passwordResetSvc     service.PasswordResetService
	emailVerificationSvc service.EmailVerificationService
	invitationSvc        service.InvitationService
	tenantSvc            service.TenantService
	tenantRoleSvc        service.TenantRoleService
	tokenSvc             service.TokenService
	totpSvc              service.TOTPService
	analyticsSvc         service.AnalyticsService

	authHnd       *handler.AuthHandler
	userHnd       *handler.UserHandler
	roleHnd       *handler.RoleHandler
	claimHnd      *handler.ClaimHandler
	tenantHnd     *handler.TenantHandler
	tenantRoleHnd *handler.TenantRoleHandler
	totpHnd       *handler.TOTPHandler
	healthHnd     *handler.HealthHandler
	analyticsHnd  *handler.AnalyticsHandler

	featureFlags *featureflags.Client

	httpEngine *gin.Engine
}

func NewServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (sp *serviceProvider) GetConfig() *config.Config {
	if sp.cfg == nil {
		sp.cfg = config.Load()
	}
	return sp.cfg
}

func (sp *serviceProvider) GetDB() *sqlx.DB {
	if sp.db == nil {
		db, err := sqlx.Connect("pgx", sp.GetConfig().Database.URL)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect to database")
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		sp.db = db

		closer.Bind(func() {
			sp.db.Close()
		})
	}
	return sp.db
}

func (sp *serviceProvider) GetOutboxEventHandler() outbox.EventHandler {
	if sp.outboxEventHandler == nil {
		var outboxHandler outbox.EventHandler
		if sp.GetConfig().Kafka.Enabled && sp.GetKafkaProducer() != nil {
			outboxHandler = outbox.NewKafkaEventHandler(sp.GetKafkaProducer())
		} else {
			outboxHandler = outbox.NewLogEventHandler()
		}

		sp.outboxEventHandler = outboxHandler
	}

	return sp.outboxEventHandler
}

func (sp *serviceProvider) GetOutboxWorker() *outbox.Worker {
	if sp.outboxWorker == nil {
		sp.outboxWorker = outbox.NewWorker(sp.GetOutboxRepo(), sp.GetOutboxEventHandler(), outbox.DefaultConfig())
		closer.Bind(sp.outboxWorker.Stop)
	}
	return sp.outboxWorker
}

func (sp *serviceProvider) GetKafkaProducer() sarama.SyncProducer {
	if sp.kafkaProducer == nil && sp.GetConfig().Kafka.Enabled {
		kafkaConfig := sarama.NewConfig()
		kafkaConfig.Producer.Return.Successes = true
		kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
		kafkaConfig.Producer.Retry.Max = 3
		producer, err := sarama.NewSyncProducer(sp.GetConfig().Kafka.Brokers, kafkaConfig)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create Kafka producer, using log handler")
		}

		sp.kafkaProducer = producer

		closer.Bind(func() {
			if err := sp.kafkaProducer.Close(); err != nil {
				logger.Error().Err(err).Msg("Error closing Kafka producer")
			}
		})
	}
	return sp.kafkaProducer
}

func (sp *serviceProvider) GetMigrator() *database.Migrator {
	if sp.migrator == nil {
		sp.migrator = database.NewMigrator(sp.GetDB(), "migrations")
	}
	return sp.migrator
}

func (sp *serviceProvider) GetInvitationsConsumer() *kafka.InvitationsConsumer {
	if sp.invitationsConsumer == nil && sp.GetConfig().Kafka.Enabled {
		consumer, err := kafka.NewInvitationsConsumer(
			sp.GetConfig().Kafka.Brokers,
			sp.GetConfig().Kafka.GroupID,
			sp.GetTenantSvc(),
		)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create Kafka consumer")
			return nil
		}

		sp.invitationsConsumer = consumer
		closer.Bind(func() {
			sp.invitationsConsumer.Stop()
		})
	}
	return sp.invitationsConsumer
}

func (sp *serviceProvider) GetUserRepo() repository.UserRepository {
	if sp.userRepo == nil {
		sp.userRepo = repository.NewUserRepository(sp.GetDB())
	}
	return sp.userRepo
}

func (sp *serviceProvider) GetRoleRepo() repository.RoleRepository {
	if sp.roleRepo == nil {
		sp.roleRepo = repository.NewRoleRepository(sp.GetDB())
	}
	return sp.roleRepo
}

func (sp *serviceProvider) GetClaimRepo() repository.ClaimRepository {
	if sp.claimRepo == nil {
		sp.claimRepo = repository.NewClaimRepository(sp.GetDB())
	}
	return sp.claimRepo
}

func (sp *serviceProvider) GetPasswordResetRepo() repository.PasswordResetRepository {
	if sp.passwordResetRepo == nil {
		sp.passwordResetRepo = repository.NewPasswordResetRepository(sp.GetDB())
	}
	return sp.passwordResetRepo
}

func (sp *serviceProvider) GetEmailVerificationRepo() repository.EmailVerificationRepository {
	if sp.emailVerificationRepo == nil {
		sp.emailVerificationRepo = repository.NewEmailVerificationRepository(sp.GetDB())
	}
	return sp.emailVerificationRepo
}

func (sp *serviceProvider) GetTenantRepo() repository.TenantRepository {
	if sp.tenantRepo == nil {
		sp.tenantRepo = repository.NewTenantRepository(sp.GetDB())
	}
	return sp.tenantRepo
}

func (sp *serviceProvider) GetTenantRoleRepo() repository.TenantRoleRepository {
	if sp.tenantRoleRepo == nil {
		sp.tenantRoleRepo = repository.NewTenantRoleRepository(sp.GetDB())
	}
	return sp.tenantRoleRepo
}

func (sp *serviceProvider) GetOutboxRepo() repository.OutboxRepository {
	if sp.outboxRepo == nil {
		sp.outboxRepo = repository.NewOutboxRepository(sp.GetDB())
	}
	return sp.outboxRepo
}

func (sp *serviceProvider) GetAnalyticsRepo() repository.AnalyticsRepository {
	if sp.analyticsRepo == nil {
		sp.analyticsRepo = repository.NewAnalyticsRepository(sp.GetDB())
	}
	return sp.analyticsRepo
}

func (sp *serviceProvider) GetUserSvc() service.UserService {
	if sp.userSvc == nil {
		sp.userSvc = service.NewUserService(sp.GetUserRepo())
	}
	return sp.userSvc
}

func (sp *serviceProvider) GetRoleSvc() service.RoleService {
	if sp.roleSvc == nil {
		sp.roleSvc = service.NewRoleService(sp.GetRoleRepo())
	}
	return sp.roleSvc
}

func (sp *serviceProvider) GetClaimSvc() service.ClaimService {
	if sp.claimSvc == nil {
		sp.claimSvc = service.NewClaimService(sp.GetClaimRepo(), sp.GetRoleRepo(), sp.GetTenantRoleRepo())
	}
	return sp.claimSvc
}

func (sp *serviceProvider) GetPasswordResetSvc() service.PasswordResetService {
	if sp.passwordResetSvc == nil {
		sp.passwordResetSvc = service.NewPasswordResetService(sp.GetUserRepo(), sp.GetPasswordResetRepo(), sp.GetOutboxRepo())
	}
	return sp.passwordResetSvc
}

func (sp *serviceProvider) GetEmailVerificationSvc() service.EmailVerificationService {
	if sp.emailVerificationSvc == nil {
		sp.emailVerificationSvc = service.NewEmailVerificationService(sp.GetUserRepo(), sp.GetEmailVerificationRepo(), sp.GetOutboxRepo())
	}
	return sp.emailVerificationSvc
}

func (sp *serviceProvider) GetInvitationsSvc() service.InvitationService {
	if sp.invitationSvc == nil {
		sp.invitationSvc = service.NewInvitationService(sp.GetOutboxRepo())
	}
	return sp.invitationSvc
}

func (sp *serviceProvider) GetTenantSvc() service.TenantService {
	if sp.tenantSvc == nil {
		sp.tenantSvc = service.NewTenantService(sp.GetTenantRepo())
	}
	return sp.tenantSvc
}

func (sp *serviceProvider) GetTenantRoleSvc() service.TenantRoleService {
	if sp.tenantRoleSvc == nil {
		sp.tenantRoleSvc = service.NewTenantRoleService(sp.GetTenantRoleRepo(), sp.GetClaimRepo())
	}
	return sp.tenantRoleSvc
}

func (sp *serviceProvider) GetTokenSvc() service.TokenService {
	if sp.tokenSvc == nil {
		sp.tokenSvc = service.NewTokenService(sp.GetConfig())
	}
	return sp.tokenSvc
}

func (sp *serviceProvider) GetTOTPSvc() service.TOTPService {
	if sp.totpSvc == nil {
		sp.totpSvc = service.NewTOTPService(sp.GetUserRepo(), sp.GetUserSvc())
	}
	return sp.totpSvc
}

func (sp *serviceProvider) GetAnalyticsSvc() service.AnalyticsService {
	if sp.analyticsSvc == nil {
		sp.analyticsSvc = service.NewAnalyticsService(sp.GetAnalyticsRepo())
	}
	return sp.analyticsSvc
}

func (sp *serviceProvider) GetAuthHnd() *handler.AuthHandler {
	if sp.authHnd == nil {
		sp.authHnd = handler.NewAuthHandler(sp.cfg, sp.GetUserSvc(), sp.GetTokenSvc(), sp.GetRoleSvc(), sp.GetTOTPSvc(), sp.GetPasswordResetSvc(), sp.GetEmailVerificationSvc(), sp.GetClaimSvc())
	}
	return sp.authHnd
}

func (sp *serviceProvider) GetUserHnd() *handler.UserHandler {
	if sp.userHnd == nil {
		sp.userHnd = handler.NewUserHandler(sp.GetUserSvc())
	}
	return sp.userHnd
}

func (sp *serviceProvider) GetRoleHnd() *handler.RoleHandler {
	if sp.roleHnd == nil {
		sp.roleHnd = handler.NewRoleHandler(sp.GetRoleSvc())
	}
	return sp.roleHnd
}

func (sp *serviceProvider) GetClaimHnd() *handler.ClaimHandler {
	if sp.claimHnd == nil {
		sp.claimHnd = handler.NewClaimHandler(sp.GetClaimSvc(), sp.GetTenantSvc())
	}
	return sp.claimHnd
}

func (sp *serviceProvider) GetTenantHnd() *handler.TenantHandler {
	if sp.tenantHnd == nil {
		sp.tenantHnd = handler.NewTenantHandler(sp.GetTenantSvc(), sp.GetTenantRoleSvc(), sp.GetInvitationsSvc(), sp.GetUserSvc())
	}
	return sp.tenantHnd
}

func (sp *serviceProvider) GetTenantRoleHnd() *handler.TenantRoleHandler {
	if sp.tenantRoleHnd == nil {
		sp.tenantRoleHnd = handler.NewTenantRoleHandler(sp.GetTenantRoleSvc())
	}
	return sp.tenantRoleHnd
}

func (sp *serviceProvider) GetTOTPHnd() *handler.TOTPHandler {
	if sp.totpHnd == nil {
		sp.totpHnd = handler.NewTOTPHandler(sp.GetTOTPSvc(), sp.GetUserSvc())
	}
	return sp.totpHnd
}

func (sp *serviceProvider) GetHealthHnd() *handler.HealthHandler {
	if sp.healthHnd == nil {
		sp.healthHnd = handler.NewHealthHandler()
	}
	return sp.healthHnd
}

func (sp *serviceProvider) GetAnalyticsHnd() *handler.AnalyticsHandler {
	if sp.analyticsHnd == nil {
		sp.analyticsHnd = handler.NewAnalyticsHandler(sp.GetAnalyticsSvc())
	}
	return sp.analyticsHnd
}

func (sp *serviceProvider) GetHttpEngine() *gin.Engine {
	if sp.httpEngine == nil {
		sp.httpEngine = gin.New()
	}
	return sp.httpEngine
}

func (sp *serviceProvider) GetFeatureFlags() *featureflags.Client {
	if sp.featureFlags == nil && sp.GetConfig().Unleash.Enabled {
		cfg := featureflags.Config{
			URL:         sp.GetConfig().Unleash.URL,
			AppName:     sp.GetConfig().Unleash.AppName,
			APIToken:    sp.GetConfig().Unleash.APIToken,
			Environment: sp.GetConfig().Tracing.Environment,
		}

		client, err := featureflags.New(cfg)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create feature flags client")
			return nil
		}

		sp.featureFlags = client

		closer.Bind(func() {
			sp.featureFlags.Close()
		})
	}
	return sp.featureFlags
}
