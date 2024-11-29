package model

type ReportPayload struct {
	AppID     string `json:"app_id"`
	Arch      string `json:"arch"`
	MachineID string `json:"machine_id"`
	OS        string `json:"os"`
	Version   string `json:"version"`
}
