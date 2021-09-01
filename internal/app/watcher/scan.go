package publicwatcher

import (
	"os"
	"strconv"

	"loggicat.com/publicwatcher/internal/app/pkg/api"
	"loggicat.com/publicwatcher/internal/app/pkg/util"
)

func (w *Watcher) ScanFiles() {
	collectedFiles := util.CollectFiles(w.Path)
	totalLeaks := []api.DataLeak{}
	for _, filePath := range collectedFiles {
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			continue
		}
		var leaks []api.DataLeak
		switch w.Scope {
		case "log":
			leaks, err = w.processLog("scanner", filePath)
		case "code":
			leaks, err = w.processCode("scanner", filePath)
		default:
			util.PrintRed("invalid scope, " + w.Scope)
		}
		if err != nil {
			util.PrintRed("failed to scan file " + filePath)
			continue
		}
		totalLeaks = append(totalLeaks, leaks...)
	}

	if len(totalLeaks) != 0 {
		util.PrintGreen(strconv.Itoa(len(totalLeaks)) + " data leaks found")
		if w.OutputMode == "local" {
			err := util.SaveDataLeaksOffline(totalLeaks, w.OutputLocation)
			if err != nil {
				util.PrintRed("failed to generate scan result file, err : " + err.Error())
			}
		}
	} else {
		util.PrintGreen("no data leaks found")
	}
}
