package telemetry

import (
	"log"
	"time"

	"vizzlo.com/mixpanel"
)

var (
	MixPanelClientID string = "35454cba1d62292948c65d1ae6b6f621"
)

type telemetryTracker interface {
	TrackEvent()
}

type EventContext struct {
	command   string
	eventName string
}

// Mixpanel implementation
type MixpanelTelemetryTracker struct {
	clientId string
	client   *mixpanel.Client
	appName  string
}

func (m MixpanelTelemetryTracker) TrackEvent(eventContext EventContext, eventProps map[string]interface{}) {
	baseProps := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"context":   m.appName,
		"command":   eventContext.command,
	}

	err := m.client.Track(m.clientId, eventContext.eventName, baseProps)

	if err != nil {
		log.Fatal(err)
	}
}

func NewMixPanelClient(clientId string, appName string) MixpanelTelemetryTracker {
	mixpanelClient := mixpanel.New(clientId)
	return MixpanelTelemetryTracker{client: mixpanelClient, clientId: clientId, appName: appName}
}
