package watcher

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func (w *Watcher) processLogWithAPI(scanMode string, fileName string) error {
	fmt.Println("Scanning ", fileName)

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println("cannot open file, error : ", err)
		return err
	}

	defer file.Close()

	var offset int64
	var newFileSize int64

	info, err := os.Stat(fileName)
	if err != nil {
		log.Error("Failed to run os.stat")
		fmt.Println("Failed to run os.stat")
		return err
	}

	offsetString := w.redisGet(fileName + ":offset")
	newFileSize = info.Size()

	if scanMode == "watcher" {
		if offsetString == "" {
			offset = 0
		} else {
			offset, err = strconv.ParseInt(offsetString, 10, 64)
			if err != nil {
				log.Error("Failed to convert filesize to string")
				fmt.Println("Failed to convert filesize to string")
				return err
			}
		}

		//some editor will trigger event twice
		if offset == newFileSize {
			return nil
		}

		//when creating a new file with the same name
		if offset > newFileSize {
			offset = 0
		}
	}
	offset = 0
	r := bufio.NewReader(file)
	_, err = r.Discard(int(offset))
	if err != nil {
		log.Error("Failed to discard old file size when reading file change")
		fmt.Println("Failed to discard old file size when reading file change")
		return err
	}

	readSize := newFileSize - offset
	buf := make([]byte, 0, readSize)
	n, err := io.ReadFull(r, buf[:cap(buf)])
	buf = buf[:n]
	s := string(buf)

	payload := map[string]string{
		"data": s,
	}

	respondBody, err := w.sendPostRequest("uploadFile", payload)
	if err != nil {
		return err
	}
	var res uploadFileResponse
	if err := json.Unmarshal(respondBody, &res); err != nil {
		fmt.Println("Failed to parse server response during Watcher registration, err : ", err)
		log.Error("Failed to parse server response during Watcher registration err : ", err)
		return err
	}

	if res.Code != 200 {
		fmt.Println("Failed to scan file" + fileName + " err : " + res.Message)
		log.Error("Failed to scan file" + fileName + " err : " + res.Message)
		return errors.New("Failed to scan file" + fileName + " err : " + res.Message)
	}

	var nonVulnerableLines string
	switch scanMode {
	case "scanner":
		nonVulnerableLines = ""
		for i := range res.VulnerableLines {
			res.VulnerableLines[i]["fileName"] = fileName
			res.VulnerableLines[i]["hostName"] = w.hostName
		}
	case "watcher":
		for i := range res.VulnerableLines {
			res.VulnerableLines[i]["fileName"] = fileName
			res.VulnerableLines[i]["hostName"] = w.hostName
		}
		nonVulnerableLines = res.CleanLines
	}

	err = w.handleScanResult(scanMode, fileName, newFileSize, res.VulnerableLines, nonVulnerableLines)

	if err != nil {
		fmt.Println("Failed to create scan result output for", fileName)
		log.Error("Failed to create scan result output for", fileName)
		return err
	}

	return nil
}

func (w *Watcher) handleScanResult(uploadMode string, fileName string, newFileSize int64, newScanRes []map[string]string, nonVulnerableLines string) error {
	var err error
	var msg string

	if len(newScanRes) != 0 {
		switch w.outputMode {
		case "online":
			payload := map[string]interface{}{
				"scanRes":    newScanRes,
				"watcherID":  w.watcherID,
				"uploadMode": uploadMode,
			}
			_, err = w.sendPostRequest("uploadScanResult", payload)
			msg = "uploaded to Loggicat Cloud"
		case "offline":
			err = w.appendNewScanResult(newScanRes)
			msg = "appened to output file " + w.outputLocation
		}
		if err != nil {
			return err
		}
		log.Info(strconv.Itoa(len(newScanRes)) + " findings found and " + msg)
		fmt.Println(strconv.Itoa(len(newScanRes)) + " findings found and " + msg)
	}

	if nonVulnerableLines != "" {
		w.writeToLog(fileName+".loggicat", nonVulnerableLines)
	}

	if newFileSize != 0 {
		w.redisSet(fileName+":offset", strconv.FormatInt(newFileSize, 10))
	}

	return nil
}

func (w *Watcher) appendNewScanResult(newScanRes []map[string]string) error {
	_, err := os.Stat(w.outputLocation)
	if !os.IsNotExist(err) {
		jsonFile, err := os.Open(w.outputLocation)
		if err != nil {
			log.Error("Failed to open output file location, err : ", err)
			fmt.Println("Failed to open output file location, err : ", err)
			return err
		}
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var oldScanRes []map[string]string
		err = json.Unmarshal([]byte(byteValue), &oldScanRes)
		if err != nil {
			log.Error("Failed to unmarshal output result, err : ", err)
			fmt.Println("Failed to unmarshal output result, err : ", err)
			return err
		}
		newScanRes = append(oldScanRes, newScanRes...)
	}
	return writeStructToJSONFile(newScanRes, w.outputLocation)
}
