package agent

import (
	"time"

	"xpanel-agent/internal/models"
)

// reportTraffic collects and reports traffic statistics to the panel.
func (a *Agent) reportTraffic() error {
	a.usersMutex.RLock()
	defer a.usersMutex.RUnlock()

	if len(a.currentUsers) == 0 {
		return nil // No users to report
	}

	traffic := make([]models.UserTrafficReport, 0, len(a.currentUsers))

	// Collect stats for each user
	for email := range a.currentUsers {
		stats, err := a.xrayAPI.GetUserStats(email)
		if err != nil {
			a.logger.Warnf("Failed to get stats for user %s: %v", email, err)
			continue
		}

		// Only report if there's traffic
		if stats.UploadBytes > 0 || stats.DownloadBytes > 0 {
			traffic = append(traffic, models.UserTrafficReport{
				UserEmail:     email,
				UploadBytes:   stats.UploadBytes,
				DownloadBytes: stats.DownloadBytes,
			})

			// Reset stats after collecting
			if err := a.xrayAPI.ResetUserStats(email); err != nil {
				a.logger.Warnf("Failed to reset stats for user %s: %v", email, err)
			}
		}
	}

	if len(traffic) == 0 {
		return nil // No traffic to report
	}

	// Send report to panel
	report := &models.TrafficReportRequest{
		NodeID:    a.cfg.Node.ID,
		Traffic:   traffic,
		Timestamp: time.Now(),
	}

	if err := a.panelClient.ReportTraffic(report); err != nil {
		return err
	}

	a.logger.Debugf("Reported traffic for %d users", len(traffic))
	return nil
}
