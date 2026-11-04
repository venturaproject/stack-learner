package di

import (
	"context"
	"time"

	apiclients "boilerplate-learner/internal/application/api_clients"
	"boilerplate-learner/internal/application/assessment"
	"boilerplate-learner/internal/application/auth"
	"boilerplate-learner/internal/application/dashboard"
	"boilerplate-learner/internal/application/dispositivos"
	"boilerplate-learner/internal/application/learning"
	"boilerplate-learner/internal/application/telefonos"
	"boilerplate-learner/internal/application/trabajadores"
	"boilerplate-learner/internal/config"
	"boilerplate-learner/internal/infrastructure/cache"
	"boilerplate-learner/internal/infrastructure/events"
	apphttp "boilerplate-learner/internal/infrastructure/http"
	"boilerplate-learner/internal/infrastructure/http/handlers"
	"boilerplate-learner/internal/infrastructure/persistence"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Container struct {
	Router *gin.Engine
	Log    *zap.Logger
	Redis  *cache.Redis
}

func Build(cfg *config.Config, log *zap.Logger) (*Container, error) {
	// Infrastructure
	db, err := persistence.NewDB(cfg)
	if err != nil {
		return nil, err
	}

	rdb, err := cache.NewRedis(cfg)
	if err != nil {
		log.Warn("Redis no disponible, continuando sin caché", zap.Error(err))
	}

	bus := events.NewBus(log)

	// Transactional outbox: events are written in-transaction and delivered
	// asynchronously by a background dispatcher onto the in-process bus.
	// Subscribe domain-event consumers (projections, notifications, …) here.
	bus.Subscribe(func(e any) error {
		if env, ok := e.(persistence.OutboxEnvelope); ok {
			log.Info("domain event", zap.String("event", env.EventName), zap.Time("occurred_at", env.OccurredAt))
		}
		return nil
	})
	dispatcher := persistence.NewOutboxDispatcher(db, bus, log)
	go dispatcher.Start(context.Background())

	txMgr := persistence.NewTxManager(db)
	publisher := persistence.NewOutboxWriter(db)

	// Repositories
	authRepo := persistence.NewAuthRepository(db)
	revokedTokenRepo := persistence.NewRevokedTokenRepository(db)
	trabRepo := persistence.NewTrabajadorRepository(db)
	trabReadRepo := persistence.NewTrabajadorReadRepository(db)
	telfRepo := persistence.NewTelefonoRepository(db)
	telfActRepo := persistence.NewActuacionTelefonoRepository(db)
	telfReadRepo := persistence.NewTelefonoReadRepository(db)
	dispRepo := persistence.NewDispositivoRepository(db)
	dispCatalogRepo := persistence.NewDispositivoCatalogRepository(db)
	dispReadRepo := persistence.NewDispositivoReadRepository(db)
	apiClientRepo := persistence.NewApiClientRepository(db)

	// Auth services (split by responsibility)
	jwtSvc := auth.NewJWTService(auth.JWTConfig{
		Secret:         cfg.JWT.Secret,
		AccessTTLSecs:  cfg.JWT.AccessTTLSecs,
		RefreshTTLSecs: cfg.JWT.RefreshTTLSecs,
	})
	authSvc := auth.NewAuthService(authRepo, authRepo, jwtSvc, txMgr, publisher)
	userSvc := auth.NewUserService(authRepo, authRepo, txMgr, publisher)
	roleSvc := auth.NewRoleService(authRepo)
	permSvc := auth.NewPermissionService(authRepo)

	// API clients service
	apiClientSvc := apiclients.NewService(apiClientRepo, txMgr)
	learningSvc := learning.NewService(persistence.NewLearningRepository(db), txMgr)
	assessmentSvc := assessment.NewService(persistence.NewAssessmentRepository(db), txMgr)

	// Trabajadores use cases
	trabUC := trabajadores.NewUseCases(trabRepo, trabReadRepo, txMgr, publisher)

	// Telefonos use cases
	telfUC := telefonos.NewUseCases(telfRepo, telfActRepo, telfReadRepo, txMgr, publisher)

	// Dispositivos use cases
	dispUC := dispositivos.NewUseCases(dispRepo, dispCatalogRepo, dispReadRepo, txMgr, publisher)

	// HTTP handlers
	authH := handlers.NewAuthHandler(authSvc, jwtSvc, revokedTokenRepo, cfg)
	userH := handlers.NewUserHandler(userSvc)
	roleH := handlers.NewRoleHandler(roleSvc)
	permH := handlers.NewPermissionHandler(permSvc)
	trabH := handlers.NewTrabajadorHandler(trabUC)
	telfH := handlers.NewTelefonoHandler(telfUC)
	dispH := handlers.NewDispositivoHandler(dispUC)
	dashH := handlers.NewDashboardHandler(dashboard.NewGetUseCase(persistence.NewDashboardReadRepository(db)))
	apiClientH := handlers.NewApiClientHandler(apiClientSvc)
	extAuthH := handlers.NewExtAuthHandler(apiClientSvc)
	learningH := handlers.NewLearningHandler(learningSvc)
	assessmentH := handlers.NewAssessmentHandler(assessmentSvc)
	extH := handlers.NewExternalHandler(trabUC.List, trabUC.Get, telfUC.List, telfUC.Get, dispUC.List, dispUC.Get)

	// Periodic cleanup of expired revoked access tokens
	go func() {
		ticker := time.NewTicker(time.Hour)
		ctx := context.Background()
		for range ticker.C {
			if _, err := revokedTokenRepo.CleanupExpired(ctx); err != nil {
				log.Warn("Error limpiando tokens revocados", zap.Error(err))
			}
		}
	}()

	// Router
	router := apphttp.NewRouter(
		cfg, rdb, jwtSvc, revokedTokenRepo, authRepo,
		authH, userH, roleH, permH,
		trabH, telfH, dispH, dashH,
		apiClientH, extAuthH, extH, learningH, assessmentH,
	)

	return &Container{
		Router: router,
		Log:    log,
		Redis:  rdb,
	}, nil
}
