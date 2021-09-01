<div align="center">

<img src="https://raw.githubusercontent.com/loggicat/Loggicat-Cloud-Wiki/main/public/watcherLog.png" height="300" />

**Loggicat solves data leaks by shifting data security left**

[![Follow on Twitter](https://img.shields.io/twitter/follow/loggicat?style=social)](https://twitter.com/loggicat)

---

</div>

<h3 align="left">
For <a href="https://github.com/loggicat/watcher-public">Loggicat Watcher</a>
</h3>
<h3 align="left">
For <a href="https://github.com/loggicat/Loggicat-Cloud-Wiki">Loggicat Cloud</a>
</h3>
<h3 align="left">
For <a href="https://github.com/loggicat/scan-engine">Loggicat Scan Engine</a>
</h3>

---

# Getting Started With Loggicat Watcher
* **[Overview](#overview)** What is Loggicat
* **[Versions](#versions)** Choose the right Loggicat Watcher version to use
* **[Features](#features)** Available features
* **[Prerequisite](#prerequisite)** Before running Loggicat Watcher
* **[Configuration](#configuration)** Settings and configurations
* **[CLI Usages](#cli-usages)** How to use Loggicat Watcher CLI
* **[Important Notes](#important-notes)** Please read this section before your first login

---

# Overview
More information is available at <a href="https://github.com/loggicat/Loggicat-Cloud-Wiki">Loggicat Cloud Wiki</a>
<br />

---

# Versions
There are two versions available. This repo is the **open-source** version and should only be used to understand how Loggicat Watcher can be integrated in your dev environment, please use the closed-source version for production use and for better performance.
|  | How to scan | Scan speed | Keep data to local only | Limit | 
| ------------- | ------------- | ------------- | ------------- | ------------- | 
| Open-Source Ver  | Use Loggicat Cloud Engine API | Slow, this depends on network speed |  No, the open-source version must send text to Loggicat Cloud | 100 MB per user per day for free users |
| Closed-Source Ver  | Scan engine is built-in  | fast | Yes, scan engine can be used locally without sending anything to Loggicat Cloud | There is no limit |

---

# Features
- Scan local code for secrets and PII (Github repo scan and commit scan will be released in the future releases)
- Scan local logs for secrets and PII
- Monitoring local logs, vulnerable lines will be temporarily on-hold until released by an user (Log streaming is currentely not supported)

---

# Prerequisite
<a href="https://golang.org/dl/">GoLang</a> 1.15 or above
---

# Configuration
A configuration json file should contain following information
```
{
  "operationMode": "",              //scan or monitor
  "scope": "",                      //log or code

  "token": "",                      //Loggicat Cloud API token
  "uuid": "",                        //Loggicat Cloud API token UUID
  
  "engineType": "",                 //cloud or local
  "engineURL": "",                  //only used when engineType is local
  
  "refreshTime": ,                  //time gap to pull releases from Loggicat Cloud
  
  "path": [                         //folders to scan
      "" 
  ],
  
  "outputMode": "",                 //cloud or local
  "outputLocation": ""              //output file location, only used when outputMode is local
}
```
## Operation Modes
- **Monitor** : monitor files, changes will be scanned as well, this should be used for logs
- **Scan** : one time scan, this should be used for logs and code
In the Monitor mode, clean logs and released security findings will be appended to new logs files with **.loggicat** extension. <br />
For example, if you are forward myAppLog.txt to a log ingestion platform, you should now use myAppLog.txt.loggicat instead.

## Scope
- **Code** : Server will try to parse the supported languages such as .go or .xml before scanning
- **Log** : Server will try to parse logs in the popular formats before scanning

## API Token, UUID
API token and UUID can be generated on Loggicat Cloud. <br />

## Engine Type, Engine URL
Put "local" for the type if you use <a href="https://github.com/loggicat/scan-engine">Loggicat Scan Engine</a>

## Refresh time
Loggicat Clouds never push contents to Loggicat Watcher so in order to append triaged result to logs, Watcher periodically pulls the result from Loggicat Cloud, this configuration value is in minutes. 

### Path
Folders to be scanned

## Output Modes, Output Location
- **Cloud** : Scan results will be sent to Loggicat Cloud
- **local** : Generate a local json file to store scan results, many features are not available in this mode. Output location is used i nthis mode.

---

# CLI Usages

## Open-source version
```
//A template will be generated if configs/watcherConfig.json does not exist
go run main.go
```
---

# Important Notes
1. Non-nessccary builtin rules should be disabled to speed up the scan speed, however, generic rules such as "Generic Secrets" should always be enabled.
2. Ignore list has higher priority than redact list, so your finding will be ignored if you have the same keyword in both ignore and redact lists, a feature to improve this behavior is under development.
3. Monitor mode should only be used for logs, local code monitoring mode and code push monitoring are under development.

