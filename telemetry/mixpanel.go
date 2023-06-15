package telemetry

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type MixpanelTelemetryTracker struct {
	client  *http.Client
	url     string
	appName string
	version string
	runId   string
}

/*
Helper func for combining two maps
This is used to combine our baseline props sent for all events
with event props given from a caller
*/
func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func (m MixpanelTelemetryTracker) TrackEvent(eventContext EventContext, eventProps map[string]interface{}) {
	baseProps := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"context":   m.appName,
		"command":   eventContext.Command,
		"version":   m.version,
	}

	// Combine our baseline props that we send for _ALL_ events with the passed in props from the event
	trackProps := mergeMaps(baseProps, eventProps)

	request := map[string]interface{}{
		"id":         m.runId,
		"event":      eventContext.EventName,
		"eventProps": trackProps,
	}
	jsonStr, err := json.Marshal(request)
	if err != nil {
		log.Println(err.Error())
		return
	}
	resp, err := m.client.Post(m.url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
	}
}

func NewMixPanelTelemetryClient(url string, appName string, version string) MixpanelTelemetryTracker {
	return MixpanelTelemetryTracker{
		client:  &http.Client{},
		url:     url,
		appName: appName,
		version: version,
		runId:   uuid.New().String(),
	}
}
