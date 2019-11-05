// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// OnLocation describes an on location message for Zello Websoocket interface
type OnLocation struct {
	Command          string  `json:"command,omitempty"`
	Channel          string  `json:"channel,omitempty"`
	From             string  `json:"from,omitempty"`
	For              string  `json:"for,omitempty"`
	MessageID        int     `json:"message_id,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
	FormattedAddress string  `json:"formatted_address,omitempty"`
	Accuracy         float64 `json:"accuracy,omitempty"`
}

// NewOnLocation returns a new ZelloOnLocation structure
func NewOnLocation() *OnLocation {
	return &OnLocation{
		Command: OnLocationEvent,
	}
}
