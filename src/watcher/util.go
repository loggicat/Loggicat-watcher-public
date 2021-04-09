package watcher

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func (w *Watcher) redisSet(key string, value string) error {
	err := w.rdb.Set(w.ctx, key, value, 0).Err()
	if err != nil {
		log.Error("error setting redis, err : ", err)
		fmt.Println("error setting redis, err : ", err)
		return err
	}
	return nil
}

func (w *Watcher) redisGet(key string) string {
	val, err := w.rdb.Get(w.ctx, key).Result()
	switch {
	case err == redis.Nil:
		log.Info("Redis key doesn't exist", key)
		fmt.Println("Redis key doesn't exist", key)
		return ""
	case err != nil:
		log.Error("error getting redis key", err)
		fmt.Println("error getting redis key", err)
		return ""
	default:
		return val
	}
}

//ReadConfig : read config files
func (w *Watcher) ReadConfig(configFile string) ConfigStruct {
	log.Info("Loading config parameters from watcher.json file")
	path := configFile
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to open the config file, err : ", err)
		log.Fatal("Failed to open the config file, err : ", err)
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Failed to read the config file, err : ", err)
		log.Fatal("Failed to read the config file, err : ", err)

	}
	var res ConfigStruct
	if err := json.Unmarshal(byteValue, &res); err != nil {
		fmt.Println("Failed to parse the config file, err : ", err)
		log.Fatal("Failed to parse the config file, err : ", err)
	}
	return res
}

func promptAndReturn(prompt string, reader *bufio.Reader) string {
	fmt.Println(prompt)
	fmt.Print("-> ")
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\r", "", -1)
	return text
}

func (w *Watcher) sendPostRequest(apiEndpoint string, payload interface{}) ([]byte, error) {
	var url string

	switch apiEndpoint {
	case "builtinRule":
		apiEndpoint = "getBuiltinRules"
	case "customRule":
		apiEndpoint = "getCustomRules"
	case "ignore":
		apiEndpoint = "getIgnoreList"
	case "redact":
		apiEndpoint = "getRedactList"
	}
	if strings.HasPrefix(w.serverurl, "https://") || strings.HasPrefix(w.serverurl, "http://") {
		url = w.serverurl + ":443/api/" + apiEndpoint
	} else {
		url = "http://" + w.serverurl + ":443/api/" + apiEndpoint
	}

	var accessToken string
	if w.tokenStorage == "redis" {
		accessToken = w.redisGet("accessToken")
	} else {
		accessToken = w.accessToken
	}
	var bearer = "Bearer " + accessToken
	var req *http.Request
	var err error

	if payload == nil {
		payload = map[string]string{
			"watcherID": w.watcherID,
		}
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(payload)
	req, err = http.NewRequest("POST", url, payloadBuf)
	if err != nil {
		log.Error("Failed to send a post request to server")
		fmt.Println("Failed to send a post request to server")
		return []byte{}, err
	}

	req.Header.Add("Authorization", bearer)
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		log.Error("Failed to send a post request to server")
		fmt.Println("Failed to send a post request to server")
		return []byte{}, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("Failed to parse server reponse of a post request")
		fmt.Println("Failed to parse server reponse of a post request")
		return []byte{}, err
	}
	return body, nil
}

func inFileExts(e string, a []string) bool {
	if e == ".loggicat" {
		return false
	}
	for _, i := range a {
		if i == e {
			return true
		}
	}
	return false
}

func visit(files *[]string, allowList []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("error in visit", err)
			fmt.Println("error in visit", err)
			log.Error("Cannot read file in Visit")
			fmt.Println("Cannot read file in Visit")
		}
		//!info.IsDir()
		if inFileExts(filepath.Ext(info.Name()), allowList) {
			*files = append(*files, path)
		}
		return nil
	}
}

