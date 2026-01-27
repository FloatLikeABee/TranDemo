package db

import (
	"encoding/json"
	"fmt"
	"log"
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

// StoreComplaintState stores complaint flow state
func (d *DB) StoreComplaintState(userID string, state *models.ComplaintState) error {
	keyStr := fmt.Sprintf("complaint:%s:%s", userID, state.ConversationID)
	log.Printf("[DB] Storing complaint state - key: %s, conversationID: %s, step: %s, exchanges: %d", 
		keyStr, state.ConversationID, state.Step, state.ExchangeCount)
	
	err := d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(keyStr)
		
		data, err := json.Marshal(state)
		if err != nil {
			log.Printf("[DB] Error marshaling state: %v", err)
			return err
		}
		
		log.Printf("[DB] Setting key in transaction, data size: %d bytes", len(data))
		if err := txn.Set(key, data); err != nil {
			log.Printf("[DB] Error setting key in transaction: %v", err)
			return err
		}
		
		log.Printf("[DB] Successfully set key in transaction")
		return nil
	})
	
	if err != nil {
		log.Printf("[DB] Error in Update transaction: %v", err)
		return err
	}
	
	log.Printf("[DB] Transaction committed successfully for key: %s", keyStr)
	return nil
}

// GetComplaintState retrieves complaint flow state
func (d *DB) GetComplaintState(userID, conversationID string) (*models.ComplaintState, error) {
	var state *models.ComplaintState
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("complaint:%s:%s", userID, conversationID))
		
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		
		return item.Value(func(val []byte) error {
			state = &models.ComplaintState{}
			return json.Unmarshal(val, state)
		})
	})
	
	if err != nil {
		return nil, err
	}
	
	return state, nil
}

// GetComplaintStateByUserID gets the most recent complaint state for a user
func (d *DB) GetComplaintStateByUserID(userID string) (*models.ComplaintState, error) {
	var state *models.ComplaintState
	var found bool
	
	prefix := fmt.Sprintf("complaint:%s:", userID)
	log.Printf("[DB] Looking for complaint state with prefix: %s", prefix)
	
	// First, let's see ALL complaint keys for debugging
	d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("complaint:")
		it := txn.NewIterator(opts)
		defer it.Close()
		
		log.Printf("[DB] DEBUG: Scanning ALL complaint keys...")
		count := 0
		for it.Rewind(); it.Valid(); it.Next() {
			count++
			key := string(it.Item().Key())
			log.Printf("[DB] DEBUG: Found complaint key #%d: %s", count, key)
		}
		log.Printf("[DB] DEBUG: Total complaint keys found: %d", count)
		return nil
	})
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		// Don't use Reverse - just iterate forward and get the last ACTIVE one
		it := txn.NewIterator(opts)
		defer it.Close()
		
		log.Printf("[DB] Starting iterator with prefix: %s", prefix)
		
		// Iterate forward and collect all, then find the most recent ACTIVE (non-complete) one
		var lastActiveKey []byte
		var lastActiveItem *badger.Item
		var lastKey []byte
		var lastItem *badger.Item
		count := 0
		activeCount := 0
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			keyStr := string(key)
			
			// Verify prefix match
			if !strings.HasPrefix(keyStr, prefix) {
				log.Printf("[DB] Key '%s' doesn't match prefix '%s', stopping", keyStr, prefix)
				break
			}
			
			count++
			lastKey = key
			lastItem = item
			log.Printf("[DB] Iterator found key #%d: %s", count, keyStr)
			
			// Check if this state is active (not complete)
			err := item.Value(func(val []byte) error {
				var tempState models.ComplaintState
				if err := json.Unmarshal(val, &tempState); err != nil {
					return err
				}
				// If this state is not complete, it's an active session
				if tempState.Step != "complete" && tempState.ConversationID != "" {
					activeCount++
					lastActiveKey = key
					lastActiveItem = item
					log.Printf("[DB] Found active state #%d: conversationID: %s, step: %s", 
						activeCount, tempState.ConversationID, tempState.Step)
				}
				return nil
			})
			if err != nil {
				log.Printf("[DB] Error reading state value: %v", err)
			}
		}
		
		log.Printf("[DB] Iterator found %d total keys, %d active keys with prefix %s", count, activeCount, prefix)
		
		// Prefer active state over completed state
		if activeCount > 0 && lastActiveItem != nil {
			found = true
			return lastActiveItem.Value(func(val []byte) error {
				log.Printf("[DB] Reading value for last ACTIVE key: %s, size: %d bytes", string(lastActiveKey), len(val))
				state = &models.ComplaintState{}
				if err := json.Unmarshal(val, state); err != nil {
					log.Printf("[DB] Error unmarshaling complaint state: %v", err)
					return err
				}
				log.Printf("[DB] Successfully retrieved ACTIVE complaint state - conversationID: %s, step: %s, exchanges: %d", 
					state.ConversationID, state.Step, state.ExchangeCount)
				return nil
			})
		}
		
		// If no active state found, return the last one (even if complete) for reference
		if count > 0 && lastItem != nil {
			found = true
			return lastItem.Value(func(val []byte) error {
				log.Printf("[DB] Reading value for last key (no active found): %s, size: %d bytes", string(lastKey), len(val))
				state = &models.ComplaintState{}
				if err := json.Unmarshal(val, state); err != nil {
					log.Printf("[DB] Error unmarshaling complaint state: %v", err)
					return err
				}
				log.Printf("[DB] Successfully retrieved complaint state - conversationID: %s, step: %s, exchanges: %d", 
					state.ConversationID, state.Step, state.ExchangeCount)
				return nil
			})
		}
		
		return nil // Don't return error if not found, just set found = false
	})
	
	if err != nil {
		log.Printf("[DB] Error retrieving complaint state: %v", err)
		return nil, err
	}
	
	if !found {
		return nil, fmt.Errorf("no complaint state found")
	}
	
	return state, nil
}

