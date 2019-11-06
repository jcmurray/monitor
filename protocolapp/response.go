// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// Response describes a response message for Zello Websoocket interface
type Response struct {
	Seq          int    `json:"seq,omitempty"`
	Success      bool   `json:"success,omitempty"`
	Error        string `json:"error,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	StreamID     int    `json:"stream_id,omitempty"`
}

// NewResponse returns a template response message
func NewResponse() *Response {
	return &Response{}
}
