package config

// ChangedMsg is sent when the configuration is reloaded.
type ChangedMsg struct {
	Config Config
}
