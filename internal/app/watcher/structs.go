package publicwatcher

import "github.com/xujiajun/nutsdb"

//Watcher : Watcher
type Watcher struct {
	EngineType     string
	EngineURL      string
	OperationMode  string
	Scope          string
	UUID           string
	Token          string
	RefreshTime    int
	Path           []string
	MonitoredFiles []string
	OutputMode     string
	OutputLocation string
	DB             *nutsdb.DB
	HostName       string
}

type releaseGrouper struct {
	Text string
	IDs  []uint
}
