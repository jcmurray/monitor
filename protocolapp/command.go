// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// Command describes an generic command message for Zello Websoocket interface
type Command struct {
	Command string `json:"command,omitempty"`
}
