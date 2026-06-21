package lib

import (
	"encoding/json"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LastRunStepInfo map[string]int64

type Scheduler struct {
	lastRunStepInfo LastRunStepInfo
	logger          zerolog.Logger
}

func ReadSchedule() (Scheduler, error) {
	logger := log.With().Str("name", "scheduler").Logger()
	content, err := os.ReadFile("./.samuraiLastRun")
	if err != nil {
		return Scheduler{
			lastRunStepInfo: LastRunStepInfo{},
			logger:          logger,
		}, err
	}

	var lastRunStepInfo LastRunStepInfo
	err = json.Unmarshal(content, &lastRunStepInfo)
	if err != nil {
		return Scheduler{
			lastRunStepInfo: LastRunStepInfo{},
			logger:          logger,
		}, err
	}

	return Scheduler{
		lastRunStepInfo: lastRunStepInfo,
		logger:          logger,
	}, nil
}

func (sc Scheduler) GetLastExecution(key string) *time.Time {
	ts, ok := sc.lastRunStepInfo[key]
	if !ok {
		return nil
	}
	t := time.UnixMilli(ts)
	return &t
}

func (sc Scheduler) UpdateLastExecution(key string) {
	ts := time.Now().UnixMilli()
	sc.lastRunStepInfo[key] = ts

	sc.WriteToDisk()
}

func (sc Scheduler) WriteToDisk() error {
	f, err := os.Create("./.samuraiLastRun")
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(sc.lastRunStepInfo)
	if err != nil {
		return err
	}
	f.Write(data)
	return nil
}

func (sc Scheduler) ShouldRunAgain(cronSchedule string, key string) bool {
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := cronParser.Parse(cronSchedule)
	if err != nil {
		sc.logger.Debug().Str("schedule", cronSchedule).Str("key", key).Msg("Parsing failed, returning true")
		return true
	}
	lastExecution := sc.GetLastExecution(key)
	if lastExecution == nil {
		return true
	}
	sc.logger.Debug().Str("schedule", cronSchedule).Str("key", key).Msgf("Last execution: %s", lastExecution.GoString())
	nextExecution := schedule.Next(*lastExecution)
	sc.logger.Debug().Str("schedule", cronSchedule).Str("key", key).Msgf("Next execution: %s", nextExecution.GoString())
	if nextExecution.Before(time.Now()) || nextExecution.Equal(time.Now()) {
		return true
	} else {
		return false
	}
}
