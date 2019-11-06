// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// OnStreamStop describes a stream stop message for Zello Websoocket interface
type OnStreamStop struct {
	Command  string `json:"command,omitempty"`
	StreamID int    `json:"stream_id,omitempty"`
}

// NewOnStreamStop returns a new OnStreamStop structure
func NewOnStreamStop() *OnStreamStop {
	return &OnStreamStop{
		Command: OnStreamStopEvent,
	}
}