// StoreVoiceProfile stores a voice profile for a user
func (d *DB) StoreVoiceProfile(profile *models.VoiceProfile) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("voice_profile:%s", profile.UserID))
		
		data, err := json.Marshal(profile)
		if err != nil {
			return err
		}
		
		return txn.Set(key, data)
	})
}

// GetVoiceProfile retrieves a voice profile by user ID
func (d *DB) GetVoiceProfile(userID string) (*models.VoiceProfile, error) {
	var profile *models.VoiceProfile
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("voice_profile:%s", userID))
		
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		
		return item.Value(func(val []byte) error {
			profile = &models.VoiceProfile{}
			return json.Unmarshal(val, profile)
		})
	})
	
	if err != nil {
		return nil, err
	}
	
	return profile, nil
}

// GetAllVoiceProfiles retrieves all voice profiles
func (d *DB) GetAllVoiceProfiles() ([]models.VoiceProfile, error) {
	var profiles []models.VoiceProfile
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("voice_profile:")
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var profile models.VoiceProfile
				if err := json.Unmarshal(val, &profile); err != nil {
					return err
				}
				profiles = append(profiles, profile)
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return profiles, nil
}

// DeleteVoiceProfile deletes a voice profile
func (d *DB) DeleteVoiceProfile(userID string) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("voice_profile:%s", userID))
		return txn.Delete(key)
	})
}

// Form Template CRUD operations

// StoreFormTemplate stores a form template
func (d *DB) StoreFormTemplate(template *models.FormTemplate) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("form_template:%s", template.ID))
		
		data, err := json.Marshal(template)
		if err != nil {
			return err
		}
		
		return txn.Set(key, data)
	})
}

// GetFormTemplate retrieves a form template by ID
func (d *DB) GetFormTemplate(id string) (*models.FormTemplate, error) {
	var template *models.FormTemplate
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("form_template:%s", id))
		
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		
		return item.Value(func(val []byte) error {
			template = &models.FormTemplate{}
			return json.Unmarshal(val, template)
		})
	})
	
	if err != nil {
		return nil, err
	}
	
	return template, nil
}

// GetAllFormTemplates retrieves all form templates
func (d *DB) GetAllFormTemplates() ([]models.FormTemplate, error) {
	var templates []models.FormTemplate
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("form_template:")
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var template models.FormTemplate
				if err := json.Unmarshal(val, &template); err != nil {
					return err
				}
				templates = append(templates, template)
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return templates, nil
}

// DeleteFormTemplate deletes a form template
func (d *DB) DeleteFormTemplate(id string) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("form_template:%s", id))
		return txn.Delete(key)
	})
}

// Form Answer CRUD operations

// StoreFormAnswer stores a form answer
func (d *DB) StoreFormAnswer(answer *models.FormAnswer) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("form_answer:%s", answer.ID))
		
		data, err := json.Marshal(answer)
		if err != nil {
			return err
		}
		
		return txn.Set(key, data)
	})
}

// GetFormAnswer retrieves a form answer by ID
func (d *DB) GetFormAnswer(id string) (*models.FormAnswer, error) {
	var answer *models.FormAnswer
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("form_answer:%s", id))
		
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		
		return item.Value(func(val []byte) error {
			answer = &models.FormAnswer{}
			return json.Unmarshal(val, answer)
		})
	})
	
	if err != nil {
		return nil, err
	}
	
	return answer, nil
}

// GetAllFormAnswers retrieves all form answers
func (d *DB) GetAllFormAnswers() ([]models.FormAnswer, error) {
	var answers []models.FormAnswer
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("form_answer:")
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var answer models.FormAnswer
				if err := json.Unmarshal(val, &answer); err != nil {
					return err
				}
				answers = append(answers, answer)
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return answers, nil
}

// GetFormAnswersByFormID retrieves all answers for a specific form
func (d *DB) GetFormAnswersByFormID(formID string) ([]models.FormAnswer, error) {
	var answers []models.FormAnswer
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("form_answer:")
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var answer models.FormAnswer
				if err := json.Unmarshal(val, &answer); err != nil {
					return err
				}
				if answer.FormID == formID {
					answers = append(answers, answer)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return answers, nil
}

// GetFormAnswersByUserID retrieves all answers submitted by a specific user
func (d *DB) GetFormAnswersByUserID(userID string) ([]models.FormAnswer, error) {
	var answers []models.FormAnswer
	
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("form_answer:")
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var answer models.FormAnswer
				if err := json.Unmarshal(val, &answer); err != nil {
					return err
				}
				if answer.UserID == userID {
					answers = append(answers, answer)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return answers, nil
}

// DeleteFormAnswer deletes a form answer
func (d *DB) DeleteFormAnswer(id string) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("form_answer:%s", id))
		return txn.Delete(key)
	})
}

// Registration flow state (one active session per user)

func (d *DB) StoreRegistrationState(userID string, state *models.RegistrationState) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("registration:%s", userID))
		data, err := json.Marshal(state)
		if err != nil {
			return err
		}
		return txn.Set(key, data)
	})
}

func (d *DB) GetRegistrationStateByUserID(userID string) (*models.RegistrationState, error) {
	var state *models.RegistrationState
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("registration:%s", userID))
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			state = &models.RegistrationState{}
			return json.Unmarshal(val, state)
		})
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (d *DB) DeleteRegistrationState(userID string) error {
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("registration:%s", userID))
		return txn.Delete(key)
	})
}

