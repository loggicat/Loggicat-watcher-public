<div align="center">

<img src="https://raw.githubusercontent.com/loggicat/Loggicat-Cloud-Wiki/main/public/watcherLog.png" height="300" />

**Loggicat solves data leaks by shifting data security left**

[![Follow on Twitter](https://img.shields.io/twitter/follow/loggicat?style=social)](https://twitter.com/loggicat)

---

</div>

<h3 align="left">
For <a href="https://github.com/loggicat/Loggicat-watcher-public">Loggicat Watcher</a>
</h3>
<h3 align="left">
For <a href="https://github.com/loggicat/Loggicat-Cloud-Wiki">Loggicat Cloud</a>
</h3>

---

# Getting Started With Loggicat Watcher
* **[Overview](#overview)** What is Loggicat
* **[Versions](#versions)** Choose the right Loggicat Watcher version to use
* **[Features](#featrues)** Available features
* **[Prerequisite](#prerequisite)** Before running Loggicat Watcher
* **[Configuration](#configuration)** Settings and configurations
* **[Start Loggicat Watcher](#configuration)** Steps to run Loggicat Watcher
* **[Important Notes](#important-notes)** Please read this section before your first login
* **[Questions](#questions)** Answers to some common questions


---

# Overview
More information is avaiable at <a href="https://github.com/loggicat/Loggicat-Cloud-Wiki">Loggicat Cloud Wiki</a>
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
redis is required, if it is not installed, download the installer from the <a href="https://redis.io/">offical website</a> or follow the <a href="https://hub.docker.com/_/redis/">docker installation guide</a>
For open-source version : Go is required, <a href="https://golang.org/doc/install">Go installation guide</a> 

---

# Configuration
A configuration json file should contain following 
```
test
```

---

## Operation Modes
- **Watcher** : monitor files, changes will be scanned as well, this should be used for logs
- **Scanner** : one time scan, this should be used for logs and code

---

## Output Modes
- **Online** : Scan results will be sent to Loggicat Cloud
- **Offline** : Generate a local json file to store scan results, many features are not available in this mode.

---

---

# Important Notes
1. Non-nessccary builtin rules should be disabled to speed up the scan speed, however, generic rules such as "Generic Secrets" should always be enabled.
2. Ignore list has higher priority than redact list, so your finding will be ignored if you have the same keyword in both ignore and redact lists, a feature to improve this behavior is under development.
3. Please use the "Contact us" button to report any bug or feature request, this is the simplest and most efficient way.

