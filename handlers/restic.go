package handlers

import (
	"c0dermo/backup-samurai/config"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ResticHandler struct {
	config config.SamuariResticHandler
	logger zerolog.Logger
}

func NewResticHandler(config config.SamuariResticHandler, logger zerolog.Logger) *ResticHandler {
	handler := ResticHandler{
		config: config,
		logger: logger,
	}
	return &handler
}

type ResticBackupSummaryMessage struct {
	FilesNew            int     `json:"files_new"`
	FilesChanged        int     `json:"files_changed"`
	FilesUnmodified     int     `json:"files_unmodified"`
	DirsNew             int     `json:"dirs_new"`
	DirsChanged         int     `json:"dirs_changed"`
	DirsUnmodified      int     `json:"dirs_unmodified"`
	DataAdded           uint64  `json:"data_added"`
	DataCompressedAdded uint64  `json:"data_added_packed"`
	TotalFilesProcessed uint64  `json:"total_files_processed"`
	TotalBytesProcessed uint64  `json:"total_bytes_processed"`
	TimeTaken           float64 `json:"total_duration"`
}

func ParseResticBackupSummary(summaryLine string) (error, ResticBackupSummaryMessage) {
	var backupMsg ResticBackupSummaryMessage
	err := json.Unmarshal([]byte(summaryLine), &backupMsg)
	return err, backupMsg
}

func FormatResticBackupSummary(backupMsg ResticBackupSummaryMessage) []string {
	return []string{
		fmt.Sprintf("Restic backup successful in %.02fs, processing %d files (%s bytes), adding %s bytes (%s compressed) to the backup.",
			backupMsg.TimeTaken,
			backupMsg.TotalFilesProcessed,
			humanize.Bytes(backupMsg.TotalBytesProcessed),
			humanize.Bytes(backupMsg.DataAdded),
			humanize.Bytes(backupMsg.DataCompressedAdded),
		), fmt.Sprintf("Files: %d new, %d changed, %d unmodified",
			backupMsg.FilesNew,
			backupMsg.FilesChanged,
			backupMsg.FilesUnmodified,
		), fmt.Sprintf("Directories: %d new, %d changed, %d unmodified",
			backupMsg.DirsNew,
			backupMsg.DirsChanged,
			backupMsg.DirsUnmodified,
		)}
}

type ResticForgetSnapshotObject struct {
	ID string
}

type ResticForgetOutputMessage struct {
	Keep   []ResticForgetSnapshotObject
	Remove []ResticForgetSnapshotObject
}

func ParseResticForgetOutput(output string) (error, []ResticForgetOutputMessage) {
	var forgetMessage []ResticForgetOutputMessage
	err := json.Unmarshal([]byte(output), &forgetMessage)
	return err, forgetMessage
}

func FormatResticForgetOutput(forgetMessage []ResticForgetOutputMessage) string {
	keep, remove := 0, 0
	for _, m := range forgetMessage {
		keep += len(m.Keep)
		remove += len(m.Remove)
	}
	return fmt.Sprintf("Restic forget was successful: %d snapshots are kept, %d snapshots are removed", keep, remove)
}

type ResticCheckSummaryMessage struct {
	NumErrors          int64    `json:"num_errors"`
	BrokenPacks        []string `json:"broken_packs"`
	SuggestRepairIndex bool     `json:"suggest_repair_index"`
	SuggestPrune       bool     `json:"suggest_prune"`
}

func ParseResticCheckOutput(output string) (error, ResticCheckSummaryMessage) {
	var checkMessage ResticCheckSummaryMessage
	err := json.Unmarshal([]byte(output), &checkMessage)
	return err, checkMessage
}

func FormatResticCheckOutput(checkMessage ResticCheckSummaryMessage) (string, bool) {
	shouldWarn := checkMessage.NumErrors > 0 || len(checkMessage.BrokenPacks) > 0 || checkMessage.SuggestRepairIndex || checkMessage.SuggestPrune
	appendix := ""
	if checkMessage.SuggestRepairIndex {
		appendix += "Repair index suggested. "
	}
	if checkMessage.SuggestPrune {
		appendix += "Prune suggested. "
	}
	return fmt.Sprintf("Restic check was completed. Errors: %d. Broken packs: %d. %s",
		checkMessage.NumErrors,
		len(checkMessage.BrokenPacks),
		appendix,
	), shouldWarn
}

func (handler ResticHandler) Execute() (error, map[string]string) {
	handler.logger.Debug().Msg("Starting execution")
	data := make(map[string]string, 0)

	binary := "restic"
	if handler.config.Binary != "" {
		binary = handler.config.Binary
	}

	var command_options []string
	command_options = append(command_options, "--json")
	command_options = append(command_options, "-r", handler.config.Repository)
	command_options = append(command_options, "--password-file", handler.config.PasswordFile)

	// Testing repository
	needsInit := false
	test_arguments := append(command_options, "cat", "config")
	test_command := exec.Command(binary, test_arguments...)
	handler.logger.Debug().Strs("cmd", test_command.Args).Msg("Executing test command")
	output, err := test_command.CombinedOutput()
	if err != nil && !TestExitCodes(err, []int{10, 12}) {
		return FormatExitError(err, "failed to establish restic repo connection", output), make(map[string]string, 0)
	}
	switch test_command.ProcessState.ExitCode() {
	case 0:
		break
	case 10:
		needsInit = true
	case 12:
		return errors.New("restic repository password incorrect"), make(map[string]string, 0)
	default:
		return fmt.Errorf("failed to establish restic repo connection: %s", string(output[len(output)-1])), make(map[string]string, 0)
	}

	if !needsInit && !handler.config.Init {
		return errors.New("restic repo does not exist and should not be initialized"), make(map[string]string, 0)
	}

	// Init
	if needsInit {
		init_arguments := append(command_options, "init")
		init_command := exec.Command(binary, init_arguments...)
		handler.logger.Debug().Strs("cmd", init_command.Args).Msg("Executing init command")
		output, err := init_command.CombinedOutput()
		if err != nil {
			return FormatExitError(err, "failed to init restic repo", output), make(map[string]string, 0)
		}
	}

	// Backup
	if handler.config.Backup.RunBackup {
		backup_arguments := append(command_options, "backup")
		for _, inc := range handler.config.Backup.Include {
			backup_arguments = append(backup_arguments, inc)
		}
		for _, exc := range handler.config.Backup.Exclude {
			backup_arguments = append(backup_arguments, "--exclude", exc)
		}
		backup_command := exec.Command(binary, backup_arguments...)
		handler.logger.Debug().Strs("cmd", backup_command.Args).Msgf("Executing backup command")
		output, err := backup_command.CombinedOutput()
		if err != nil {
			return FormatExitError(err, "failed to run restic backup", output), make(map[string]string, 0)
		}
		err, backup_message := ParseResticBackupSummary(ExtractLastLine(output))
		if err != nil {
			log.Warn().Msg("could not parse restic backup result message")
		} else {
			for _, line := range FormatResticBackupSummary(backup_message) {
				log.Info().Msg(line)
			}

			data["backup.files_new"] = strconv.Itoa(backup_message.FilesNew)
			data["backup.files_changed"] = strconv.Itoa(backup_message.FilesChanged)
			data["backup.files_unmodified"] = strconv.Itoa(backup_message.FilesUnmodified)
			data["backup.dirs_new"] = strconv.Itoa(backup_message.DirsNew)
			data["backup.dirs_changed"] = strconv.Itoa(backup_message.DirsChanged)
			data["backup.dirs_unmodified"] = strconv.Itoa(backup_message.DirsUnmodified)
			data["backup.data_added"] = strconv.FormatUint(backup_message.DataAdded, 10)
			data["backup.data_compressed_added"] = strconv.FormatUint(backup_message.DataCompressedAdded, 10)
			data["backup.total_files_processed"] = strconv.FormatUint(backup_message.TotalFilesProcessed, 10)
			data["backup.total_bytes_processed"] = strconv.FormatUint(backup_message.TotalBytesProcessed, 10)
			data["backup.time_taken"] = strconv.FormatFloat(backup_message.TimeTaken, 'f', 2, 64)
		}
	}

	// Forget
	if handler.config.Forget.RunForget {
		forget_arguments := append(command_options, "forget")
		if handler.config.Forget.KeepLast > 0 {
			forget_arguments = append(forget_arguments, "--keep-last", strconv.Itoa(handler.config.Forget.KeepLast))
		}
		if handler.config.Forget.KeepHourly > 0 {
			forget_arguments = append(forget_arguments, "--keep-hourly", strconv.Itoa(handler.config.Forget.KeepHourly))
		}
		if handler.config.Forget.KeepDaily > 0 {
			forget_arguments = append(forget_arguments, "--keep-daily", strconv.Itoa(handler.config.Forget.KeepDaily))
		}
		if handler.config.Forget.KeepWeekly > 0 {
			forget_arguments = append(forget_arguments, "--keep-weekly", strconv.Itoa(handler.config.Forget.KeepWeekly))
		}
		if handler.config.Forget.KeepMonthly > 0 {
			forget_arguments = append(forget_arguments, "--keep-monthly", strconv.Itoa(handler.config.Forget.KeepMonthly))
		}
		if handler.config.Forget.KeepYearly > 0 {
			forget_arguments = append(forget_arguments, "--keep-yearly", strconv.Itoa(handler.config.Forget.KeepYearly))
		}
		if handler.config.Forget.DryRun {
			forget_arguments = append(forget_arguments, "--dry-run")
		}

		forget_command := exec.Command(binary, forget_arguments...)
		handler.logger.Debug().Strs("cmd", forget_command.Args).Msg("Executing forget command")
		output, err := forget_command.CombinedOutput()
		if err != nil {
			return FormatExitError(err, "failed to run restic forget", output), make(map[string]string, 0)
		}
		err, forget_message := ParseResticForgetOutput(ExtractLastLine(output))
		if err != nil {
			log.Warn().Msg("could not parse restic forget result message")
		} else {
			log.Info().Msg(FormatResticForgetOutput(forget_message))

			keep, remove := 0, 0
			for _, m := range forget_message {
				keep += len(m.Keep)
				remove += len(m.Remove)
			}

			data["forget.snapshots_kept"] = strconv.Itoa(keep)
			data["forget.snapshots_removed"] = strconv.Itoa(remove)
		}
	}

	// Prune
	if handler.config.RunPrune {
		prune_arguments := append(command_options, "prune")
		prune_command := exec.Command(binary, prune_arguments...)
		handler.logger.Debug().Strs("cmd", prune_command.Args).Msg("Executing prune command")
		output, err := prune_command.CombinedOutput()
		if err != nil {
			return FormatExitError(err, "failed to run restic prune", output), make(map[string]string, 0)
		}
	}

	// Check
	if handler.config.RunCheck {
		check_arguments := append(command_options, "check")
		check_command := exec.Command(binary, check_arguments...)
		handler.logger.Debug().Strs("cmd", check_command.Args).Msg("Executing check command")
		output, err := check_command.CombinedOutput()
		if err != nil {
			return FormatExitError(err, "failed to run restic check", output), make(map[string]string, 0)
		}
		err, check_message := ParseResticCheckOutput(ExtractLastLine(output))
		if err != nil {
			log.Warn().Msg("could not parse restic check result message")
		} else {
			msg, shouldWarn := FormatResticCheckOutput(check_message)
			if shouldWarn {
				log.Warn().Msg(msg)
			} else {
				log.Info().Msg(msg)
			}

			data["check.num_errors"] = strconv.FormatInt(check_message.NumErrors, 10)
			data["check.num_broken_packs"] = strconv.Itoa(len(check_message.BrokenPacks))
			if check_message.SuggestRepairIndex {
				data["check.suggest_repair_index"] = "yes"
			} else {
				data["check.suggest_repair_index"] = "no"
			}
			if check_message.SuggestPrune {
				data["check.suggest_prune"] = "yes"
			} else {
				data["check.suggest_prune"] = "no"
			}
		}
	}

	return nil, data
}

func (handler ResticHandler) ValidateConfig() error {
	if handler.config.Repository == "" {
		return errors.New("restic repository must be set")
	}
	if handler.config.PasswordFile == "" {
		return errors.New("restic password file must be set")
	}
	return nil
}
