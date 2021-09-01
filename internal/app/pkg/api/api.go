package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"loggicat.com/publicwatcher/internal/app/pkg/util"
)

type ServerResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Leaks   []DataLeak `json:"leaks"`
	Release []Release  `json:"release"`
}

func ValidateToken(token string, uuid string) (bool, error) {
	payload := map[string]string{
		"uuid": uuid,
	}
	_, err := SendPostRequest("test", token, payload)

	if err != nil {
		return false, err
	}

	return true, nil
}

func SendPostRequest(apiEndpoint string, token string, payload interface{}) (ServerResponse, error) {
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	var decodedRes ServerResponse

	url := "https://app.loggicat.com/watcherapi/" + apiEndpoint

	var header = "Watcher " + token
	var req *http.Request
	var err error

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(payload)
	req, err = http.NewRequest("POST", url, payloadBuf)
	if err != nil {
		util.PrintRed("Failed to generate new request, err : " + err.Error())
		return decodedRes, err
	}

	req.Header.Add("Authorization", header)
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

type Match struct {
	LineNumber int
	Line       string
}

type DataLeak struct {
	Leak         string
	RuleName     string
	RuleType     string
	Line         string
	DependentID  string
	FileName     string
	LineNumber   int
	RedactedLine string
}

type Release struct {
	ID   uint   `json:"id"`
	Line string `json:"line"`
	Path string `json:"path"`
}

func ScanCodeSnippet(token string, payload interface{}) ([]DataLeak, error) {
	res, err := SendPostRequest("scanCodeSnippet", token, payload)
	if err != nil {
		return nil, err
	}
	return res.Leaks, nil
}

func ScanLogSnippet(token string, payload interface{}) ([]DataLeak, error) {
	res, err := SendPostRequest("scanLogSnippet", token, payload)
	if err != nil {
		return nil, err
	}
	return res.Leaks, nil
}

func GetRelease(token string, payload interface{}) ([]Release, error) {
	res, err := SendPostRequest("release/get", token, payload)
	if err != nil {
		return nil, err
	}
	return res.Release, nil
}

func ConfirmRelease(token string, payload interface{}) error {
	_, err := SendPostRequest("release/confirm", token, payload)
	if err != nil {
		return err
	}
	return nil
}
