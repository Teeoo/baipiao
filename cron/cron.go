package cron

import "github.com/robfig/cron/v3"

type TaskId *cron.EntryID

var Task = cron.New(cron.WithParser(cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)))
