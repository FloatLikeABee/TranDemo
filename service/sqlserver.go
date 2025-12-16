package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/models"

	_ "github.com/microsoft/go-mssqldb"
)

type SQLServerService struct {
	db            *sql.DB
	resultsStorage *ResultsStorage
}

func NewSQLServerService(cfg config.SQLServerConfig, resultsDir string, sitesDir string) (*SQLServerService, error) {
	if cfg.Server == "" || cfg.Database == "" {
		return nil, fmt.Errorf("SQL Server configuration is incomplete")
	}

	connectionString := buildConnectionString(cfg)

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL Server connection: %w", err)
	}

	// Test the connection
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		// Log a warning but do not fail service initialization.
		// This allows the application to start even if SQL Server is temporarily unavailable.
		log.Printf("Warning: failed to ping SQL Server during initialization: %v", err)
	}

	// Initialize results storage
	resultsStorage, err := NewResultsStorage(resultsDir, sitesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize results storage: %w", err)
	}

	return &SQLServerService{
		db:            db,
		resultsStorage: resultsStorage,
	}, nil
}

func buildConnectionString(cfg config.SQLServerConfig) string {
	connStr := fmt.Sprintf("server=%s;port=%s;database=%s",
		cfg.Server, cfg.Port, cfg.Database)

	if cfg.UserID != "" {
		connStr += fmt.Sprintf(";user id=%s;password=%s", cfg.UserID, cfg.Password)
	} else {
		connStr += ";trusted_connection=true"
	}

	if cfg.Encrypt {
		// Use TLS but skip CA verification so self-signed / internal certs work.
		// NOTE: For production, you should configure proper certificates instead.
		connStr += ";encrypt=true;TrustServerCertificate=true"
	} else {
		connStr += ";encrypt=false"
	}

	return connStr
}

func (s *SQLServerService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLServerService) ExecuteQuery(query string) (*models.SQLResult, error) {
	return s.ExecuteQueryWithSave(query, "", false)
}

func (s *SQLServerService) ExecuteQueryWithSave(query string, format string, save bool) (*models.SQLResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("SQL Server connection is not initialized")
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return &models.SQLResult{
			Error: err.Error(),
		}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return &models.SQLResult{
			Error: err.Error(),
		}, err
	}

	var resultRows [][]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &models.SQLResult{
				Error: err.Error(),
			}, err
		}

		// Convert []interface{} to proper types
		row := make([]interface{}, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = nil
			} else {
				// Convert to string for JSON serialization
				row[i] = fmt.Sprintf("%v", val)
			}
		}

		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return &models.SQLResult{
			Error: err.Error(),
		}, err
	}

	result := &models.SQLResult{
		Columns: columns,
		Rows:    resultRows,
	}

	// Save result if requested
	if save && s.resultsStorage != nil {
		result.Filename = ""
		if format == "csv" {
			filename, err := s.resultsStorage.SaveResultAsCSV(result, query)
			if err == nil {
				result.Filename = filename
			}
		} else {
			// Default to JSON
			filename, err := s.resultsStorage.SaveResultAsJSON(result, query)
			if err == nil {
				result.Filename = filename
			}
		}
	}

	return result, nil
}

func (s *SQLServerService) GetResultsStorage() *ResultsStorage {
	return s.resultsStorage
}

func (s *SQLServerService) ExecuteNonQuery(query string) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("SQL Server connection is not initialized")
	}

	result, err := s.db.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *SQLServerService) IsConnected() bool {
	if s.db == nil {
		return false
	}
	return s.db.Ping() == nil
}

