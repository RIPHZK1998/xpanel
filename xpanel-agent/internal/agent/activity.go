package agent

import (
	"time"

	"xpanel-agent/internal/models"
)

// UserActivityState tracks the activity state of a single user.
type UserActivityState struct {
	Email        string
	LastSeen     time.Time
	LastUplink   int64 // Cumulative uplink bytes at last check
	LastDownlink int64 // Cumulative downlink bytes at last check
	IsActive     bool  // Had traffic since last check
}

// trackActivity detects active users by monitoring traffic changes.
// It compares current traffic stats with cached values to detect activity.
func (a *Agent) trackActivity() map[string]*UserActivityState {
	a.usersMutex.RLock()
	defer a.usersMutex.RUnlock()

	if len(a.currentUsers) == 0 {
		return nil
	}

	now := time.Now()
	activeUsers := make(map[string]*UserActivityState)

	for email := range a.currentUsers {
		stats, err := a.xrayAPI.GetUserStats(email)
		if err != nil {
			a.logger.Debugf("Failed to get stats for user %s: %v", email, err)
			continue
		}

		currentUplink := stats.UploadBytes
		currentDownlink := stats.DownloadBytes

		// Get or create activity state for this user
		state, exists := a.activityCache[email]
		if !exists {
			// First time seeing this user - initialize cache
			state = &UserActivityState{
				Email:        email,
				LastUplink:   currentUplink,
				LastDownlink: currentDownlink,
				IsActive:     false,
			}
			a.activityCache[email] = state
		}

		// Check if there's new traffic since last check
		uplinkDelta := currentUplink - state.LastUplink
		downlinkDelta := currentDownlink - state.LastDownlink

		if uplinkDelta > 0 || downlinkDelta > 0 {
			// User has new traffic - they're active
			state.IsActive = true
			state.LastSeen = now
			activeUsers[email] = state

			a.logger.Debugf("User %s is active (up: +%d, down: +%d)", email, uplinkDelta, downlinkDelta)
		} else {
			state.IsActive = false
		}

		// Update cached values
		state.LastUplink = currentUplink
		state.LastDownlink = currentDownlink
	}

	return activeUsers
}

// reportActivity sends activity data to the panel.
func (a *Agent) reportActivity() error {
	// Track activity and get active users
	activeUsers := a.trackActivity()

	// Build activity report for all users with recent activity
	activities := make([]models.UserActivityReport, 0)

	// Report all users in cache that have been seen
	for _, state := range a.activityCache {
		// Only report users that have been seen at some point
		if !state.LastSeen.IsZero() {
			report := models.UserActivityReport{
				Email:    state.Email,
				LastSeen: state.LastSeen,
				IsOnline: state.IsActive || time.Since(state.LastSeen) < 2*time.Minute,
			}

			// Get online IPs for this user using Xray's built-in API
			if report.IsOnline {
				ips, err := a.xrayAPI.GetOnlineIPs(state.Email)
				if err != nil {
					a.logger.Debugf("Failed to get online IPs for %s: %v", state.Email, err)
				} else if len(ips) > 0 {
					// Convert IP map to list
					ipList := make([]string, 0, len(ips))
					for ip := range ips {
						ipList = append(ipList, ip)
					}
					report.DeviceIPs = ipList
					report.DeviceCount = len(ipList)
				}
			}

			activities = append(activities, report)
		}
	}

	if len(activities) == 0 {
		return nil // Nothing to report
	}

	// Send report to panel
	report := &models.ActivityReportRequest{
		NodeID:    a.cfg.Node.ID,
		Users:     activities,
		Timestamp: time.Now(),
	}

	if err := a.panelClient.ReportActivity(report); err != nil {
		return err
	}

	if len(activeUsers) > 0 {
		a.logger.Infof("Activity report: %d active users, %d total tracked", len(activeUsers), len(activities))
	}

	return nil
}

// activityLoop periodically tracks and reports user activity.
func (a *Agent) activityLoop() {
	ticker := time.NewTicker(a.cfg.Intervals.ActivityReport)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if err := a.reportActivity(); err != nil {
				a.logger.Warnf("Failed to report activity: %v", err)
			}
		}
	}
}

// cleanupActivityCache removes entries for users no longer in currentUsers.
func (a *Agent) cleanupActivityCache() {
	a.usersMutex.RLock()
	defer a.usersMutex.RUnlock()

	for email := range a.activityCache {
		if _, exists := a.currentUsers[email]; !exists {
			delete(a.activityCache, email)
		}
	}
}
