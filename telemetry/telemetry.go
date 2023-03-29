package telemetry

type telemetryTracker interface {
	TrackEvent()
}

type EventContext struct {
	command   string
	eventName string
}
