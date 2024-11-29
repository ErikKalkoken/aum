package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"

	"example/telemetry/internal/model"
)

const (
	timeout = time.Second * 5
)

// Report submits a new report to the server.
func Report(host, appID, version string) error {
	return report(host, appID, version, nil)
}

// Report submits a new report to the server.
func ReportWithHttpClient(host, appID, version string, httpClient *http.Client) error {
	return report(host, appID, version, httpClient)
}

// Report submits a new report to the server.
func report(host, appID, version string, httpClient *http.Client) error {
	if host == "" || appID == "" || version == "" {
		return fmt.Errorf("incomplete data")
	}
	machineID, err := machineid.ProtectedID(appID)
	if err != nil {
		return err
	}
	v := model.ReportPayload{
		AppID:     appID,
		Arch:      runtime.GOARCH,
		MachineID: machineID,
		OS:        runtime.GOOS,
		Version:   version,
	}
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/create-report", host)
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	res, err := httpClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("report not created. Status: %d", res.StatusCode)
	}
	return nil
}
