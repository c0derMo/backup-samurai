package notifiers

import (
	"c0dermo/backup-samurai/config"
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

type NtfyNotifier struct {
	config config.SamuraiNtfyNotifier
	logger zerolog.Logger
}

func NewNtfyNotifier(config config.SamuraiNtfyNotifier, logger zerolog.Logger) *NtfyNotifier {
	handler := NtfyNotifier{
		config: config,
		logger: logger,
	}
	return &handler
}

func _formatSuccessMessage(template string, successfulSteps map[string]map[string]string, skippedSteps []string, ignoredFails []string) string {
	amount_successes := len(successfulSteps)
	amount_skips := len(skippedSteps)
	amount_ignored := len(ignoredFails)
	successful_steps := strings.Join(slices.Collect(maps.Keys(successfulSteps)), ", ")
	skipped_steps := strings.Join(skippedSteps, ", ")
	ignored_fails := strings.Join(ignoredFails, ", ")

	if template == "" {
		out := fmt.Sprintf("Samurai ran succesfully! %d steps were executed sucessfully: %s", amount_successes, successful_steps)
		if amount_skips > 0 {
			out += fmt.Sprintf(", %d steps were skipped: %s", amount_skips, skipped_steps)
		}
		if amount_ignored > 0 {
			out += fmt.Sprintf(", %d errors were ignored: %s", amount_ignored, ignored_fails)
		}
		return out
	} else {
		var replaceMap []string
		for step, met := range successfulSteps {
			for k, v := range met {
				replaceMap = append(replaceMap, fmt.Sprintf("{%s.%s}", step, k), v)
			}
		}
		replaceMap = append(replaceMap,
			"{amount_successes}", strconv.Itoa(amount_successes),
			"{successes}", successful_steps,
			"{amount_skips}", strconv.Itoa(amount_skips),
			"{skips}", skipped_steps,
			"{amount_ignored}", strconv.Itoa(amount_ignored),
			"{ignored_fails}", ignored_fails,
		)
		replacer := strings.NewReplacer(replaceMap...)

		return replacer.Replace(template)
	}
}

func _formatFailureMessage(template string, successfulSteps map[string]map[string]string, failedSteps []string, skippedSteps []string, ignoredFails []string) string {
	amount_successes := len(successfulSteps)
	amount_fails := len(failedSteps)
	amount_skips := len(skippedSteps)
	amount_ignored := len(ignoredFails)
	successful_steps := strings.Join(slices.Collect(maps.Keys(successfulSteps)), ", ")
	failed_steps := strings.Join(failedSteps, ", ")
	skipped_steps := strings.Join(skippedSteps, ", ")
	ignored_fails := strings.Join(ignoredFails, ", ")

	if template == "" {
		out := fmt.Sprintf("Samurai ran with errors! %d steps were executed sucessfully: %s, but %d steps failed: %s", amount_successes, successful_steps, amount_fails, failed_steps)
		if amount_skips > 0 {
			out += fmt.Sprintf(", %d steps were skipped: %s", amount_skips, skipped_steps)
		}
		if amount_ignored > 0 {
			out += fmt.Sprintf(", %d errors were ignored: %s", amount_ignored, ignored_fails)
		}
		return out
	} else {
		var replaceMap []string
		for step, met := range successfulSteps {
			for k, v := range met {
				replaceMap = append(replaceMap, fmt.Sprintf("{%s.%s}", step, k), v)
			}
		}
		replaceMap = append(replaceMap,
			"{amount_successes}", strconv.Itoa(amount_successes),
			"{successes}", successful_steps,
			"{amount_fails}", strconv.Itoa(amount_fails),
			"{fails}", failed_steps,
			"{amount_skips}", strconv.Itoa(amount_skips),
			"{skips}", skipped_steps,
			"{amount_ignored}", strconv.Itoa(amount_ignored),
			"{ignored_fails}", ignored_fails,
		)
		replacer := strings.NewReplacer(replaceMap...)

		return replacer.Replace(template)
	}
}

func _formatUpdateMessage(template string, currentVersion string, newerVersion string) string {
	if template == "" {
		return fmt.Sprintf("A newer version of backup-samurai is available! Current: %s, latest: %s")
	}

	var replaceMap []string
	replaceMap = append(replaceMap,
		"{current_version}", currentVersion,
		"{latest_version}", newerVersion,
	)

	replacer := strings.NewReplacer(replaceMap...)

	return replacer.Replace(template)
}

func (notifier NtfyNotifier) _formatAuth() string {
	if notifier.config.Authorization.Username != "" && notifier.config.Authorization.Password != "" {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(notifier.config.Authorization.Username+":"+notifier.config.Authorization.Password))
	}
	if notifier.config.Authorization.Token != "" {
		return notifier.config.Authorization.Token
	}
	return ""
}

