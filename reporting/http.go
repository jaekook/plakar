package reporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/PlakarKorp/plakar/utils"
)

type HttpEmitter struct {
	url   string
	token string
}

func (emitter *HttpEmitter) Emit(report *Report) error {
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to encode report: %s", err)
	}

	req, err := http.NewRequest("POST", emitter.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("plakar/%s (%s/%s)", utils.VERSION, runtime.GOOS, runtime.GOARCH))
	if emitter.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", emitter.token))
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if 200 <= res.StatusCode && res.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("request failed with status %s", res.Status)
}
