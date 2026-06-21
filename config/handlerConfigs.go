package config

type SamuraiHandler struct {
	Restic   map[string]SamuariResticHandler
	Pgbackup map[string]SamuraiPgbackupHandler
	Rsync    map[string]SamuraiRsyncHandler
	Rclone   map[string]SamuraiRcloneHandler
	Script   map[string]SamuraiScriptHandler
}

type SamuraiHandlerBase struct {
	Skip         bool
	IgnoreFail   bool   `toml:"ignore_fail"`
	CronSchedule string `toml:"cron_schedule"`
}

type SamuraiResticHandlerBackup struct {
	RunBackup bool `toml:"run_backup"`
	Include   []string
	Exclude   []string
}

type SamuraiResticHandlerForget struct {
	RunForget   bool `toml:"run_forget"`
	KeepLast    int  `toml:"keep_last"`
	KeepHourly  int  `toml:"keep_hourly"`
	KeepDaily   int  `toml:"keep_daily"`
	KeepWeekly  int  `toml:"keep_weekly"`
	KeepMonthly int  `toml:"keep_monthly"`
	KeepYearly  int  `toml:"keep_yearly"`
	DryRun      bool `toml:"dry_run"`
}

type SamuariResticHandler struct {
	SamuraiHandlerBase
	Repository   string
	PasswordFile string `toml:"password_file"`
	Init         bool
	Backup       SamuraiResticHandlerBackup
	Forget       SamuraiResticHandlerForget
	RunCheck     bool `toml:"run_check"`
	RunPrune     bool `toml:"run_prune"`
	Options      string
	Binary       string
}

type SamuraiPgbackupHandlerDocker struct {
	RunInContainer bool `toml:"run_in_container"`
	Container      string
}

type SamuraiPgbackupHandler struct {
	SamuraiHandlerBase
	Docker       SamuraiPgbackupHandlerDocker
	DatabaseUser string `toml:"database_user"`
	Destination  string
	Gzip         bool
	Binary       string
}

type SamuraiRsyncHandlerLocation struct {
	Remote   bool
	User     string
	Host     string
	Protocol string
	Location string
}

type SamuraiRsyncHandler struct {
	SamuraiHandlerBase
	Source      SamuraiRsyncHandlerLocation
	Destination SamuraiRsyncHandlerLocation
	Binary      string
}

type SamuraiRcloneHandlerLocation struct {
	Path    string
	Backend string
	Flags   []string
}

type SamuraiRcloneHandler struct {
	SamuraiHandlerBase
	Source      SamuraiRcloneHandlerLocation
	Destination SamuraiRcloneHandlerLocation
	Binary      string
}

type SamuraiScriptHandler struct {
	SamuraiHandlerBase
	Binary  string
	Options []string
}
