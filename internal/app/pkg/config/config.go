package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"loggicat.com/publicwatcher/internal/app/pkg/util"
)

//ConfigStruct : ConfigStruct
type ConfigStruct struct {
	EngineType     string   `json:"engineType"`
	EngineURL      string   `json:"engineURL"`
	OperationMode  string   `json:"operationMode"`
	Token          string   `json:"token"`
	UUID           string   `json:"uuid"`
	RefreshTime    int      `json:"refreshTime"`
	OutputMode     string   `json:"outputMode"`
	OutputLocation string   `json:"outputLocation"`
	Path           []string `json:"path"`
}

func promptAndReturn(prompt string, reader *bufio.Reader) string {
	fmt.Println(prompt)
	fmt.Print("-> ")
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\r", "", -1)
	return text
}

func GenerateConfig() ConfigStruct {
	reader := bufio.NewReader(os.Stdin)

	util.PrintGreen("Config file not found, generating a new config file...")

	fmt.Println("[+] Which Loggicat Scan Engine are you using?")
	fmt.Println("[++] Enter 'cloud' if you use Loggicat cloud ")
	fmt.Println("[++] Enter 'local' if you use the open source Loggicat Scan Engine ")
	engineType := promptAndReturn("[+] Scan Engine Type : ", reader)

	var uuid string
	var token string
	var engineURL string
	var refreshTimeInt int
	var outputMode string
	switch engineType {
	case "cloud":
		uuid = promptAndReturn("[+] Please enter the API token UUID: ", reader)
		token = promptAndReturn("[+] Please enter the API token : ", reader)

		refreshTime := promptAndReturn("Please enter the refresh time between each pull, this is in minute : ", reader)
		refreshTimeInt, err := strconv.Atoi(refreshTime)
		if err != nil {
			util.PrintRedFatal("User entered a non-integer value for refresh time when generating config file, exiting...")
		}
		if refreshTimeInt < 0 || refreshTimeInt > 720 {
			util.PrintRedFatal("User entered an integer value greater than 1 day or less than 1 minute for refresh time when generating config file, exiting...")
		}

		outputMode = promptAndReturn("Please enter the output mode, this can be local or online : ", reader)
		if outputMode != "online" && outputMode != "local" {
			util.PrintRedFatal("User entered a wrong output mode when generating config file, exiting...")
		}

	case "local":
		engineURL = promptAndReturn("[+] Please enter the Scan Engine URL : ", reader)
		outputMode = "local"

	default:
		util.PrintRedFatal("Invalid engine type " + engineType + " ,this must be local or cloud")
	}

	mode := promptAndReturn("[+] Please enter the watcher mode, this can be scan or monitor : ", reader)
	if mode != "scan" && mode != "monitor" {
		util.PrintRedFatal("User entered a wrong Watcher mode when generating config file, exiting...")
	}

	filesString := promptAndReturn("Please enter the paths to scan or monitor, seperate multiple values with ',', for example: /var/logs/,/var/logs/log2.txt ", reader)
	files := strings.Split(filesString, ",")
	if len(files) == 0 {
		util.PrintRedFatal("User entered 0 file to monitor format when generating config file, exiting...")
	}

	outputLoc := promptAndReturn("Please enter the output location, an output file will be generated when Watcher is running offline, for example : /var/loggicat/output.json ", reader)

	configLoc := promptAndReturn("Please enter the absolute path for generated json file, for example : /var/loggicat/configs/watcherConfig.json", reader)

	ret := ConfigStruct{
		EngineType:     engineType,
		EngineURL:      engineURL,
		UUID:           uuid,
		Token:          token,
		OperationMode:  mode,
		RefreshTime:    refreshTimeInt,
		Path:           files,
		OutputMode:     outputMode,
		OutputLocation: outputLoc,
	}

	file, _ := json.MarshalIndent(ret, "", " ")
	err := ioutil.WriteFile(configLoc, file, 0600)
	if err != nil {
		util.PrintRedFatal("Failed to generate file watcherConfig.json" + err.Error())
	}

	util.PrintGreen("New config file watcherConfig.json generated  in current directory")
	return ret
}

//ReadConfig : read config files
func ReadConfig(configFile string) ConfigStruct {
	util.PrintGreen("Loading config parameters from " + configFile)
	path := configFile
	jsonFile, err := os.Open(path)
	if err != nil {
		util.PrintRedFatal("Failed to open the config file, " + err.Error())
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		util.PrintRedFatal("Failed to read the config file, " + err.Error())
	}
	var res ConfigStruct
	if err := json.Unmarshal(byteValue, &res); err != nil {
		util.PrintRedFatal("Failed to parse the config file, " + err.Error())
	}
	util.PrintGreen("Config loaded successfully")
	return res
}
