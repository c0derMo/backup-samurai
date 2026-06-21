package handlers

import (
	"c0dermo/backup-samurai/config"
	"errors"
	"os"
	"os/exec"

	"github.com/rs/zerolog"
)

type PgBackupHandler struct {
	config config.SamuraiPgbackupHandler
	logger zerolog.Logger
}

func NewPgBackupHandler(config config.SamuraiPgbackupHandler, logger zerolog.Logger) *PgBackupHandler {
	handler := PgBackupHandler{
		config: config,
		logger: logger,
	}
	return &handler
}

func (handler PgBackupHandler) Execute() (error, map[string]string) {
	handler.logger.Debug().Msg("Starting execution")

	var command_options []string
	binary := "pg_basebackup"
	if handler.config.Binary != "" {
		binary = handler.config.Binary
	}

	if handler.config.Docker.RunInContainer {
		command_options = append(command_options, "exec", handler.config.Docker.Container, binary)
		binary = "docker"
	}

	command_options = append(command_options, "-Ft", "-Xf")

	if handler.config.Gzip {
		command_options = append(command_options, "-z")
	}
	if handler.config.DatabaseUser != "" {
		command_options = append(command_options, "-U", handler.config.DatabaseUser)
	}

	if handler.config.Docker.RunInContainer {
		command_options = append(command_options, "-D", "-")
	} else {
		command_options = append(command_options, "-D", handler.config.Destination)
	}

	backup_command := exec.Command(binary, command_options...)
	handler.logger.Debug().Strs("cmd", backup_command.Args).Msgf("Executing backup command")
	output, err := backup_command.CombinedOutput()
	if err != nil {
		return FormatExitError(err, "failed to run pg_basebackup", output), make(map[string]string, 0)
	}

	if handler.config.Docker.RunInContainer {
		f, err := os.Create(handler.config.Destination)
		if err != nil {
			handler.logger.Err(err).Msg("Could not write backup to file!")
			return err, make(map[string]string, 0)
		}
		defer f.Close()
		f.Write(output)
	}

	return nil, make(map[string]string, 0)
}

func (handler PgBackupHandler) ValidateConfig() error {
	if handler.config.Destination == "" {
		return errors.New("backup destination must be set")
	}
	if handler.config.Docker.RunInContainer && handler.config.Docker.Container == "" {
		return errors.New("container must be set if backup is running in container")
	}
	return nil
}
