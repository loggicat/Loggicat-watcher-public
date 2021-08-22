package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

//Init : Watcher init
func (w *Watcher) Init(configFile string) {
	log.Info("Watcher Initiating...")
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Println("Loggicat Cloud is down, err : ", err)
		log.Fatal("Loggicat Cloud is down, err : ", err)
	}

	w.hostName = hostName
	var conf ConfigStruct

	if configFile == "" {
		conf = w.GenerateConfig()
	} else {
		conf = w.ReadConfig(configFile)
	}

	w.outputMode = conf.OutputMode
	if w.outputMode != "offline" && w.outputMode != "online" {
		fmt.Println("Invalid value for outputmode, this can only be online or offline, exiting...")
		log.Fatal("Invalid value for outputmode, this can only be online or offline, exiting...")
	}
	w.OperationMode = conf.OperationMode
	if w.OperationMode != "watcher" && w.OperationMode != "scanner" {
		fmt.Println("Invalid value for OperationMode, this can only be watcher or scanner, exiting...")
		log.Fatal("Invalid value for OperationMode, this can only be watcher or scanner, exiting...")
	}

	w.redisurl = conf.Redisurl

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     w.redisurl,
		Password: "",
		DB:       0,
	})
	w.ctx = ctx
	w.rdb = rdb

	w.tokenStorage = conf.TokenStorage
	w.watcherID = w.redisGet("watcherID")

	switch w.tokenStorage {
	case "redis":
		if w.redisGet("refreshToken") == "" {
			w.redisSet("refreshToken", conf.RefreshToken)
			log.Info("Using Redis to store refresh tokens")
		}
	case "memory":
		w.refreshToken = conf.RefreshToken
		w.accessTokenExpire = ""
		log.Info("Using Memory to store refresh tokens")
	}
	w.checkAccessToken()

	w.monitoredFiles = []string{}
	w.refreshTime = conf.RefreshTime
	w.serverurl = conf.Serverurl
	w.files = conf.Files
	temp := []string{}
	for _, v := range conf.FileExtensions {
		if !strings.HasPrefix(v, ".") {
			v = "." + v
		}
		temp = append(temp, v)
	}
	w.fileExtensions = temp
	w.outputLocation = conf.OutputLocation

	watcherID := w.redisGet("watcherID")
	if watcherID == "" {
		log.Info("WatcherID is not in redis, registering...")
		esID, err := w.register()
		if err != nil {
			fmt.Println("Failed to register watcher, err : ", err)
			log.Fatal("Failed to register watcher, err : ", err)
		}
		w.redisSet("watcherID", esID)
		w.watcherID = esID
		log.Info("Watcher registered on Loggicat Cloud")
	} else {
		log.Info("WatcherID obtained from redis")
		w.watcherID = watcherID
	}

	w.healthCheck()
}

/*
	Monitor files : For Watcher mode
*/
//Noted that fsnotify can only watch ~8k files, https://github.com/fsnotify/fsnotify
func (w *Watcher) MonitorFiles() {
	log.Info("Start monitoring files")
	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Failed to start fsnotify, err :", err)
		log.Fatal("Failed to start fsnotify, err :", err)
	}
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-fsn.Events:
				if !inFileExts(filepath.Ext(event.Name), w.fileExtensions) {
					continue
				}
				_, err := os.Stat(event.Name)
				if os.IsNotExist(err) {
					w.redisSet(event.Name+":offset", "")
					continue
				}
				switch event.Op.String() {
				case "WRITE":
					err := w.processLogWithAPI("watcher", event.Name)
					if err != nil {
						log.Error("Error in scanFile", err)
						fmt.Println("Error in scanFile", err)
						continue
					}

				case "CREATE":
					if err := fsn.Add(event.Name); err != nil {
						log.Error("Error in fsn add", err)
						fmt.Println("Error in fsn add", err)
						log.Error("Failed to add file to fsn", event.Name)
						fmt.Println("Failed to add file to fsn", event.Name)
					}
					err := w.processLogWithAPI("watcher", event.Name)
					if err != nil {
						log.Error("Error in scanFile", err)
						fmt.Println("Error in scanFile", err)
						continue
					}
				}
			// watch for errors
			case err := <-fsn.Errors:
				log.Error("FSN ERROR : ", err)
				fmt.Println("FSN ERROR : ", err)
			}
		}
	}()
	w.checkFilesAndAddToFSN(fsn)
	<-done
}

/*
	ScanFiles : For Scanner mode
*/
func (w *Watcher) ScanFiles() {
	allFiles := w.collectFiles()

	for _, fileName := range allFiles {
		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			continue
		}
		log.Info("Scanning file", fileName)
		err = w.processLogWithAPI("scanner", fileName)
		if err != nil {
			log.Error("Error in scanFile", err)
			fmt.Println("Error in scanFile", err)
			continue
		}
		log.Info("Finished scanning", fileName)
	}

}

func (w *Watcher) checkFilesAndAddToFSN(fsn *fsnotify.Watcher) {
	allFiles := w.collectFiles()
	for _, f := range allFiles {
		if err := fsn.Add(f); err != nil {
			log.Error("Error in fsn add", err)
			fmt.Println("Error in fsn add", err)
			log.Error("Failed to add file to fsn")
			fmt.Println("Failed to add file to fsn")
		}
		//start := time.Now()
		err := w.processLogWithAPI("watcher", f)
		if err != nil {
			log.Error("Error in scanFile", err)
			fmt.Println("Error in scanFile", err)
			continue
		}
		//end := time.Now()
		//fmt.Println("Time diffrence is: ", end.Sub(start).Seconds())
	}
}
