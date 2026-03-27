package main

import (
	"log"

	"attendancemgmt/backend/internal/config"
	"attendancemgmt/backend/internal/database"
	"attendancemgmt/backend/internal/handlers"
	"attendancemgmt/backend/internal/middleware"
	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"
	"attendancemgmt/backend/internal/service"

	_ "attendancemgmt/backend/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Attendance Management System API
// @version 1.0
// @description API for managing centers, departments, sewadars, and attendance.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	database.Seed(db, cfg.SuperAdminUsername, cfg.SuperAdminPassword)

	// Repositories
	centerRepo := repository.NewCenterRepository(db)
	deptRepo := repository.NewDepartmentRepository(db)
	sewadarRepo := repository.NewSewadarRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Services
	sewadarSvc := service.NewSewadarService(sewadarRepo, deptRepo)
	attendanceSvc := service.NewAttendanceService(attendanceRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	centerHandler := handlers.NewCenterHandler(centerRepo)
	deptHandler := handlers.NewDepartmentHandler(deptRepo)
	sewadarHandler := handlers.NewSewadarHandler(sewadarSvc)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceSvc, sewadarSvc)
	userHandler := handlers.NewUserHandler(userRepo)

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check route (public)
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(503, gin.H{"status": "error", "message": "database error", "error": err.Error()})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "error", "message": "database unreachable", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"status":   "ok",
			"database": "connected",
			"version":  "1.0.0",
		})
	})

	// Auth routes (public)
	r.POST("/auth/login", authHandler.Login)

	// Protected routes
	api := r.Group("/api", middleware.AuthMiddleware(cfg.JWTSecret))
	{
		api.GET("/me", authHandler.Me)

		// Dashboard
		api.GET("/dashboard", attendanceHandler.Dashboard)

		// Centers
		centers := api.Group("/centers")
		{
			centers.GET("", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin, models.RoleOperator), centerHandler.List)
			centers.POST("", middleware.RequireRole(models.RoleSuperAdmin), centerHandler.Create)
			centers.PUT("/:id", middleware.RequireRole(models.RoleSuperAdmin), centerHandler.Update)
			centers.DELETE("/:id", middleware.RequireRole(models.RoleSuperAdmin), centerHandler.Delete)
		}

		// Users
		users := api.Group("/users")
		{
			users.GET("", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), userHandler.List)
			users.POST("", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), userHandler.Create)
			users.PUT("/:id", middleware.RequireRole(models.RoleSuperAdmin), userHandler.Update)
			users.DELETE("/:id", middleware.RequireRole(models.RoleSuperAdmin), userHandler.Delete)
		}

		// Departments
		depts := api.Group("/departments")
		{
			depts.GET("", deptHandler.List)
			depts.GET("/:id", deptHandler.GetByID)
			depts.POST("", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), deptHandler.Create)
			depts.PUT("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), deptHandler.Update)
			depts.DELETE("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), deptHandler.Delete)
		}

		// Sewadars
		sewadars := api.Group("/sewadars")
		{
			sewadars.GET("", sewadarHandler.List)
			sewadars.GET("/export", sewadarHandler.Export)
			sewadars.GET("/:id", sewadarHandler.GetByID)
			sewadars.GET("/u/:uuid", sewadarHandler.GetByUUID)
			sewadars.POST("", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), sewadarHandler.Create)
			sewadars.PUT("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), sewadarHandler.Update)
			sewadars.DELETE("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), sewadarHandler.Delete)
			sewadars.POST("/transfer", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), sewadarHandler.Transfer)
			sewadars.POST("/bulk-upload", middleware.RequireRole(models.RoleSuperAdmin, models.RoleCenterAdmin), sewadarHandler.BulkUpload)
		}

		// Attendance
		attendance := api.Group("/attendance")
		{
			attendance.GET("", attendanceHandler.List)
			attendance.POST("/check-in", middleware.RequireRole(models.RoleCenterAdmin, models.RoleOperator), attendanceHandler.CheckIn)
			attendance.PUT("/:id/check-out", middleware.RequireRole(models.RoleCenterAdmin, models.RoleOperator), attendanceHandler.CheckOut)
			attendance.PUT("/:id", middleware.RequireRole(models.RoleCenterAdmin), attendanceHandler.Update)
			attendance.GET("/export", middleware.RequireRole(models.RoleCenterAdmin, models.RoleOperator), attendanceHandler.Export)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
