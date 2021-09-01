package publicwatcher

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xujiajun/nutsdb"
	"loggicat.com/publicwatcher/internal/app/pkg/api"
	"loggicat.com/publicwatcher/internal/app/pkg/config"
	"loggicat.com/publicwatcher/internal/app/pkg/dbactions"
	"loggicat.com/publicwatcher/internal/app/pkg/util"
)

//Init : Watcher init
func (w *Watcher) Init(configFile string) {

	util.PrintGreen("Watcher Initiating...")

	conf := config.ReadConfig(configFile)

	w.EngineType = conf.EngineType
	w.EngineURL = conf.EngineURL

	w.OutputMode = conf.OutputMode
	if w.OutputMode != "local" && w.OutputMode != "online" {
		util.PrintRedFatal("Invalid value for outputmode, this can only be online or offline, exiting...")
	}
	util.PrintGreen("Output mode set to " + w.OutputMode)

	w.OperationMode = conf.OperationMode
	if w.OperationMode != "monitor" && w.OperationMode != "scan" {
		util.PrintRedFatal("Invalid operation mode, this can only be monitor or scan, exiting...")
	}
	util.PrintGreen("Operation mode set to " + w.OperationMode)

	w.Scope = conf.Scope
	if w.Scope != "code" && w.OperationMode != "log" {
		util.PrintRedFatal("Invalid  scope, this can only be code or log, exiting...")
	}
	util.PrintGreen("Scope set to " + w.Scope)

	opt := nutsdb.DefaultOptions
	opt.Dir = "nutsdb"
	db, err := nutsdb.Open(opt)
	if err != nil {
		util.PrintRedFatal("Failed to connect to nutsdb, err : " + err.Error())
	}
	w.DB = db
	dbactions.Test(db)
	util.PrintGreen("nutsdb connected")

	switch w.EngineType {
	case "local":
		isEngineUp, err := api.EngineHealthCheck(w.EngineURL)
		if err != nil {
			util.PrintRedFatal("Failed to check Scan Engine health")
		}
		if !isEngineUp {
			util.PrintRedFatal("Loggicat Scan Engine is down")
		}
		util.PrintGreen("Loggicat Scan Engine URL set")
	case "cloud":
		isValid, err := api.ValidateToken(conf.Token, conf.UUID)
		if err != nil {
			util.PrintRedFatal("Failed to validate API credentials")
		}
		if !isValid {
			util.PrintRedFatal("Invalid API credentials")
		}

		w.Token = conf.Token
		w.UUID = conf.UUID
		util.PrintGreen("API secret and UUID validated")
		util.PrintGreen("API secret set")
		util.PrintGreen("API secret UUID set")
	default:
		util.PrintRedFatal("Invalid Scan Engine type " + w.EngineType + " ,this must be local or cloud")
	}

	w.Path = conf.Path
	util.PrintGreen("Path set ")

	w.OutputLocation = conf.OutputLocation
	util.PrintGreen("Output location set to " + w.OutputLocation)

	hostName, err := os.Hostname()
	if err != nil {
		if err != nil {
			util.PrintRedFatal("Failed to get hostname, err : " + err.Error())
		}
	}

	refresh := conf.RefreshTime
	if refresh == 0 {
		util.PrintRedFatal("refresh time is set to 0 mins")
	}
	w.RefreshTime = refresh

	w.HostName = hostName

	util.PrintGreen("Watcher Initiated")
}

func (w *Watcher) processCode(scanMode string, filePath string) ([]api.DataLeak, error) {
	util.PrintGreen("Start processing code " + filePath)
	info, err := os.Stat(filePath)
	if err != nil {
		util.PrintRed("failed to get file info, err : " + err.Error())
		return nil, err
	}
	fileSize := info.Size()

	if fileSize >= 10000000 {
		util.PrintRed("code size greater than 10MB is not supported")
		return nil, errors.New("code size greater than 10MB for file " + filePath)
	}

	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		util.PrintRed("failed to process code, err : " + err.Error())
		return nil, err
	}

	codeSnippet := string(buf)

	var leaks []api.DataLeak

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)

	switch w.EngineType {
	case "cloud":
		payload := map[string]interface{}{
			"codeSnippet":   codeSnippet,
			"uuid":          w.UUID,
			"filePath":      filePath,
			"fileName":      filepath.Base(filePath),
			"storeOnServer": w.OutputMode == "cloud",
			"hostName":      w.HostName,
			"extension":     ext,
		}
		leaks, err = api.ScanCodeSnippet(w.Token, payload)
		if err != nil {
			return nil, err
		}
	case "local":
		payload := map[string]string{
			"code":      codeSnippet,
			"extension": ext,
		}
		leaks, err = api.EngineScanCode(w.EngineURL, payload)
		if err != nil {
			return nil, err
		}
	}

	util.PrintGreen("Finished processing code " + filePath)

	return leaks, nil

}

