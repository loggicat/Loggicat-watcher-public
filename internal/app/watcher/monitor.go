package publicwatcher

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"loggicat.com/publicwatcher/internal/app/pkg/api"
	"loggicat.com/publicwatcher/internal/app/pkg/dbactions"
	"loggicat.com/publicwatcher/internal/app/pkg/util"
)

func (w *Watcher) MonitorFiles() {
	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		util.PrintRedFatal("failed to start file monitor, err : " + err.Error())
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-fsn.Events:
				ext := filepath.Ext(event.Name)
				if ext != ".log" && ext != ".txt" {
					continue
				}

				_, err := os.Stat(event.Name)
				if os.IsNotExist(err) {
					dbactions.Set(w.DB, event.Name, "")
					continue
				}

				//should check os statu first to see if we need to scan
				switch event.Op.String() {
				case "WRITE":
					totalLeaks, err := w.processLog("monitor", event.Name)
					if err != nil {
						util.PrintRed("failed to scan file " + event.Name)
						continue
					}
					if len(totalLeaks) != 0 {
						util.PrintGreen(strconv.Itoa(len(totalLeaks)) + " data leaks found")
						if w.OutputMode == "local" {
							err := util.SaveDataLeaksOffline(totalLeaks, w.OutputLocation)
							if err != nil {
								util.PrintRed("failed to generate scan result file, err : " + err.Error())
							}
						}
					}

				case "CREATE":
					if err := fsn.Add(event.Name); err != nil {
						util.PrintRed("failed to monitor path " + event.Name + ", err : " + err.Error())
						continue
					}
					totalLeaks, err := w.processLog("monitor", event.Name)
					if err != nil {
						util.PrintRed("failed to scan file " + event.Name)
						continue
					}
					if len(totalLeaks) != 0 {
						util.PrintGreen(strconv.Itoa(len(totalLeaks)) + " data leaks found")
						if w.OutputMode == "local" {
							err := util.SaveDataLeaksOffline(totalLeaks, w.OutputLocation)
							if err != nil {
								util.PrintRed("failed to generate scan result file, err : " + err.Error())
							}
						}
					}
				}
			case err := <-fsn.Errors:
				util.PrintRed("encoutnered FSn error err : " + err.Error())
			}
		}
	}()
	collectedFiles := util.CollectFiles(w.Path)
	for _, filePath := range collectedFiles {
		if err := fsn.Add(filePath); err != nil {
			util.PrintRed("failed to monitor path " + filePath + ", err : " + err.Error())
		}
	}
	<-done
}

func (w *Watcher) GetRelease() {
	sleep := w.RefreshTime
	for {
		util.PrintGreen("getting triaged logs")
		payload := map[string]string{
			"hostName": w.HostName,
			"uuid":     w.UUID,
		}
		triagedLeaks, err := api.GetRelease(w.Token, payload)
		if err != nil {
			util.PrintRed("failed to get release, err : " + err.Error())
			continue
		}
		grouped := groupRelease(triagedLeaks)

		toConfirm := []uint{}
		for filePath, info := range grouped {
			err = writeToLog(filePath, info.Text)
			if err == nil {
				toConfirm = append(toConfirm, info.IDs...)
			}
		}

		if len(toConfirm) != 0 {
			payload := map[string]interface{}{
				"uuid": w.UUID,
				"ids":  toConfirm,
			}
			err = api.ConfirmRelease(w.Token, payload)
			if err != nil {
				continue
			}
		}

		util.PrintGreen("finished getting triaged logs, going to sleep...zzzzz...")
		time.Sleep(time.Duration(sleep) * time.Minute)
	}
}

func groupRelease(leaks []api.Release) map[string]releaseGrouper {
	out := map[string]releaseGrouper{}
	for _, leak := range leaks {
		if cur, ok := out[leak.Path]; ok {
			cur.IDs = append(cur.IDs, leak.ID)
			cur.Text += leak.Line
			if !strings.HasSuffix(cur.Text, "\n") {
				cur.Text += "\n"
			}
			out[leak.Path] = cur
		} else {
			IDs := []uint{leak.ID}
			text := leak.Line
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			out[leak.Path] = releaseGrouper{
				IDs:  IDs,
				Text: text,
			}
		}
	}
	return out
}

func writeToLog(filePath string, text string) error {
	f, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		util.PrintRed("failed to open log file, err : " + err.Error())
		f.Close()
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		util.PrintRed("failed to write to log file, err : " + err.Error())
		f.Close()
		return err
	}
	f.Close()
	return nil
}
