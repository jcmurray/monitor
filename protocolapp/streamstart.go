// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// OnStreamStart describes a stream start message for Zello Websoocket interface
type OnStreamStart struct {
	Command        string `json:"command,omitempty"`
	Type           string `json:"type,omitempty"`
	Codec          string `json:"codec,omitempty"`
	PacketDuration int    `json:"packet_duration,omitempty"`
	StreamID       int    `json:"stream_id,omitempty"`
	Channel        string `json:"channel,omitempty"`
	From           string `json:"from,omitempty"`
	CodecHeader    string `json:"codec_header,omitempty"`
	For            string `json:"for,omitempty"`
}

// NewOnStreamStart returns a new ZelloOnStreamStart structure
func NewOnStreamStart() *OnStreamStart {
	return &OnStreamStart{
		Command: OnStreamStartEvent,
	}
}
