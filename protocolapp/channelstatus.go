// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// OnChannelStatus describes an on channel status message for Zello Websoocket interface
type OnChannelStatus struct {
	Command            string `json:"command,omitempty"`
	Channel            string `json:"channel,omitempty"`
	Status             string `json:"status,omitempty"`
	UsersOnline        int    `json:"users_online,omitempty"`
	ImagesSupported    bool   `json:"images_supported,omitempty"`
	TextingSupported   bool   `json:"texting_supported,omitempty"`
	LocationsSupported bool   `json:"locations_supported,omitempty"`
	Error              string `json:"error,omitempty"`
	ErrorType          string `json:"error_type,omitempty"`
}

// NewOnChannelStatus returns a new OnChannelStatus structure
func NewOnChannelStatus() *OnChannelStatus {
	return &OnChannelStatus{
		Command: OnChannelStatusEvent,
	}
}
