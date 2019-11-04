// cSpell.language:en-GB
// cSpell:disable

package protocol

// Logon describes a logon message for Zello Websoocket interface
type Logon struct {
	Command      string `json:"command,omitempty"`
	Seq          int    `json:"seq,omitempty"`
	AuthToken    string `json:"auth_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Channel      string `json:"channel,omitempty"`
	ListenOnly   bool   `json:"listen_only,omitempty"`
}

// NewLogon returns a template logon message
func NewLogon() Logon {
	return Logon{
		Command: LogonRequest,
		Seq:     1,
	}
}

// IncompleteCredentials checks completeness of logon details
func (p Logon) IncompleteCredentials() bool {
	return (p.Channel == "" || p.AuthToken == "")
}
