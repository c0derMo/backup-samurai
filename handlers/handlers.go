package handlers

import (
	"c0dermo/backup-samurai/config"
	"strings"

	"github.com/rs/zerolog/log"
)

type Handler interface {
	Execute() (error, map[string]string)
	ValidateConfig() error
}

func GetConfigByKey(cfg *config.SamuraiHandler, key string) (Handler, config.SamuraiHandlerBase) {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return nil, config.SamuraiHandlerBase{}
	}
	if parts[0] != "handler" {
		return nil, config.SamuraiHandlerBase{}
	}
	handler_type := strings.ToLower(parts[1])
	handler_key := parts[2]

	var handler Handler
	var baseCfg config.SamuraiHandlerBase
	logger := log.With().Str("handler", handler_type).Str("name", handler_key).Logger()

	switch handler_type {
	case "restic":
		baseCfg = cfg.Restic[handler_key].SamuraiHandlerBase
		handler = NewResticHandler(cfg.Restic[handler_key], logger)
	case "pgbackup":
		baseCfg = cfg.Pgbackup[handler_key].SamuraiHandlerBase
		handler = NewPgBackupHandler(cfg.Pgbackup[handler_key], logger)
	case "rsync":
		baseCfg = cfg.Rsync[handler_key].SamuraiHandlerBase
		handler = NewRsyncHandler(cfg.Rsync[handler_key], logger)
	case "rclone":
		baseCfg = cfg.Rclone[handler_key].SamuraiHandlerBase
		handler = NewRcloneHandler(cfg.Rclone[handler_key], logger)
	case "script":
		baseCfg = cfg.Script[handler_key].SamuraiHandlerBase
		handler = NewScriptHandler(cfg.Script[handler_key], logger)
	default:
		return nil, config.SamuraiHandlerBase{}
	}

	return handler, baseCfg
}
