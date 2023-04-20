package telemetry

import (
	"github.com/google/uuid"
	"log"
	"time"

	"vizzlo.com/mixpanel"
)

type MixpanelTelemetryTracker struct {
	clientId string
	client   *mixpanel.Client
	appName  string
	version  string
	runId    string
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
	
	err := m.client.Track(m.runId, eventContext.EventName, trackProps)

	if err != nil {
		log.Println(trackProps)
		log.Println(err.Error())
	}
}

func NewMixPanelTelemetryClient(clientId string, appName string, version string) MixpanelTelemetryTracker {
	mixpanelClient := mixpanel.New(clientId)
	return MixpanelTelemetryTracker{
		client:   mixpanelClient,
		clientId: clientId,
		appName:  appName,
		version:  version,
		runId:    uuid.New().String(),
	}
}
