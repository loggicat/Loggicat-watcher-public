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
* **[Operation Modes](#operation-modes)** Different operation modes
* **[Output Modes](#modes)** Different output modes
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

# Operation Modes
- Watcher : monitor files, changes will be scanned as well, this should be used for logs
- Scanner : one time scan, this should be used for logs and code

---

# Watcher Management
In order to leverage all features on Loggicat Cloud, a Loggicat Watcher must be used, Watcher Management is to monitor watcher activities and generate refresh tokens.

---

# Integrations
**All tokens/secrets/webhooks mentioned in this section are encrypted on Loggicat Cloud.**<br />
**Loggicat Cloud will never return plaintext secrets/token back to users, neither from UI or APIs.**<br />
Loggicat has integrated Github and Slack, other integrations(including Gitlab, Jira, Jenkins, etc.) are under development and will be released in the future.

## Github
Github integration turns logs with sensitive data to the exact code location, this can help developers to fix issues much faster. <br />
_Noted that : Github Code search/scan and commit monitoring are not released yet._ <br />

In order to use Github integration, a <a href="https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token">Github Personal Access Token</a> must be created and stored on Loggicat. <br />
Following scopes are required : 
  - Full access to repos, this is required in order to scan and search in private repos. public_repo if only for public repos
  - read:org 
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/github1.PNG" height="200"/>

Github Tokens should be added from the "Github Integration" tab and a name must be provided. <br />

<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/github2.PNG" height="200"/>
In this page, you can choose to add/remove/enable/disable Github tokens, the "Test" button will validate the entered Github token and return a list of repos. </ br>

Once at least one token is added to Loggicat Cloud, now users can go to "Findings" tab and trigger a scan manually. <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/github3.PNG" height="200"/>

_Noted that : Scans will be triggered automatically for newly added findings._ <br />

Users might see following Github search status:
  - Not started : A job has been been created, a manual scan might be needed
  - Pending : A job has been created and will be triggered soon
  - Owner Infomation Found : Search is done and Loggicat has found the owner
  - Owner Infomation Not Found : Search is done and Loggicat has not found the owner
  - No Github token available : No Github tokens to use
  - Invalid Gtihub tokens or Invalid confidence setting : Expired Github tokens

Once the result is ready, users can click on the "Display Owner Information" button(as shown in the previous paragraph) to view owner information. <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/github4.PNG" height="200"/>

Confidence is used to measure the accuracy, when the returned infomration seems irrelevant, raise the confidence level. When Loggicat can't find any owner information for many of the findings, try lower the confidence level. <br />
The default confidence is 70% and is configurable in "Github Integration" -> "Github Integration Settings" <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/github5.PNG" height="200"/>


## Slack
With Slack integration, Loggicat Cloud will be able to notify the right person or channel in real time. <br />

### Create a slack bot user and generate bot token
1. Create a slack app following this <a href="https://api.slack.com/authentication/basics">guide</a>. <br />
2. Once a slack app is created for your workplace, go to "OAuth & Permissions" to create a bot token. <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack1.PNG" height="200"/>
3. Create a bot token with following bot token scopes.
  - chat:write
  - im:write
  - incoming-webhook
  - users.profile:read
  - users:read
  - users:read.email
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack2.PNG" height="200"/>
4. Create a channel for Loggicat notifications on slack and install the app to the that channel. <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack3.PNG" height="200"/>
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack4.PNG" height="200"/>
5. Now you should have a slack token starting with xoxb-

### Add a slack bot token to Loggicat
Simply go to "Slack Integration" page. <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack5.PNG" height="200"/> <br />
Users can choose to make the default notification target to either a channel or an user. <br />
_noted that the user full name won't work, you will need to either use the email address or the user ID_ <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack6.PNG" height="200"/> <br />
Messages sent to this channel/user will only contain the number and the categories of findings. <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack7.PNG" />

### Add a user/channel mapping
Loggicat Cloud currentely supports 3 types of mappings in "Slack Integration" -> "Slack Integration Settings"
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack9.PNG" height="200"/>
  1. Repo name to Slack channel/Username : This can be used to notify the owner of a github repo, for example, a github teamA/repo1 should be mapped to a slack channel owned by teamA, so whenever Loggicat finds a vulnerability in that repo, they will be notified ASAP. 
  2. Username to Slack channle/Username : This maps an username on github to an username on slack, so whenever this user commits anything vulnerable, she/he will be notified.
  3. Hostname to Slack channel/Username : Security findings reported by Loggicat Watcher will always include a hostname, users can choose to use the hostname as a mapping source. For example, teamA owns a build machine jenkinsA so Loggicat will notify the team channel whenever it sees findings from jenkinsA.

_note : The first two mappings(repo name and username) will only be triggered with after Github owner search, while the hostname mapping doesn't need Github integration._

Sample message : <br />
<img src="https://github.com/loggicat/Loggicat-Cloud-Wiki/blob/main/public/slack8.PNG" height="200"/>


---

# Important Notes
1. Non-nessccary builtin rules should be disabled to speed up the scan speed, however, generic rules such as "Generic Secrets" should always be enabled.
2. Ignore list has higher priority than redact list, so your finding will be ignored if you have the same keyword in both ignore and redact lists, a feature to improve this behavior is under development.
3. Please use the "Contact us" button to report any bug or feature request, this is the simplest and most efficient way.