func (w *Watcher) processLog(scanMode string, filePath string) ([]api.DataLeak, error) {
	util.PrintGreen("Start processing log file " + filePath)
	totalLeaks := []api.DataLeak{}

	file, err := os.Open(filePath)

	if err != nil {
		util.PrintRed("failed to open file to scan, err : " + err.Error())
		return nil, err
	}

	defer file.Close()

	var offset int64 = 0
	var newFileSize int64

	info, err := os.Stat(filePath)
	if err != nil {
		util.PrintRed("failed to get file info, err : " + err.Error())
		return nil, err
	}

	offsetStr, err := dbactions.Get(w.DB, filePath)
	if err != nil {
		return nil, err
	}

	newFileSize = info.Size()

	if scanMode == "monitor" {
		if offsetStr != "" {
			offset, err = strconv.ParseInt(offsetStr, 10, 64)
			if err != nil {
				util.PrintRed("Failed to convert filesize to string, err : " + err.Error())
				return nil, err
			}
		}

		if offset == newFileSize {
			util.PrintGreen("file has not changed, finished scanning")
			return totalLeaks, nil
		}
		if offset > newFileSize {
			offset = 0
		}
		_, err = file.Seek(offset, 0)
		if err != nil {
			util.PrintRed("Failed to discard offset, err : " + err.Error())
			return nil, err
		}
	}

	r := bufio.NewReader(file)

	toBeScanned := newFileSize - offset

	if toBeScanned == 0 {
		return nil, nil
	}

	carryOver := []byte{}
	var carryOverFlag bool = false
	var curScanSize int64
	for {
		//scan 10MB at a time
		if toBeScanned > 10000000 {
			curScanSize = 10000000
			carryOverFlag = true
		} else {
			curScanSize = toBeScanned
		}

		buf := make([]byte, 0, curScanSize)
		n, err := io.ReadFull(r, buf[:cap(buf)])

		if err != nil {
			util.PrintRed("Failed to read file content, err : " + err.Error())
			return nil, err
		}
		buf = buf[:n]

		if len(carryOver) != 0 {
			buf = append(carryOver, buf...)
			carryOver = []byte{}
		}

		if carryOverFlag {
			buf, carryOver, err = util.GetCarryOver(buf)
			if err != nil {
				return nil, err
			}
			carryOverFlag = false
		}

		logSnippet := string(buf)

		var leaks []api.DataLeak

		switch w.EngineType {
		case "cloud":
			payload := map[string]interface{}{
				"logSnippet":    logSnippet,
				"uuid":          w.UUID,
				"filePath":      filePath,
				"fileName":      filepath.Base(filePath),
				"storeOnServer": w.OutputMode == "cloud",
				"hostName":      w.HostName,
			}
			leaks, err = api.ScanLogSnippet(w.Token, payload)
			if err != nil {
				return nil, err
			}
		case "local":
			payload := map[string]string{
				"logSnippet": logSnippet,
			}
			leaks, err = api.EngineScanLog(w.EngineURL, payload)
			if err != nil {
				return nil, err
			}
		}

		for i, leak := range leaks {
			logSnippet = strings.ReplaceAll(logSnippet, leak.Line, leak.RedactedLine)
			temp := leak
			temp.FileName = filePath
			leaks[i] = temp
		}

		if w.OutputMode == "local" {
			totalLeaks = append(totalLeaks, leaks...)
		}

		err = util.WriteToLoggicatLog(filePath+".loggicat", logSnippet)
		if err != nil {
			return nil, err
		}

		offset += curScanSize
		if len(carryOver) != 0 {
			offset -= int64(len(carryOver))
		}
		offsetStr = strconv.FormatInt(offset, 10)

		util.PrintGreen("reader currentely at position " + offsetStr)
		err = dbactions.Set(w.DB, filePath, offsetStr)
		if err != nil {
			return nil, err
		}

		toBeScanned -= curScanSize
		if toBeScanned == 0 {
			break
		}
	}

	util.PrintGreen("Finished processing log " + filePath)

	return totalLeaks, nil
}
