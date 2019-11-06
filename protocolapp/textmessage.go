// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// OnTextMessage describes an on location message for Zello Websoocket interface
type OnTextMessage struct {
	Command   string `json:"command,omitempty"`
	Channel   string `json:"channel,omitempty"`
	From      string `json:"from,omitempty"`
	For       string `json:"for,omitempty"`
	MessageID int    `json:"message_id,omitempty"`
	Text      string `json:"text,omitempty"`
}

// NewOnTextMessage returns a new ZelloOnTextMessage structure
func NewOnTextMessage() *OnTextMessage {
	return &OnTextMessage{
		Command: OnTextMessageEvent,
	}
}