func (w *Watcher) writeToLog(fileName string, text string) error {
	f, err := os.OpenFile(fileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Failed to create .loggicat log file, err : ", err)
		fmt.Println("Failed to create .loggicat log file, err : ", err)
		f.Close()
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		log.Error("Failed to write to .loggicat log file, err : ", err)
		fmt.Println("Failed to write to .loggicat log file, err : ", err)
		f.Close()
		return err
	}
	f.Close()
	return nil
}

func writeStructToJSONFile(inputStruct interface{}, outputLocation string) error {
	result, err := json.Marshal(inputStruct)
	if err != nil {
		log.Error("Failed to marshal input struct when writing struct to output json file, err : ", err)
		fmt.Println("Failed to marshal input struct when writing struct to output json file, err : ", err)
		return err
	}
	f, err := os.OpenFile(outputLocation, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Error("Failed open output location when writing struct to output json file, err : ", err)
		fmt.Println("Failed open output location when writing struct to output json file, err : ", err)
		f.Close()
		return err
	}
	_, err = io.WriteString(f, string(result))
	if err != nil {
		log.Error("Failed to write struct to output json file, err : ", err)
		fmt.Println("Failed to write struct to output json file, err : ", err)
		f.Close()
		return err
	}
	f.Close()
	return nil
}

func (w *Watcher) checkAccessToken() {
	var accessTokenExpire string
	var accessToken string
	var refreshToken string
	switch w.tokenStorage {
	case "redis":
		accessTokenExpire = w.redisGet("accessTokenExpire")
		accessToken = w.redisGet("accessToken")
		refreshToken = w.redisGet("refreshToken")
	case "memory":
		accessTokenExpire = w.accessTokenExpire
		accessToken = w.accessToken
		refreshToken = w.refreshToken
	default:
		fmt.Println("Invalid value for tokenStorage, it must be redis or memory, exiting...")
		log.Fatal("Invalid value for tokenStorage, it must be redis or memory, exiting...")
	}

	if accessTokenExpire != "" && accessToken != "" && refreshToken != "" {
		accessTokenExpireInt64, err := strconv.ParseInt(accessTokenExpire, 10, 64)
		if err != nil {
			log.Error("Failed to parse accessToken expire date, treating as valid token, err : ", err)
			fmt.Println("Failed to parse accessToken expire date, treating as valid token, err : ", err)
			return
		}
		now := time.Now().Unix()
		if accessTokenExpireInt64-now >= 7200 {
			log.Info("Checking access token expire time, do not need to renew.")
			fmt.Println("Checking access token expire time, do not need to renew.")
			return
		}
	}
	w.setAccessToken()
}

func (w *Watcher) findgFilesUnderFolder(fileLocationString string) ([]string, error) {
	var subFiles []string
	err := filepath.Walk(fileLocationString, visit(&subFiles, w.fileExtensions))
	if err != nil {
		log.Error("Error in filepath.walk", err)
		fmt.Println("Error in filepath.walk", err)
		log.Error("Unable to walk on file path when adding to FSN")
		fmt.Println("Unable to walk on file path when adding to FSN")
	}
	return subFiles, err
}

func (w *Watcher) collectFiles() []string {
	allFiles := []string{}
	for _, fileLocationString := range w.files {
		fileLocation, err := os.Stat(fileLocationString)
		if os.IsNotExist(err) {
			log.Error("error in os IsNotExist", err)
			fmt.Println("error in os IsNotExist", err)
			log.Error("File doesn't exist when collecting files")
			fmt.Println("File doesn't exist when collecting files")
			continue
		}
		if fileLocation.IsDir() {
			subFiles, err := w.findgFilesUnderFolder(fileLocationString)
			if err != nil {
				continue
			}
			allFiles = append(allFiles, subFiles...)
		} else {
			if len(w.fileExtensions) == 0 || inFileExts(filepath.Ext(fileLocationString), w.fileExtensions) {
				allFiles = append(allFiles, fileLocationString)
			}
		}
	}
	return allFiles
}

