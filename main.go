package main

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"loggicat.com/loggicatWatcher/src/watcher"
)

func main() {
	logo := `
	_       ___    ____   ____  ____     __   ____  ______      __    __   ____  ______     __  __ __    ___  ____  
	| |     /   \  /    | /    ||    |   /  ] /    ||      |    |  |__|  | /    ||      |   /  ]|  |  |  /  _]|    \ 
	| |    |     ||   __||   __| |  |   /  / |  o  ||      |    |  |  |  ||  o  ||      |  /  / |  |  | /  [_ |  D  )
	| |___ |  O  ||  |  ||  |  | |  |  /  /  |     ||_|  |_|    |  |  |  ||     ||_|  |_| /  /  |  _  ||    _]|    / 
	|     ||     ||  |_ ||  |_ | |  | /   \_ |  _  |  |  |      |  '  '  ||  _  |  |  |  /   \_ |  |  ||   [_ |    \ 
	|     ||     ||     ||     | |  | \     ||  |  |  |  |       \      / |  |  |  |  |  \     ||  |  ||     ||  .  \
	|_____| \___/ |___,_||___,_||____| \____||__|__|  |__|        \_/\_/  |__|__|  |__|   \____||__|__||_____||__|\_|	API Ver`
	fmt.Println(logo)

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: time.RFC3339})
	//log.SetFormatter(&log.JSONFormatter{})

	lumberjacklog := &lumberjack.Logger{
		Filename:   "logs/loggicatWatcher.log",
		MaxSize:    5000,
		MaxBackups: 180,
		MaxAge:     180,
		Compress:   false,
	}
	logMultiWriter := io.MultiWriter(os.Stdout, lumberjacklog)
	log.SetOutput(logMultiWriter)

	configFile := ""
	args := os.Args
	if len(args) > 1 {
		configFile = os.Args[1]
	}

	watcher := watcher.Watcher{}
	watcher.Init(configFile)

	switch watcher.OperationMode {
	case "watcher":
		log.Info("Switching to Watcher Mode...")
		go watcher.GetRelease()
		watcher.MonitorFiles()
	case "scanner":
		log.Info("Switching to Scanner Mode...")
		watcher.ScanFiles()
		log.Info("Scanned all files...Exiting...")
	}
}
