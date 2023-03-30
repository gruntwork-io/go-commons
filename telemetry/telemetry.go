package telemetry

type telemetryTracker interface {
	TrackEvent()
}

type EventContext struct {
	Command   string
	EventName string
}
