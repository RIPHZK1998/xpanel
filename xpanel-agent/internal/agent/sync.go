package agent

import (
	"xpanel-agent/internal/models"
)

// syncUsers synchronizes users from the panel to xray-core.
func (a *Agent) syncUsers() error {
	// Fetch users from panel
	users, err := a.panelClient.FetchUserSync(a.cfg.Node.ID)
	if err != nil {
		return err
	}

	// Build map of panel users
	panelUsers := make(map[string]*models.UserConfig)
	for i := range users {
		panelUsers[users[i].Email] = &users[i]
	}

	a.usersMutex.Lock()
	defer a.usersMutex.Unlock()

	added := 0
	removed := 0

	// Add new users
	for email, user := range panelUsers {
		if _, exists := a.currentUsers[email]; !exists {
			// Add user to xray
			xrayUser := &models.XrayUser{
				Email:    user.Email,
				UUID:     user.UUID,
				Level:    0,
				Protocol: a.cfg.Xray.Protocol,
			}

			// Add flow for Reality
			if a.cfg.Xray.RealityEnabled {
				xrayUser.Flow = "xtls-rprx-vision"
			}

			if err := a.xrayAPI.AddUser(xrayUser); err != nil {
				a.logger.Errorf("Failed to add user %s: %v", email, err)
				continue
			}

			a.currentUsers[email] = user
			added++
		}
	}

	// Remove users that are no longer active
	for email := range a.currentUsers {
		if _, exists := panelUsers[email]; !exists {
			// Remove user from xray
			if err := a.xrayAPI.RemoveUser(email); err != nil {
				a.logger.Errorf("Failed to remove user %s: %v", email, err)
				continue
			}

			delete(a.currentUsers, email)
			removed++
		}
	}

	// Only log if there were changes, otherwise just show current count
	if added > 0 || removed > 0 {
		a.logger.Infof("Sync: %d active users (added: %d, removed: %d)", len(a.currentUsers), added, removed)
	} else {
		a.logger.Infof("Sync: %d active users", len(a.currentUsers))
	}

	return nil
}
