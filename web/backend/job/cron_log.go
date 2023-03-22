package job

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
)

var cron_log = logging.Logger("cron")
var _ cron.Logger = (*CronLog)(nil)

type CronLog struct {
	log *logging.ZapEventLogger
}

func NewCronLog() *CronLog {
	return &CronLog{
		log: cron_log,
	}
}

// Info logs routine messages about cron's operation.
func (cronLog *CronLog) Info(msg string, keysAndValues ...interface{}) {
	cronLog.log.With(keysAndValues...).Info(msg)
}

// Error logs an error condition.
func (cronLog *CronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	cronLog.log.With(keysAndValues...).Error(msg + err.Error())
}
