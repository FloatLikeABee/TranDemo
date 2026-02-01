package handlers

import (
	"sync"

	"idongivaflyinfa/models"
)

var (
	pendingFormMu   sync.RWMutex
	pendingFormByUser = make(map[string]*models.FormTemplate)
)

func getPendingForm(userID string) *models.FormTemplate {
	pendingFormMu.RLock()
	defer pendingFormMu.RUnlock()
	return pendingFormByUser[userID]
}

func setPendingForm(userID string, t *models.FormTemplate) {
	pendingFormMu.Lock()
	defer pendingFormMu.Unlock()
	if t == nil {
		delete(pendingFormByUser, userID)
		return
	}
	pendingFormByUser[userID] = t
}

func clearPendingForm(userID string) {
	setPendingForm(userID, nil)
}
