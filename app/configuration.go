package app

import "time"

type Configuration struct {
	DevelopmentMode bool
	Host            string
	Port            uint
	AssetsDir       string
	ResultsDir      string
	QueueSleepTime  time.Duration
	JobTimeout      time.Duration
	LogPath         string
	QueuePath       string
}

func DefaultConfiguration() *Configuration {
	return &Configuration{
		AssetsDir:      "assets",
		QueueSleepTime: time.Second * 60,
		JobTimeout:     time.Hour * 4,
		LogPath:        "app.log",
		QueuePath:      "queue.gob",
		ResultsDir:     "assets/results",
	}
}
