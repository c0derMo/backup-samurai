package handlers

import (
	"c0dermo/backup-samurai/config"
	"errors"
	"os/exec"

	"github.com/rs/zerolog"
)

type RcloneHandler struct {
	config config.SamuraiRcloneHandler
	logger zerolog.Logger
}

func NewRcloneHandler(config config.SamuraiRcloneHandler, logger zerolog.Logger) *RcloneHandler {
	handler := RcloneHandler{
		config: config,
		logger: logger,
	}
	return &handler
}

func buildRcloneLocation(location config.SamuraiRcloneHandlerLocation) string {
	location_string := ""
	if location.Backend != "" {
		location_string += ":" + location.Backend + ":"
	}
	location_string += location.Path
	return location_string
}

func (handler RcloneHandler) Execute() (error, map[string]string) {
	handler.logger.Debug().Msg("Starting execution")

	binary := "rclone"
	if handler.config.Binary != "" {
		binary = handler.config.Binary
	}

	var command_options []string
	command_options = append(command_options, "sync")
	command_options = append(command_options, handler.config.Source.Flags...)
	command_options = append(command_options, buildRcloneLocation(handler.config.Source))
	command_options = append(command_options, handler.config.Destination.Flags...)
	command_options = append(command_options, buildRcloneLocation(handler.config.Destination))

	command := exec.Command(binary, command_options...)
	handler.logger.Debug().Strs("cmd", command.Args).Msgf("Executing sync command")
	output, err := command.CombinedOutput()
	if err != nil {
		return FormatExitError(err, "failed to run rclone", output), make(map[string]string, 0)
	}

	return nil, make(map[string]string, 0)
}

func (handler RcloneHandler) ValidateConfig() error {
	if handler.config.Source.Path == "" {
		return errors.New("source path must be set")
	}
	if handler.config.Destination.Path == "" {
		return errors.New("destination must be set")
	}
	return nil
}
