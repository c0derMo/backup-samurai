package main

import (
	"c0dermo/backup-samurai/config"
	"c0dermo/backup-samurai/handlers"
	"c0dermo/backup-samurai/lib"
	"c0dermo/backup-samurai/notifiers"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

type CliFlags struct {
	Verbose        bool
	ForceExecution bool
	ConfigFile     string
	LogTarget      string
}

func buildFlags() CliFlags {
	flags := CliFlags{
		Verbose:        false,
		ForceExecution: false,
		ConfigFile:     "./config.toml",
		LogTarget:      "-",
	}

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Backup Samurai - backup orchestration tool")
		fmt.Fprintln(os.Stderr, "Command line flags:")
		flag.PrintDefaults()
	}

	flag.BoolVarP(&flags.Verbose, "verbose", "v", false, "enable verbose logging")
	flag.BoolVarP(&flags.ForceExecution, "force-execution", "f", false, "force execution, even if cron schedule is too recent")
	flag.StringVarP(&flags.ConfigFile, "config", "c", "./config.toml", "path to config file")
	flag.StringVarP(&flags.LogTarget, "log-file", "l", "-", "file to log to. use - to log to stdout")

	flag.Parse()

	return flags
}

func main() {
	flags := buildFlags()
	samuraiLogging := lib.InitializeLogging(flags.Verbose, flags.LogTarget)
	defer samuraiLogging.Cleanup()

	var tomlConf config.SamuraiConfig
	meta, err := toml.DecodeFile(flags.ConfigFile, &tomlConf)
	if err != nil {
		log.Err(err).Str("file", flags.ConfigFile).Msg("An error occured while decoding")
		return
	}

	scheduler, s_err := lib.ReadSchedule()
	if s_err != nil {
		log.Warn().Err(s_err).Msg("The scheduler encountered an error. Cron scheduling will not be available.")
	}

	if tomlConf.CronSchedule != "" {
		if !scheduler.ShouldRunAgain(tomlConf.CronSchedule, "root") {
			if flags.ForceExecution {
				log.Info().Msg("Last execution is too short ago according to cron schedule, but we're forcing the execution.")
			} else {
				log.Warn().Msg("Last execution is too short ago, according to cron schedule. Aborting.")
				return
			}
		}
	}

	successfulSteps := make(map[string]map[string]string, 0)
	var failedSteps []string
	var skippedSteps []string
	var ignoredFails []string

	for _, n := range meta.Keys() {
		if strings.HasPrefix(n.String(), "handler.") && strings.Count(n.String(), ".") == 2 {
			log.Info().Msgf("Beginning execution of handler %s", n.String())
			handler, baseCfg := handlers.GetConfigByKey(&tomlConf.Handler, n.String())
			if handler == nil {
				log.Error().Msgf("Could not find handler for %s", n)
			}
			if baseCfg.CronSchedule != "" {
				if !scheduler.ShouldRunAgain(baseCfg.CronSchedule, n.String()) {
					log.Warn().Msgf("Last execution of %s is too short ago, according to handler cron schedule. Skipping handler.", n.String())
					skippedSteps = append(skippedSteps, n.String())
					continue
				}
			}
			if baseCfg.Skip {
				log.Info().Msg("Handler skipped as configured.")
				skippedSteps = append(skippedSteps, n.String())
				continue
			}
			err = handler.ValidateConfig()
			if err != nil {
				log.Err(err).Type("errorType", err).Msgf("Config for %s is invalid: ", n.String())
			}

			err, data := handler.Execute()
			if err != nil {
				if baseCfg.IgnoreFail {
					log.Info().Type("errorType", err).Msgf("An error occured but was ignored while handling %s:", n.String())
					ignoredFails = append(ignoredFails, n.String())
				}
				log.Err(err).Type("errorType", err).Msgf("An error occurred while handling %s:", n.String())
				failedSteps = append(failedSteps, n.String())
			} else {
				successfulSteps[n.String()] = data
				if baseCfg.CronSchedule != "" {
					scheduler.UpdateLastExecution(n.String())
				}
				log.Info().Msgf("Successfully executed %s", n)
			}
		}

		if strings.HasPrefix(n.String(), "notifier.") && strings.Count(n.String(), ".") == 2 {
			log.Info().Msgf("Beginning execution of notifier %s", n)
			notifier := notifiers.GetConfigByKey(&tomlConf.Notifier, n.String())
			if notifier == nil {
				log.Error().Msgf("Could not find notifier for %s", n)
			}
			if len(failedSteps) > 0 {
				notifier.SendFailure(successfulSteps, failedSteps, skippedSteps, ignoredFails)
			} else {
				notifier.SendSuccess(successfulSteps, skippedSteps, ignoredFails)
			}
			log.Info().Msgf("Successfully executed %s", n)
		}
	}

	scheduler.UpdateLastExecution("root")
}
