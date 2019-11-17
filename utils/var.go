package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	Logger  *log.Entry
	SerfLog *os.File
)

func Init() {
	// init logger
	init_log()
}

func init_log() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05"})

	// 日志参数
	if Config().Debug {
		log.SetLevel(log.DebugLevel)
	}
	root, _ := os.Getwd()

	logpath := root + "/var/"
	_, err := os.Stat(logpath)
	if err != nil {
		os.Mkdir(logpath, os.ModePerm)
	}
	logfile := logpath + "app.log"
	serfile := logpath + "serf.log"

	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.Info("Failed to log app.log to file, using default stderr")
	}
	SerfLog, err = os.OpenFile(serfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Info("Failed to log serf.log to file, using default stderr")
	}
	Logger = log.WithFields(log.Fields{"Commit By": "usdn"})
}
