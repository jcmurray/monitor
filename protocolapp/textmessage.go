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

// SendTextMessage message for network worker
type SendTextMessage struct {
	Command string `json:"command,omitempty"`
	Seq     int    `json:"seq,omitempty"`
	Text    string `json:"text,omitempty"`
	For     string `json:"for,omitempty"`
}

// NewSendTextMessage returns a new Zello SendTextMessage structure
func NewSendTextMessage() *SendTextMessage {
	return &SendTextMessage{
		Command: TextMessageSendRequest,
	}
}

// TextMessageRequest comes via a POST request
type TextMessageRequest struct {
	For     string `json:"for,omitempty"`
	Message string `json:"message,omitempty"`
}

// TextMessageResponse to POST request
type TextMessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
