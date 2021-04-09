package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//Perform health check on Loggicat Cloud
func (w *Watcher) healthCheck() {
	var healthURL string
	if strings.HasPrefix(w.serverurl, "https://") || strings.HasPrefix(w.serverurl, "http://") {
		healthURL = w.serverurl + ":443/api/health"
	} else {
		healthURL = "http://" + w.serverurl + ":443/api/health"
	}
	response, err := http.Get(healthURL)
	if err != nil {
		fmt.Println("Failed to check Loggicat Cloud status, err : ", err)
		log.Fatal("Failed to check Loggicat Cloud status, err : ", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to parse Loggicat Cloud status, err : ", err)
		log.Fatal("Failed to parse Loggicat Cloud status, err : ", err)
	}

	var res serverResponseString
	err = json.Unmarshal(body, &res)

	if err != nil {
		fmt.Println("Failed to check Loggicat Cloud status, err : ", err)
		log.Fatal("Failed to check Loggicat Cloud status, err : ", err)
	}

	if res.Code != 200 {
		fmt.Println("Loggicat Cloud is down, err : ", res.Message)
		log.Fatal("Loggicat Cloud is down, err : ", res.Message)
	}

	log.Info("Loggicat Cloud is online")
	fmt.Println("Loggicat Cloud is online")
}

func (w *Watcher) setAccessToken() {
	var refreshToken string
	switch w.tokenStorage {
	case "redis":
		refreshToken = w.redisGet("refreshToken")
	case "memory":
		refreshToken = w.refreshToken
	}

	payload := map[string]string{
		"refreshToken": refreshToken,
	}
	respondBody, err := w.sendPostRequest("getAccessToken", payload)
	if err != nil {
		fmt.Println("Failed to send request to Loggicat Cloud, err : ", err)
		log.Fatal("Failed to send request to Loggicat Cloud, err : ", err)
	}

	var res accessTokenStruct
	if err := json.Unmarshal(respondBody, &res); err != nil {
		fmt.Println("Failed to parse access token from Loggicat Cloud, err : ", err)
		log.Fatal("Failed to parse access token from Loggicat Cloud, err : ", err)
	}

	if res.Code != 200 {
		w.redisSet("refreshToken", "")
		w.redisSet("accessToken", "")
		w.redisSet("accessTokenExpire", "")
		fmt.Println("Failed to get access token from Loggicat Cloud, err : ", res.ErrorMessage)
		log.Fatal("Failed to get access token from Loggicat Cloud, err : ", res.ErrorMessage)
	}

	now := time.Now().Unix()
	expireAt := now + res.ExpireTime
	expireAtString := strconv.FormatInt(expireAt, 10)

	switch w.tokenStorage {
	case "redis":
		w.redisSet("accessTokenExpire", expireAtString)
		w.redisSet("accessToken", res.AccessToken)
		w.redisSet("refreshToken", res.RefreshToken)
	case "memory":
		w.accessTokenExpire = expireAtString
		w.accessToken = res.AccessToken
		w.refreshToken = res.RefreshToken
	}
	log.Info("New access token created")
	fmt.Println("New access token created")
}

func (w *Watcher) register() (string, error) {
	payload := map[string]string{
		"hostName":  w.hostName,
		"watcherID": w.watcherID,
	}
	respondBody, err := w.sendPostRequest("registerWatcher", payload)
	if err != nil {
		return "", err
	}
	var res serverResponseString
	if err := json.Unmarshal(respondBody, &res); err != nil {
		fmt.Println("Failed to read server response during Watcher registration, err : ", err)
		log.Error("Failed to read server response during Watcher registration, err : ", err)
		return "", err
	}

	if res.Code != 200 || res.Message == "" {
		log.Info("Failed to obtain WatcherID, err : ", res.Message)
		fmt.Println("Failed to obtain WatcherID, err : ", res.Message)
		return "", errors.New(res.Message)
	}

	log.Info("WatcherID obtained from server")
	fmt.Println("WatcherID obtained from server")
	return res.Message, nil
}

func (w *Watcher) removeRelease(toRemoveString string) error {
	if toRemoveString != "" {
		toRemoveList := strings.Split(toRemoveString, ",")
		payload := map[string]interface{}{
			"esIDs":     toRemoveList,
			"watcherID": w.watcherID,
		}
		respondBody, err := w.sendPostRequest("removeRelease", payload)
		if err != nil {
			return err
		}
		if strings.Contains(string(respondBody), "200") {
			log.Info("Updated release information")
			fmt.Println("Updated release information")
			w.redisSet("toRemove", "")
		} else {
			return err
		}
	}
	return nil
}

//GetRelease : GetRelease
func (w *Watcher) GetRelease() {
	for {
		log.Info("GetReleaseSubProcess - Checking released findings from Loggicat Cloud")
		fmt.Println("GetReleaseSubProcess - Checking released findings from Loggicat Cloud")
		toRemoveString := w.redisGet("toRemove")
		if toRemoveString != "" {
			err := w.removeRelease(toRemoveString)
			if err != nil {
				continue
			}
		}

		payload := map[string]string{
			"hostName":  w.hostName,
			"watcherID": w.watcherID,
		}

		response, err := w.sendPostRequest("getRelease", payload)
		if err != nil {
			continue
		}
		var res serverResponseArray
		if err := json.Unmarshal(response, &res); err != nil {
			log.Error("Failed to parse server reponse when getting relased findings, err : ", err)
			fmt.Println("Failed to parse server reponse when getting relased findings, err : ", err)
			continue
		}
		if len(res.Message) != 0 {
			toRemoveString := w.redisGet("toRemove")
			allLines := map[string]string{}
			var count int

			for _, r := range res.Message {
				var curText string
				if val, ok := allLines[r["fileName"]]; ok {
					curText = val
				} else {
					curText = ""
				}
				curText += r["raw"]
				if !strings.HasSuffix(r["raw"], "\n") {
					curText += "\n"
				}
				allLines[r["fileName"]] = curText
				if toRemoveString != "" {
					toRemoveString += ","
				}
				toRemoveString += r["esID"]
				count++
			}

			for fn, text := range allLines {
				err := w.writeToLog(fn+".loggicat", text)
				if err != nil {
					return
				}
			}

			err := w.removeRelease(toRemoveString)
			if err != nil {
				return
			}
			w.redisSet("toRemove", toRemoveString)

		}
		log.Info("GetReleaseSubProcess - Finished checking released findings from Loggicat Cloud, sleeping now...")
		fmt.Println("GetReleaseSubProcess - Finished checking released findings from Loggicat Cloud, sleeping now...")
		time.Sleep(time.Duration(w.refreshTime) * time.Minute)
	}
}
