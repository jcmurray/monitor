// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// OnImage describes an on image message for Zello Websoocket interface
type OnImage struct {
	Command   string `json:"command,omitempty"`
	Channel   string `json:"channel,omitempty"`
	From      string `json:"from,omitempty"`
	For       string `json:"for,omitempty"`
	MessageID int    `json:"message_id,omitempty"`
	Type      string `json:"type,omitempty"`
	Height    int    `json:"height,omitempty"`
	Width     int    `json:"width,omitempty"`
	Source    string `json:"source,omitempty"`
}

// NewOnImage returns a new ZelloOnImage structure
func NewOnImage() *OnImage {
	return &OnImage{
		Command: OnImageEvent,
	}
}
