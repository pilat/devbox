package project

type stateFileStruct struct {
	Mounts   map[string]string `json:"mounts"`
	HasHosts bool              `json:"hasHosts"`
}
