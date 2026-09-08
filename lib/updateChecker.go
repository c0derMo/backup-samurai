package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	Version        = "dev"
	CommitHash     = "n/a"
	BuildTimestamp = "n/a"

	releaseApi = "https://api.github.com/repos/c0derMo/backup-samurai/releases/latest"
)

func BuildVersion() string {
	return fmt.Sprintf("backup-samurai %s (commit %s, built %s)", Version, CommitHash, BuildTimestamp)
}

type GithubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
}

func CheckForUpdate() (string, error) {
	logger := log.With().Str("name", "updatechecker").Logger()

	resp, err := http.Get(releaseApi)
	if err != nil {
		return "", fmt.Errorf("an error occured while checking for updates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("an error occured while checking for updates: %w", err)
	}

	var release GithubRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", fmt.Errorf("an error occured while parsing latest github release: %w", err)
	}

	logger.Debug().Str("local_buildtime", BuildTimestamp).Str("latest_release", release.PublishedAt).Msgf("Raw timestamps")

	releaseDT, err := time.Parse(time.RFC3339, release.PublishedAt)
	if err != nil {
		return "", fmt.Errorf("an error occured while parsing newest release date: %w", err)
	}
	localDT, err := time.Parse(time.RFC3339, BuildTimestamp)
	if err != nil {
		return "", fmt.Errorf("an error occured while parsing local release date: %w", err)
	}

	logger.Debug().Time("local_buildtime", localDT).Time("latest_release", releaseDT).Msgf("Comparing build time & release time")

	if releaseDT.After(localDT) {
		if release.Name == release.TagName {
			return release.TagName, nil
		} else {
			return fmt.Sprintf("%s (%s)", release.TagName, release.Name), nil
		}
	} else {
		return "", nil
	}
}
