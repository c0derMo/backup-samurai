package notifiers

import (
	"c0dermo/backup-samurai/config"
	"strings"

	"github.com/rs/zerolog/log"
)

type Notifier interface {
	SendSuccess(successfulSteps map[string]map[string]string, skippedSteps []string, ignoredFails []string) error
	SendFailure(successfulSteps map[string]map[string]string, failedSteps []string, skippedSteps []string, ignoredFails []string) error
	SendUpdateAvailable(currentVersion string, newerVersion string) error
	ValidateConfig() error
}

func GetConfigByKey(cfg *config.SamuraiNotifier, key string, category string) Notifier {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return nil
	}
	if parts[0] != category {
		return nil
	}
	notifier_type := strings.ToLower(parts[1])
	notifier_key := parts[2]

	var notifier Notifier
	logger := log.With().Str(category, notifier_type).Str("name", notifier_key).Logger()

	switch notifier_type {
	case "ntfy":
		notifier = NewNtfyNotifier(cfg.Ntfy[notifier_key], logger)
	case "webhook":
		notifier = NewWebhookNotifier(cfg.Webhook[notifier_key], logger)
	default:
		return nil
	}

	return notifier
}