func (notifier NtfyNotifier) SendSuccess(successfulSteps map[string]map[string]string, skippedSteps []string, ignoredFails []string) error {
	full_uri := notifier.config.Server + "/" + notifier.config.Topic
	req, err := http.NewRequest("POST", full_uri, strings.NewReader(_formatSuccessMessage(notifier.config.Success.Message, successfulSteps, skippedSteps, ignoredFails)))
	if err != nil {
		return err
	}
	if notifier.config.Success.Title != "" {
		req.Header.Set("Title", notifier.config.Success.Title)
	}
	if notifier.config.Success.Priority != "" {
		req.Header.Set("Priority", notifier.config.Success.Priority)
	}
	if notifier.config.Success.Tags != "" {
		req.Header.Set("Tags", notifier.config.Success.Tags)
	}
	if notifier._formatAuth() != "" {
		req.Header.Set("Authorization", notifier._formatAuth())
	}

	notifier.logger.Debug().Str("uri", full_uri).Msg("Sending success request")
	_, err = http.DefaultClient.Do(req)

	return err
}

func (notifier NtfyNotifier) SendFailure(successfulSteps map[string]map[string]string, failedSteps []string, skippedSteps []string, ignoredFails []string) error {
	full_uri := notifier.config.Server + "/" + notifier.config.Topic
	req, err := http.NewRequest("POST", full_uri, strings.NewReader(_formatFailureMessage(notifier.config.Failure.Message, successfulSteps, failedSteps, skippedSteps, ignoredFails)))
	if err != nil {
		return err
	}
	if notifier.config.Failure.Title != "" {
		req.Header.Set("Title", notifier.config.Failure.Title)
	}
	if notifier.config.Failure.Priority != "" {
		req.Header.Set("Priority", notifier.config.Failure.Priority)
	}
	if notifier.config.Failure.Tags != "" {
		req.Header.Set("Tags", notifier.config.Failure.Tags)
	}
	if notifier._formatAuth() != "" {
		req.Header.Set("Authorization", notifier._formatAuth())
	}

	notifier.logger.Debug().Str("uri", full_uri).Msg("Sending fail request")
	_, err = http.DefaultClient.Do(req)

	return err
}

func (notifier NtfyNotifier) SendUpdateAvailable(currentVersion string, newerVersion string) error {
	full_uri := notifier.config.Server + "/" + notifier.config.Topic
	req, err := http.NewRequest("POST", full_uri, strings.NewReader(_formatUpdateMessage(notifier.config.Update.Message, currentVersion, newerVersion)))
	if err != nil {
		return err
	}
	if notifier.config.Failure.Title != "" {
		req.Header.Set("Title", notifier.config.Failure.Title)
	}
	if notifier.config.Failure.Priority != "" {
		req.Header.Set("Priority", notifier.config.Failure.Priority)
	}
	if notifier.config.Failure.Tags != "" {
		req.Header.Set("Tags", notifier.config.Failure.Tags)
	}
	if notifier._formatAuth() != "" {
		req.Header.Set("Authorization", notifier._formatAuth())
	}

	notifier.logger.Debug().Str("uri", full_uri).Msg("Sending update request")
	_, err = http.DefaultClient.Do(req)

	return err
}

func (notifier NtfyNotifier) ValidateConfig() error {
	if notifier.config.Server == "" {
		return errors.New("ntfy server is required")
	}
	if notifier.config.Topic == "" {
		return errors.New("ntfy topic is required")
	}
	if (notifier.config.Authorization.Password != "" || notifier.config.Authorization.Username != "") && notifier.config.Authorization.Token != "" {
		return errors.New("ntfy token and username/password must not be used together")
	}
	if notifier.config.Authorization.Password != "" && notifier.config.Authorization.Username == "" {
		return errors.New("ntfy username must be set if password is used")
	}
	if notifier.config.Authorization.Password == "" && notifier.config.Authorization.Username != "" {
		return errors.New("ntfy password must be set if username is used")
	}
	return nil
}
