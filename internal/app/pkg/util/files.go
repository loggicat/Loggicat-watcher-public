package util

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

func CollectFiles(path []string) []string {
	collected := []string{}
	PrintGreen("started to collect files under given path")
	for _, fileLocationString := range path {
		fileLocation, err := os.Stat(fileLocationString)
		if err != nil {
			if os.IsNotExist(err) {
				PrintGreen("file doesn't exist when collecting files, skipping...")
				continue
			} else {
				PrintRed("failed to get file location, err : " + err.Error())
				continue
			}
		}

		if fileLocation.IsDir() {
			subFiles, err := GatherFilesInDir(fileLocationString)
			if err != nil {
				continue
			}
			collected = append(collected, subFiles...)
		} else {
			ext := filepath.Ext(fileLocationString)
			if ext == ".log" || ext == ".txt" {
				collected = append(collected, fileLocationString)
			}
		}
	}
	PrintGreen("finished to collect files under given path")
	return collected
}

func GatherFilesInDir(path string) ([]string, error) {
	var subFiles []string
	err := filepath.Walk(path, Visit(&subFiles))
	if err != nil {
		PrintRed("Failed to gather files in dir, err : " + err.Error())
	}
	return subFiles, err
}

func Visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			PrintRed("Failed to visit files in dir, err : " + err.Error())
		}
		//!info.IsDir()
		ext := filepath.Ext(info.Name())
		if ext == ".log" || ext == ".txt" {
			*files = append(*files, path)
		}
		return nil
	}
}

func SaveDataLeaksOffline(inputStruct interface{}, outputLocation string) error {
	result, err := json.Marshal(inputStruct)
	if err != nil {
		PrintRed("failed to marshal dataleaks, err : " + err.Error())
		return err
	}
	currentTime := time.Now()
	curTime := currentTime.Format("2006-Jan-02_3-4-5")

	f, err := os.OpenFile(curTime+"_"+outputLocation, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		PrintRed("failed to open output location, err : " + err.Error())
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, string(result))
	if err != nil {
		PrintRed("failed to save dataleaks to output file, err : " + err.Error())
		return err
	}
	PrintGreen("scan result updated at " + outputLocation)
	return nil
}

func WriteToLoggicatLog(fileName string, text string) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		PrintRed("failed to create .loggicat file, err : " + err.Error())
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		PrintRed("failed to write .loggicat file, err : " + err.Error())
		return err
	}
	return nil
}

func GetCarryOver(text []byte) ([]byte, []byte, error) {
	l := len(text) - 1
	for {
		cur := text[l]
		if cur == 10 {
			buf := text[:l]
			carryOver := text[l+1:]
			return buf, carryOver, nil
		}
		l -= 1
		if l == 0 {
			break
		}
	}
	return nil, nil, errors.New("can not find new line character in the given text")
}
