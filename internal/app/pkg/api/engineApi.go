package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"loggicat.com/publicwatcher/internal/app/pkg/util"
)

func EngineHealthCheck(engineURL string) (bool, error) {
	var healthUrl string
	if strings.HasSuffix(engineURL, "/") {
		healthUrl = engineURL + "api/health"
	} else {
		healthUrl = engineURL + "/api/health"
	}
	req, err := http.NewRequest("GET", healthUrl, nil)
	if err != nil {
		util.PrintRed("Failed to create health check request, err : " + err.Error())
		return false, err
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		util.PrintRed("Failed to send health check request, err : " + err.Error())
		return false, err
	}

	return response.StatusCode == 200, nil
}

func EngineScanCode(engineURL string, payload map[string]string) ([]DataLeak, error) {
	var scanURL string
	if strings.HasSuffix(engineURL, "/") {
		scanURL = engineURL + "api/scan/code"
	} else {
		scanURL = engineURL + "/api/scan/code"
	}
	res, err := SendEngineRequest(scanURL, payload)
	if err != nil {
		return nil, err
	}
	return res.Leaks, nil
}

func EngineScanLog(engineURL string, payload map[string]string) ([]DataLeak, error) {
	var scanURL string
	if strings.HasSuffix(engineURL, "/") {
		scanURL = engineURL + "api/scan/log"
	} else {
		scanURL = engineURL + "/api/scan/log"
	}
	res, err := SendEngineRequest(scanURL, payload)
	if err != nil {
		return nil, err
	}
	return res.Leaks, nil
}

func SendEngineRequest(url string, payload map[string]string) (ServerResponse, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var decodedRes ServerResponse

	var req *http.Request
	var err error

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(payload)
	req, err = http.NewRequest("POST", url, payloadBuf)
	if err != nil {
		util.PrintRed("Failed to generate new request, err : " + err.Error())
		return decodedRes, err
	}

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		util.PrintRed("Failed to send a post request to server, err : " + err.Error())
		return decodedRes, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		util.PrintRed("Failed to parse server reponse of a post request, err : " + err.Error())
		return decodedRes, err
	}

	err = json.Unmarshal(body, &decodedRes)
	if err != nil {
		util.PrintRed("Failed to unmarshal server response, err : " + err.Error())
		return decodedRes, err
	}

	if decodedRes.Code != 200 {
		util.PrintRed("Response status not 400, response : " + string(body))
		return decodedRes, errors.New(string(body))
	}

	return decodedRes, nil
}
