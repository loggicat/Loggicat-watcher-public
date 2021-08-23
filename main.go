package main

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"loggicat.com/publicwatcher/internal/app/pkg/config"
	"loggicat.com/publicwatcher/internal/app/pkg/util"
	watcher "loggicat.com/publicwatcher/internal/app/watcher"
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
	fmt.Print(logo)
	fmt.Print("\n\n\n")

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: time.RFC3339})

	lumberjacklog := &lumberjack.Logger{
		Filename:   "logs/loggicatWatcher.log",
		MaxSize:    5000,
		MaxBackups: 180,
		MaxAge:     180,
		Compress:   false,
	}
	logMultiWriter := io.MultiWriter(os.Stdout, lumberjacklog)
	log.SetOutput(logMultiWriter)

	configFile := "./configs/watcherConfig.json"
	if configFile == "" {
		config.GenerateConfig()
	}

	watcher := watcher.Watcher{}
	watcher.Init(configFile)

	switch watcher.OperationMode {
	case "scan":
		util.PrintGreen("Switched to Scan Mode")
		watcher.ScanFiles()
	case "monitor":
		util.PrintGreen("Switched to Monitor Mode")
		go watcher.GetRelease()
		watcher.ScanFiles()
		watcher.MonitorFiles()
	}
}
