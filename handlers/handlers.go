package handlers

import (
	"idongivaflyinfa/ai"
	"idongivaflyinfa/db"
	"idongivaflyinfa/service"
)

// @title           Transfinder Form/Report Assistant API
// @version         1.0
// @description     Transfinder Form/Report Assistant API - Generate SQL queries using AI and execute them against SQL Server
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9090
// @BasePath  /

// @schemes   http https

// Handlers contains all handler dependencies
type Handlers struct {
	db               *db.DB
	aiService        *ai.AIService
	sqlService       *service.SQLServerService
	complaintService *service.ComplaintService
	sqlFilesDir      string
}

// New creates a new Handlers instance
func New(db *db.DB, aiService *ai.AIService, sqlService *service.SQLServerService, sqlFilesDir string) *Handlers {
	return &Handlers{
		db:               db,
		aiService:        aiService,
		sqlService:       sqlService,
		complaintService: service.NewComplaintService(),
		sqlFilesDir:      sqlFilesDir,
	}
}
