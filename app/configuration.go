package app

import "time"

type Configuration struct {
	WorkerSleepTime time.Duration
	AppLogPath      string
	RequestLogPath  string
	QueueStorePath  string
}

func DefaultConfiguration() *Configuration {
	return &Configuration{
		WorkerSleepTime: time.Second * 60,
		AppLogPath:      "app.log",
		RequestLogPath:  "request.log",
		QueueStorePath:  "queue.gob",
	}
}
