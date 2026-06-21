package handlers

import (
	"c0dermo/backup-samurai/config"
	"errors"
	"os/exec"

	"github.com/rs/zerolog"
)

type ScriptHandler struct {
	config config.SamuraiScriptHandler
	logger zerolog.Logger
}

func NewScriptHandler(config config.SamuraiScriptHandler, logger zerolog.Logger) *ScriptHandler {
	handler := ScriptHandler{
		config: config,
		logger: logger,
	}
	return &handler
}

func (handler ScriptHandler) Execute() (error, map[string]string) {
	handler.logger.Debug().Msg("Starting execution")

	command := exec.Command(handler.config.Binary, handler.config.Options...)
	handler.logger.Debug().Strs("cmd", command.Args).Msgf("Executing custom command")
	output, err := command.CombinedOutput()
	if err != nil {
		return FormatExitError(err, "failed to run script", output), make(map[string]string, 0)
	}

	return nil, make(map[string]string, 0)
}

func (handler ScriptHandler) ValidateConfig() error {
	if handler.config.Binary == "" {
		return errors.New("Binary must be set (e.g. /bin/bash)")
	}
	return nil
}
