package watcher

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type uploadFileResponse struct {
	Code            int
	VulnerableLines []map[string]string
	CleanLines      string
	Message         string
}

//ConfigStruct : ConfigStruct
type ConfigStruct struct {
	OperationMode  string   `json:"operationMode"`
	RefreshToken   string   `json:"refreshToken"`
	TokenStorage   string   `json:"tokenStorage"`
	RefreshTime    int      `json:"refreshTime"`
	Serverurl      string   `json:"serverurl"`
	Redisurl       string   `json:"redisurl"`
	Files          []string `json:"files"`
	OutputMode     string   `json:"outputMode"`
	OutputLocation string   `json:"outputLocation"`
	FileExtensions []string `json:"fileExtensions"`
}

//Watcher : Watcher
type Watcher struct {
	OperationMode     string
	hostName          string
	fileExtensions    []string
	watcherID         string
	refreshToken      string
	accessToken       string
	accessTokenExpire string
	tokenStorage      string
	refreshTime       int
	serverurl         string
	redisurl          string
	files             []string
	monitoredFiles    []string
	ctx               context.Context
	rdb               *redis.Client
	outputMode        string
	outputLocation    string
}

type ScanRes struct {
	findings  []map[string]string
	redacted  string
	lineCount int
}

type serverResponseString struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type serverResponseArray struct {
	Code    int64               `json:"code"`
	Message []map[string]string `json:"message"`
}

type accessTokenStruct struct {
	Code         int64  `json:"code"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpireTime   int64  `json:"expireTime"`
	ErrorMessage string `json:"message"`
}
