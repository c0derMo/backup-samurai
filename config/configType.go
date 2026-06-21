package config

type SamuraiConfig struct {
	CronSchedule string `toml:"cron_schedule"`
	Handler      SamuraiHandler
	Notifier     SamuraiNotifier
}
