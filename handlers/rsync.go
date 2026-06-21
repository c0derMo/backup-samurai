package handlers

import (
	"c0dermo/backup-samurai/config"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog"
)

type RsyncHandler struct {
	config config.SamuraiRsyncHandler
	logger zerolog.Logger
}

func NewRsyncHandler(config config.SamuraiRsyncHandler, logger zerolog.Logger) *RsyncHandler {
	handler := RsyncHandler{
		config: config,
		logger: logger,
	}
	return &handler
}

func buildRsyncLocation(location config.SamuraiRsyncHandlerLocation) string {
	location_string := ""
	if location.Remote {
		if location.Protocol != "" {
			location_string += location.Protocol + "://"
		}
		if location.User != "" {
			location_string += location.User + "@"
		}
		location_string += location.Host + ":" + location.Location
	} else {
		location_string += location.Location
	}
	return location_string
}

func (handler RsyncHandler) Execute() (error, map[string]string) {
	handler.logger.Debug().Msg("Starting execution")

	binary := "rsync"
	if handler.config.Binary != "" {
		binary = handler.config.Binary
	}

	var command_options []string
	command_options = append(command_options, "-a")
	command_options = append(command_options, buildRsyncLocation(handler.config.Source))
	command_options = append(command_options, buildRsyncLocation(handler.config.Destination))

	command := exec.Command(binary, command_options...)
	handler.logger.Debug().Strs("cmd", command.Args).Msgf("Executing sync command")
	output, err := command.CombinedOutput()
	if err != nil {
		return FormatExitError(err, "failed to run rsync", output), make(map[string]string, 0)
	}

	return nil, make(map[string]string, 0)
}

func validateLocation(location config.SamuraiRsyncHandlerLocation, name string) error {
	if location.Location == "" {
		return fmt.Errorf("location for %s must be set", name)
	}
	if location.Remote && location.Host == "" {
		return fmt.Errorf("host for %s must be set if location is remote", name)
	}
	return nil
}

func (handler RsyncHandler) ValidateConfig() error {
	err := validateLocation(handler.config.Source, "source")
	if err != nil {
		return err
	}
	err = validateLocation(handler.config.Destination, "destination")
	return err
}
