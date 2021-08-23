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
		leaks, err := w.processLog("scanner", filePath)
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
				util.PrintRed("failed to generate scan result file, " + err.Error())
			}
		}
	} else {
		util.PrintGreen("no data leaks found")
	}
}