func (w *Watcher) GenerateConfig() ConfigStruct {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Generating new config file...")

	token := promptAndReturn("Please enter the refresh token : ", reader)
	mode := promptAndReturn("Please enter the watcher mode, this can be scanner or watcher : ", reader)
	if mode != "scanner" && mode != "watcher" {
		fmt.Println("User entered a wrong Watcher mode when generating config file, exiting...")
		log.Fatal("User entered a wrong Watcher mode when generating config file, exiting...")
	}
	tokenStorage := promptAndReturn("Please enter the token stroage mode, this can be redis or memory : ", reader)
	if tokenStorage != "redis" && mode != "memory" {
		fmt.Println("User entered a wrong token storage mode when generating config file, exiting...")
		log.Fatal("User entered a wrong token storage mode when generating config file, exiting...")
	}
	refreshTime := promptAndReturn("Please enter the refresh time between each pull, this is in minute(30 seems to be a good option) : ", reader)
	refreshTimeInt, err := strconv.Atoi(refreshTime)
	if err != nil {
		fmt.Println("User entered a non-integer value for refresh time when generating config file, exiting...")
		log.Fatal("User entered a non-integer value for refresh time when generating config file, exiting...")
	}
	if refreshTimeInt < 0 || refreshTimeInt > 720 {
		fmt.Println("User entered an integer value greater than 1 day or less than 1 minute for refresh time when generating config file, exiting...")
		log.Fatal("User entered an integer value greater than 1 day or less than 1 minute for refresh time when generating config file, exiting...")
	}
	serverURL := promptAndReturn("Please enter the url for Loggicat Cloud : ", reader)
	redisURL := promptAndReturn("Please enter the url for Reddis, including port number, usually this is localhost:6379 : ", reader)
	fileExtensionsString := promptAndReturn("Please enter the file formats for logs, seperate multiple values with ',', for example: log,txt ", reader)
	fileExtensions := strings.Split(fileExtensionsString, ",")
	if len(fileExtensions) == 0 {
		fmt.Println("User entered 0 file format when generating config file, exiting...")
		log.Fatal("User entered 0 file format when generating config file, exiting...")
	}
	filesString := promptAndReturn("Please enter the log files to monitor, seperate multiple values with ',', for example: /var/logs/,/var/logs/log2.txt ", reader)
	files := strings.Split(filesString, ",")
	if len(files) == 0 {
		fmt.Println("User entered 0 file to monitor when generating config file, exiting...")
		log.Fatal("User entered 0 file to monitor format when generating config file, exiting...")
	}
	outputMode := promptAndReturn("Please enter the output mode, this can be offline or online : ", reader)
	if outputMode != "online" && outputMode != "offline" {
		fmt.Println("User entered a wrong output mode when generating config file, exiting...")
		log.Fatal("User entered a wrong output mode when generating config file, exiting...")
	}
	outputLoc := promptAndReturn("Please enter the output location, an output file will be generated when Watcher is running offline, for example : /var/loggicat/output.json ", reader)

	ret := ConfigStruct{
		OperationMode:  mode,
		RefreshToken:   token,
		TokenStorage:   tokenStorage,
		RefreshTime:    refreshTimeInt,
		Serverurl:      serverURL,
		Redisurl:       redisURL,
		Files:          files,
		OutputMode:     outputMode,
		OutputLocation: outputLoc,
		FileExtensions: fileExtensions,
	}

	file, _ := json.MarshalIndent(ret, "", " ")
	err = ioutil.WriteFile("watcherConfig.json", file, 0600)
	if err != nil {
		fmt.Println("Failed to generate file watcherConfig.json", err)
		log.Fatal("Failed to generate file watcherConfig.json", err)
	}
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current path", err)
		log.Fatal("Failed to get current path", err)
	}

	fmt.Println("New config file watcherConfig.json generated in", path)
	log.Info("New config file watcherConfig.json generated in", path)
	return ret
}
