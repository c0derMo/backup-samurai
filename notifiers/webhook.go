package notifiers

import (
	"bytes"
	"c0dermo/backup-samurai/config"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/rs/zerolog"
)

type WebhookNotifier struct {
	config config.SamuraiWebhookNotifier
	logger zerolog.Logger
}

func NewWebhookNotifier(config config.SamuraiWebhookNotifier, logger zerolog.Logger) *WebhookNotifier {
	handler := WebhookNotifier{
		config: config,
		logger: logger,
	}
	return &handler
}

type webhookData struct {
	Metadata        map[string]string
	SuccessfulSteps map[string]map[string]string
	FailedSteps     []string
	SkippedSteps    []string
	IgnoredFails    []string
}

func (notifier WebhookNotifier) _sendHook(url string, method string, metadata map[string]string, successfulSteps map[string]map[string]string, failedSteps []string, skippedSteps []string, ignoredFails []string) error {
	data := webhookData{
		Metadata:        metadata,
		SuccessfulSteps: successfulSteps,
		FailedSteps:     failedSteps,
		SkippedSteps:    skippedSteps,
		IgnoredFails:    ignoredFails,
	}

	if len(metadata) == 0 {
		data.Metadata = make(map[string]string, 0)
	}
	if len(successfulSteps) == 0 {
		data.SuccessfulSteps = make(map[string]map[string]string, 0)
	}
	if len(failedSteps) == 0 {
		data.FailedSteps = make([]string, 0)
	}
	if len(skippedSteps) == 0 {
		data.SkippedSteps = make([]string, 0)
	}
	if len(ignoredFails) == 0 {
		data.IgnoredFails = make([]string, 0)
	}

	s_data, err := json.Marshal(data)
	if err != nil {
		return err
	}

	notifier.logger.Debug().Msg(string(s_data))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(s_data))
	req.Header.Set("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	return err
}

func (notifier WebhookNotifier) SendSuccess(successfulSteps map[string]map[string]string, skippedSteps []string, ignoredFails []string) error {
	err := notifier._sendHook(
		notifier.config.URL,
		notifier.config.Method,
		notifier.config.Meta,
		successfulSteps,
		make([]string, 0),
		skippedSteps,
		ignoredFails,
	)
	if err == nil {
		notifier.logger.Info().Str("url", notifier.config.URL).Str("method", notifier.config.Method).Msg("Webhook request sent")
	}
	return err
}

func (notifier WebhookNotifier) SendFailure(successfulSteps map[string]map[string]string, failedSteps []string, skippedSteps []string, ignoredFails []string) error {
	err := notifier._sendHook(
		notifier.config.URL,
		notifier.config.Method,
		notifier.config.Meta,
		successfulSteps,
		failedSteps,
		skippedSteps,
		ignoredFails,
	)
	if err == nil {
		notifier.logger.Info().Str("url", notifier.config.URL).Str("method", notifier.config.Method).Msg("Webhook request sent")
	}
	return err
}

type webhookUpdateData struct {
	Metadata       map[string]string
	CurrentVersion string
	LatestVersion  string
}

func (notifier WebhookNotifier) SendUpdateAvailable(currentVersion string, newerVersion string) error {
	data := webhookUpdateData{
		Metadata:       notifier.config.Meta,
		CurrentVersion: currentVersion,
		LatestVersion:  newerVersion,
	}

	if len(notifier.config.Meta) == 0 {
		data.Metadata = make(map[string]string, 0)
	}

	s_data, err := json.Marshal(data)
	if err != nil {
		return err
	}

	notifier.logger.Debug().Msg(string(s_data))

	req, err := http.NewRequest(notifier.config.Method, notifier.config.URL, bytes.NewBuffer(s_data))
	req.Header.Set("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)

	if err == nil {
		notifier.logger.Info().Str("url", notifier.config.URL).Str("method", notifier.config.Method).Msg("Webhook request sent")
	}

	return err
}

func (notifier WebhookNotifier) ValidateConfig() error {
	if notifier.config.URL == "" {
		return errors.New("webhook url is required")
	}
	if !slices.Contains([]string{"POST", "PUT", "DELETE", "PATCH"}, notifier.config.Method) {
		return errors.New("method must be one of: POST, PUT, DELETE, PATCH")
	}
	return nil
}
