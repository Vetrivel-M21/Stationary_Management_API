package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"stationery-management/internal/config"
	"stationery-management/internal/handler"
	"stationery-management/internal/middleware"
	"stationery-management/internal/repository"
	"stationery-management/internal/service"
	"stationery-management/pkg/email"
	"stationery-management/pkg/logger"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize file logger (logs/app.log) & terminal logger
	if err := logger.InitLogger("logs/app.log"); err != nil {
		log.Printf("Failed to initialize logger file: %v\n", err)
	}
	defer logger.Close()

	cfg := config.LoadConfig()

	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Printf("[WARNING] Database connection failed: %v. Running in memory / offline mode if DB not available.\n", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	branchRepo := repository.NewBranchRepository(db)
	productRepo := repository.NewProductRepository(db)
	reqRepo := repository.NewRequestRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	chatRepo := repository.NewChatRepository(db)
	slaRepo := repository.NewSlaRepository(db)

	// Utilities & Services
	emailSvc := email.NewEmailService(cfg)
	authSvc := service.NewAuthService(userRepo, auditRepo, cfg)
	userSvc := service.NewUserService(userRepo, auditRepo)
	branchSvc := service.NewBranchService(branchRepo, auditRepo)
	productSvc := service.NewProductService(productRepo, auditRepo)
	reqSvc := service.NewRequestService(reqRepo, userRepo, auditRepo)
	monitorSvc := service.NewMonitorService(reqRepo, userRepo, emailSvc, auditRepo)
	chatSvc := service.NewChatService(chatRepo, reqRepo, userRepo, auditRepo)
	slaSvc := service.NewSlaService(slaRepo, reqRepo, auditRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	branchHandler := handler.NewBranchHandler(branchSvc)
	productHandler := handler.NewProductHandler(productSvc)
	reqHandler := handler.NewRequestHandler(reqSvc, userSvc)
	monitorHandler := handler.NewMonitorHandler(monitorSvc)
	dashboardHandler := handler.NewDashboardHandler(reqRepo, auditRepo)
	chatHandler := handler.NewChatHandler(chatSvc)
	slaHandler := handler.NewSlaHandler(slaSvc)

	if cfg.Env == "production" || cfg.Env == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.PanicRecovery())

	// CORS Setup
	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Root & Health Check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"message": "Stationery Management System API is running",
			"health":  "/health",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "Stationery Management System API",
		})
	})

	// Public Routes
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// Protected Routes
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(cfg))
		{
			protected.GET("/auth/me", authHandler.GetMe)
			protected.POST("/auth/change-password", middleware.RequireRoles("ADMIN"), authHandler.ChangePassword)
			protected.POST("/auth/logout", authHandler.Logout)

			// Admin Routes
			admin := protected.Group("")
			admin.Use(middleware.RequireRoles("ADMIN"))
			{
				// User Management
				admin.GET("/users", userHandler.GetAllUsers)
				admin.POST("/users", userHandler.CreateUser)
				admin.PUT("/users/:id", userHandler.UpdateUser)
				admin.DELETE("/users/:id", userHandler.DeleteUser)
				admin.POST("/users/reset-password", userHandler.ResetPassword)

				// Branch Management (Admin write)
				admin.POST("/branches", branchHandler.CreateBranch)
				admin.PUT("/branches/:id", branchHandler.UpdateBranch)

				// Product Management (Admin write)
				admin.POST("/products", productHandler.CreateProduct)
				admin.PUT("/products/:id", productHandler.UpdateProduct)
				admin.DELETE("/products/:id", productHandler.DeleteProduct)

				// SLA Settings Management
				admin.PUT("/sla-settings", slaHandler.UpdateSlaSettings)

				// Audit Logs
				admin.GET("/dashboard/audit-logs", dashboardHandler.GetAuditLogs)
			}

			// Read-only catalog routes (Accessible by all authenticated users)
			protected.GET("/branches", branchHandler.GetAllBranches)
			protected.GET("/products", productHandler.GetAllProducts)
			protected.GET("/sla-settings", slaHandler.GetSlaSettings)

			// Request Workflow Routes
			// Branch Requester: Create Request
			protected.POST("/requests", middleware.RequireRoles("ADMIN", "BRANCH_REQUESTER"), reqHandler.CreateRequest)
			protected.GET("/requests", reqHandler.GetRequests)
			protected.GET("/requests/:id", reqHandler.GetRequestByID)

			// REST API Request Chat
			protected.GET("/requests/:id/chat", chatHandler.GetChatMessages)
			protected.POST("/requests/:id/chat", chatHandler.SendChatMessage)

			// Approver: Process Approval
			protected.POST("/requests/:id/approve", middleware.RequireRoles("ADMIN", "APPROVER"), reqHandler.ProcessApproval)

			// Agency: Process Delivery
			protected.POST("/requests/:id/deliver", middleware.RequireRoles("ADMIN", "AGENCY"), reqHandler.ProcessDelivery)

			// Requester: Process Verification
			protected.POST("/requests/:id/verify", middleware.RequireRoles("ADMIN", "BRANCH_REQUESTER"), reqHandler.ProcessVerification)

			// Monitor Routes
			protected.GET("/monitor/delayed-orders", middleware.RequireRoles("ADMIN", "MONITOR"), slaHandler.GetDelayedOrders)
			protected.POST("/monitor/remind", middleware.RequireRoles("ADMIN", "MONITOR"), monitorHandler.SendReminder)

			// Dashboard Metrics
			protected.GET("/dashboard/metrics", dashboardHandler.GetMetrics)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Stationery Management System Server starting on port %s...\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting successfully.")
}