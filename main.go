package main

import (
	"log"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/cache"
	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	_ "idongivaflyinfa/docs" // Swagger docs
	"idongivaflyinfa/handlers"
	"idongivaflyinfa/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfg := config.GetConfig()

	// Initialize database
	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize cache
	appCache := cache.New()

	// Initialize Gemini AI client
	aiService, err := ai.New(cfg.GeminiAPIKey, cfg.ModelName, appCache)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini: %v", err)
	}
	defer aiService.Close()

	// Initialize SQL Server service (optional)
	var sqlService *service.SQLServerService
	if cfg.SQLServer.Server != "" && cfg.SQLServer.Database != "" {
		sqlService, err = service.NewSQLServerService(cfg.SQLServer, cfg.ResultsDir, cfg.SitesDir)
		if err != nil {
			log.Printf("Warning: Failed to initialize SQL Server service: %v", err)
			log.Println("SQL Server features will be unavailable")
		} else {
			defer sqlService.Close()
			log.Println("SQL Server service initialized successfully")
		}
	}

	// Load existing SQL files from directory into DB
	sqlFiles, err := database.LoadSQLFilesFromDir(cfg.SQLFilesDir)
	if err == nil {
		for _, sqlFile := range sqlFiles {
			database.StoreSQLFile(sqlFile.Name, sqlFile.Content)
		}
		log.Printf("Loaded %d SQL files into database", len(sqlFiles))
	}

	// Initialize handlers
	h := handlers.New(database, aiService, sqlService, cfg.SQLFilesDir, cfg.VoiceSamplesDir)

	// Setup Gin router
	r := gin.Default()

	// CORS configuration - Allow ALL origins, headers, and methods
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Always allow the requesting origin (allows all origins dynamically)
		// If no origin header, use * (for non-browser requests)
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		}
		
		// Allow all headers and methods
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-User-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
			}
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-User-ID")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Routes
	r.GET("/health", h.HealthHandler)
	r.POST("/api/chat", h.ChatHandler)
	r.POST("/api/sql/upload", h.UploadSQLFileHandler)
	r.GET("/api/sql/files", h.ListSQLFilesHandler)
	r.POST("/api/sql/execute", h.ExecuteSQLHandler)
	
	// Result file routes
	r.GET("/api/results/files", h.ListResultFilesHandler)
	r.GET("/api/results/file/:filename", h.GetResultFileHandler)
	r.POST("/api/results/generate-html", h.GenerateHTMLHandler)
	r.GET("/api/results/html/:filename", h.ServeHTMLHandler)
	
	// Voice recognition routes
	r.POST("/api/voice/register", h.RegisterVoiceHandler)
	r.POST("/api/voice/recognize", h.RecognizeVoiceHandler)
	r.GET("/api/voice/profiles", h.ListVoiceProfilesHandler)
	r.DELETE("/api/voice/profile/:user_id", h.DeleteVoiceProfileHandler)

	// Serve static files (for React app)
	r.Static("/static", "./frontend/build/static")
	r.StaticFile("/", "./frontend/build/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/build/index.html")
	})

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
