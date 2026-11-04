package http

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	appauth "boilerplate-learner/internal/application/auth"
	"boilerplate-learner/internal/config"
	domain "boilerplate-learner/internal/domain/api_clients"
	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/infrastructure/cache"
	"boilerplate-learner/internal/infrastructure/http/handlers"
	"boilerplate-learner/internal/infrastructure/http/middleware"
)

func NewRouter(
	cfg *config.Config,
	redis *cache.Redis,
	jwtSvc *appauth.JWTService,
	revokedRepo domainauth.RevokedTokenRepository,
	authz appauth.ReadRepository,
	authH *handlers.AuthHandler,
	userH *handlers.UserHandler,
	roleH *handlers.RoleHandler,
	permH *handlers.PermissionHandler,
	trabH *handlers.TrabajadorHandler,
	telfH *handlers.TelefonoHandler,
	dispH *handlers.DispositivoHandler,
	dashH *handlers.DashboardHandler,
	apiClientH *handlers.ApiClientHandler,
	extAuthH *handlers.ExtAuthHandler,
	extH *handlers.ExternalHandler,
	learningH *handlers.LearningHandler,
	assessmentH *handlers.AssessmentHandler,
) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-CSRFToken"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))

	authMiddleware := middleware.Auth(jwtSvc, revokedRepo, authz, cfg.Auth.FullAccessRoles)
	csrfMiddleware := middleware.CSRFGuard()
	loginLimiter := middleware.NewRateLimiter(redis, "login", 10, time.Minute)
	refreshLimiter := middleware.NewRateLimiter(redis, "refresh", 60, time.Minute)
	extTokenLimiter := middleware.NewRateLimiter(redis, "ext-token", 30, time.Minute)
	can := middleware.RequirePermission

	api := r.Group("/api")
	api.Use(csrfMiddleware)
	api.Use(middleware.IdempotencyKey(redis))

	// Health
	api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Notifications stub
	api.GET("/notifications", authMiddleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"notifications": []interface{}{}, "count": 0})
	})

	// Auth (self-service — no permission gate)
	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/login", loginLimiter.Middleware(), authH.Login)
		authRoutes.POST("/logout", authH.Logout)
		authRoutes.POST("/refresh", refreshLimiter.Middleware(), authH.Refresh)
		authRoutes.GET("/me", authMiddleware, authH.GetMe)
		authRoutes.PATCH("/me", authMiddleware, authH.UpdateMe)
		authRoutes.GET("/me/avatar", authMiddleware, authH.GetMyAvatar)
		authRoutes.POST("/me/avatar", authMiddleware, authH.UploadAvatar)
	}

	// Users
	userRoutes := api.Group("/users", authMiddleware)
	{
		userRoutes.GET("", can("users.view"), userH.List)
		userRoutes.POST("", can("users.create"), userH.Create)
		userRoutes.GET("/:id", can("users.view"), userH.Get)
		userRoutes.GET("/:id/avatar", can("users.view"), authH.GetUserAvatar)
		userRoutes.PUT("/:id", can("users.edit"), userH.Update)
		userRoutes.PATCH("/:id", can("users.edit"), userH.Update)
		userRoutes.DELETE("/:id", can("users.delete"), userH.Delete)
	}

	// Roles
	roleRoutes := api.Group("/roles", authMiddleware)
	{
		roleRoutes.GET("", can("roles.view"), roleH.List)
		roleRoutes.POST("", can("roles.create"), roleH.Create)
		roleRoutes.GET("/:id", can("roles.view"), roleH.Get)
		roleRoutes.PUT("/:id", can("roles.edit"), roleH.Update)
		roleRoutes.DELETE("/:id", can("roles.delete"), roleH.Delete)
	}

	// Permissions
	permRoutes := api.Group("/permissions", authMiddleware)
	{
		permRoutes.GET("", can("permissions.view"), permH.List)
		permRoutes.POST("", can("permissions.create"), permH.Create)
		permRoutes.DELETE("/:id", can("permissions.delete"), permH.Delete)
	}

	// API Clients (manage external API credentials)
	apiClientRoutes := api.Group("/api-clients", authMiddleware, can("api_clients.manage"))
	{
		apiClientRoutes.GET("", apiClientH.List)
		apiClientRoutes.POST("", apiClientH.Create)
		apiClientRoutes.DELETE("/:id", apiClientH.Revoke)
	}

	// Trabajadores
	trabRoutes := api.Group("/trabajadores", authMiddleware)
	{
		trabRoutes.GET("", can("trabajadores.view"), trabH.List)
		trabRoutes.POST("", can("trabajadores.create"), trabH.Create)
		trabRoutes.GET("/:id", can("trabajadores.view"), trabH.Get)
		trabRoutes.PUT("/:id", can("trabajadores.edit"), trabH.Update)
		trabRoutes.PATCH("/:id", can("trabajadores.edit"), trabH.Update)
		trabRoutes.DELETE("/:id", can("trabajadores.delete"), trabH.Delete)
		trabRoutes.POST("/:id/dar-de-baja", can("trabajadores.edit"), trabH.DarDeBaja)
	}

	// Teléfonos
	telfRoutes := api.Group("/telefonos", authMiddleware)
	{
		telfRoutes.GET("", can("telefonos.view"), telfH.List)
		telfRoutes.POST("", can("telefonos.create"), telfH.Create)
		telfRoutes.GET("/estados", can("telefonos.view"), telfH.ListEstados)
		telfRoutes.GET("/tipologias", can("telefonos.view"), telfH.ListTipologias)
		telfRoutes.DELETE("/actuaciones/:id", can("telefonos.edit"), telfH.DeleteActuacion)
		telfRoutes.GET("/:id", can("telefonos.view"), telfH.Get)
		telfRoutes.PUT("/:id", can("telefonos.edit"), telfH.Update)
		telfRoutes.PATCH("/:id", can("telefonos.edit"), telfH.Update)
		telfRoutes.DELETE("/:id", can("telefonos.delete"), telfH.Delete)
		telfRoutes.POST("/:id/desactivar", can("telefonos.edit"), telfH.Desactivar)
		telfRoutes.POST("/:id/asignar", can("telefonos.edit"), telfH.Asignar)
		telfRoutes.POST("/:id/actuaciones", can("telefonos.edit"), telfH.CreateActuacion)
	}

	// Dispositivos
	dispRoutes := api.Group("/dispositivos", authMiddleware)
	{
		dispRoutes.GET("", can("dispositivos.view"), dispH.List)
		dispRoutes.POST("", can("dispositivos.create"), dispH.Create)
		dispRoutes.GET("/marcas", can("dispositivos.view"), dispH.ListBrands)
		dispRoutes.POST("/marcas", can("dispositivos.edit"), dispH.CreateBrand)
		dispRoutes.PUT("/marcas/:id", can("dispositivos.edit"), dispH.UpdateBrand)
		dispRoutes.PATCH("/marcas/:id", can("dispositivos.edit"), dispH.UpdateBrand)
		dispRoutes.DELETE("/marcas/:id", can("dispositivos.edit"), dispH.DeleteBrand)
		dispRoutes.GET("/modelos", can("dispositivos.view"), dispH.ListModels)
		dispRoutes.POST("/modelos", can("dispositivos.edit"), dispH.CreateModel)
		dispRoutes.PUT("/modelos/:id", can("dispositivos.edit"), dispH.UpdateModel)
		dispRoutes.PATCH("/modelos/:id", can("dispositivos.edit"), dispH.UpdateModel)
		dispRoutes.DELETE("/modelos/:id", can("dispositivos.edit"), dispH.DeleteModel)
		dispRoutes.GET("/:id", can("dispositivos.view"), dispH.Get)
		dispRoutes.PUT("/:id", can("dispositivos.edit"), dispH.Update)
		dispRoutes.PATCH("/:id", can("dispositivos.edit"), dispH.Update)
		dispRoutes.DELETE("/:id", can("dispositivos.delete"), dispH.Delete)
		dispRoutes.POST("/:id/entregar", can("dispositivos.edit"), dispH.Entregar)
	}

	// LMS routes. Every organization-scoped request validates active membership.
	learningRoutes := api.Group("", authMiddleware)
	{
		learningRoutes.GET("/organizations", learningH.ListOrganizations)
		learningRoutes.POST("/organizations", learningH.CreateOrganization)
		learningRoutes.PUT("/organizations/:orgID", learningH.UpdateOrganization)
		learningRoutes.DELETE("/organizations/:orgID", learningH.DeleteOrganization)
		learningRoutes.POST("/organizations/:orgID/restore", learningH.RestoreOrganization)
		learningRoutes.POST("/organizations/:orgID/members", learningH.AddMember)
		learningRoutes.GET("/organizations/:orgID/students", learningH.ListStudents)
		learningRoutes.POST("/organizations/:orgID/students", learningH.AddStudent)
		learningRoutes.DELETE("/organizations/:orgID/students/:studentID", learningH.RemoveStudent)
		learningRoutes.GET("/organizations/:orgID/courses", learningH.ListCourses)
		learningRoutes.POST("/organizations/:orgID/courses", learningH.CreateCourse)
		learningRoutes.PUT("/organizations/:orgID/courses/:courseID", learningH.UpdateCourse)
		learningRoutes.DELETE("/organizations/:orgID/courses/:courseID", learningH.DeleteCourse)
		learningRoutes.GET("/organizations/:orgID/courses/:courseID", learningH.GetCourse)
		learningRoutes.GET("/organizations/:orgID/courses/:courseID/progress", learningH.GetCourseProgress)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/lessons/:lessonID/complete", learningH.CompleteLesson)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/lessons/:lessonID/submissions", assessmentH.Submit)
		learningRoutes.GET("/organizations/:orgID/courses/:courseID/lessons/:lessonID/submissions", assessmentH.ListStudentSubmissions)
		learningRoutes.GET("/organizations/:orgID/courses/:courseID/submissions", assessmentH.ListCourseSubmissions)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/submissions/:submissionID/review", assessmentH.Review)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/publish", learningH.PublishCourse)
		learningRoutes.GET("/organizations/:orgID/cohorts", learningH.ListCohorts)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/cohorts", learningH.CreateCohort)
		learningRoutes.GET("/organizations/:orgID/courses/:courseID/cohorts", learningH.ListCohorts)
		learningRoutes.PUT("/organizations/:orgID/cohorts/:cohortID", learningH.UpdateCohort)
		learningRoutes.PUT("/organizations/:orgID/cohorts/:cohortID/status", learningH.TransitionCohort)
		learningRoutes.GET("/organizations/:orgID/cohorts/:cohortID/instructors", learningH.ListCohortInstructors)
		learningRoutes.POST("/organizations/:orgID/cohorts/:cohortID/instructors", learningH.AssignCohortInstructor)
		learningRoutes.DELETE("/organizations/:orgID/cohorts/:cohortID/instructors/:trainerID", learningH.RemoveCohortInstructor)
		learningRoutes.GET("/organizations/:orgID/cohorts/:cohortID/enrollments", learningH.ListCohortEnrollments)
		learningRoutes.POST("/organizations/:orgID/cohorts/:cohortID/enrollments", learningH.EnrollStudent)
		learningRoutes.DELETE("/organizations/:orgID/cohorts/:cohortID/enrollments/:studentID", learningH.WithdrawStudent)
		learningRoutes.GET("/organizations/:orgID/courses/:courseID/instructors", learningH.ListCourseTrainers)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/instructors", learningH.AssignCourseTrainer)
		learningRoutes.DELETE("/organizations/:orgID/courses/:courseID/instructors/:trainerID", learningH.RemoveCourseTrainer)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/modules", learningH.AddModule)
		learningRoutes.PUT("/organizations/:orgID/courses/:courseID/modules/order", learningH.ReorderModules)
		learningRoutes.PUT("/organizations/:orgID/courses/:courseID/modules/:moduleID", learningH.UpdateModule)
		learningRoutes.DELETE("/organizations/:orgID/courses/:courseID/modules/:moduleID", learningH.DeleteModule)
		learningRoutes.POST("/organizations/:orgID/courses/:courseID/modules/:moduleID/lessons", learningH.AddLesson)
		learningRoutes.PUT("/organizations/:orgID/courses/:courseID/modules/:moduleID/lessons/order", learningH.ReorderLessons)
		learningRoutes.PUT("/organizations/:orgID/courses/:courseID/modules/:moduleID/lessons/:lessonID", learningH.UpdateLesson)
		learningRoutes.DELETE("/organizations/:orgID/courses/:courseID/modules/:moduleID/lessons/:lessonID", learningH.DeleteLesson)
	}

	// Dashboard (any authenticated user)
	api.GET("/dashboard", authMiddleware, dashH.Get)

	// ── External API (Bearer token, read-only, scope-gated) ──────────────────
	extAuth := api.Group("/ext/auth")
	{
		extAuth.POST("/token", extTokenLimiter.Middleware(), extAuthH.Token)
		extAuth.POST("/refresh", extTokenLimiter.Middleware(), extAuthH.Refresh)
	}

	bearerAuth := middleware.BearerAuth(apiClientH.Svc())

	ext := api.Group("/ext", bearerAuth)
	{
		ext.GET("/trabajadores",
			middleware.RequireScope(domain.ScopeTrabajadoresRead),
			extH.ListTrabajadores,
		)
		ext.GET("/trabajadores/:id",
			middleware.RequireScope(domain.ScopeTrabajadoresRead),
			extH.GetTrabajador,
		)
		ext.GET("/telefonos",
			middleware.RequireScope(domain.ScopeTelefonosRead),
			extH.ListTelefonos,
		)
		ext.GET("/telefonos/:id",
			middleware.RequireScope(domain.ScopeTelefonosRead),
			extH.GetTelefono,
		)
		ext.GET("/dispositivos",
			middleware.RequireScope(domain.ScopeDispositivosRead),
			extH.ListDispositivos,
		)
		ext.GET("/dispositivos/:id",
			middleware.RequireScope(domain.ScopeDispositivosRead),
			extH.GetDispositivo,
		)
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
