package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"idongivaflyinfa/models"
)

type DB struct {
	badgerDB *badger.DB
}

func New(dbPath string) (*DB, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable badger logging for cleaner output

	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &DB{badgerDB: badgerDB}, nil
}

func (d *DB) Close() error {
	return d.badgerDB.Close()
}

func (d *DB) StoreSQLFile(name string, content string) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("sql_file:%s", name))
		return txn.Set(key, []byte(content))
	})
}

func (d *DB) GetSQLFiles() ([]models.SQLFile, error) {
	var sqlFiles []models.SQLFile

	err := d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("sql_file:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			name := strings.TrimPrefix(string(key), "sql_file:")

			err := item.Value(func(val []byte) error {
				sqlFiles = append(sqlFiles, models.SQLFile{
					Name:    name,
					Content: string(val),
				})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return sqlFiles, err
}

func (d *DB) StoreChatHistory(userID string, message string, response string) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		timestamp := time.Now().Unix()
		key := []byte(fmt.Sprintf("chat:%s:%d", userID, timestamp))

		history := models.ChatHistory{
			Message:   message,
			Response:  response,
			Timestamp: fmt.Sprintf("%d", timestamp),
		}

		data, err := json.Marshal(history)
		if err != nil {
			return err
		}

		return txn.Set(key, data)
	})
}

func (d *DB) LoadSQLFilesFromDir(sqlFilesDir string) ([]models.SQLFile, error) {
	var sqlFiles []models.SQLFile

	// Create directory if it doesn't exist
	if err := os.MkdirAll(sqlFilesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create SQL files directory: %w", err)
	}

	err := filepath.Walk(sqlFilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sql") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			sqlFiles = append(sqlFiles, models.SQLFile{
				Name:    info.Name(),
				Content: string(content),
			})
		}
		return nil
	})

	return sqlFiles, err
}

