package core

// =================================
// Application Event
// =================================

type CloseEvent struct {
	App App
}

type BeforeApplicationBootstrapped struct {
	App App
}

type AfterApplicationBootstrapped struct {
	App        App
	ConfigPath string
}

type AfterServiceExecuted struct {
	App       App
	Module    string
	Event     string
	Payload   map[string]interface{}
	OccuredAt int64
}
