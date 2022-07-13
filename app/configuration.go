package app

import "time"

type Configuration struct {
	Host           string
	Port           uint
	AssetsDir      string
	ResultsDir     string
	QueueSleepTime time.Duration
	JobTimeout     time.Duration
	AppLogPath     string
	WebLogPath     string
	QueuePath      string
}

func DefaultConfiguration() *Configuration {
	return &Configuration{
		AssetsDir:      "assets",
		QueueSleepTime: time.Second * 60,
		JobTimeout:     time.Hour * 4,
		AppLogPath:     "app.log",
		WebLogPath:     "web.log",
		QueuePath:      "queue.gob",
		ResultsDir:     "assets/results",
	}
}
