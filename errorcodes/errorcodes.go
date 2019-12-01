// cSpell.language:en-GB
// cSpell:disable

package errorcodes

var (
	d map[string]string
)

func init() {
	d = map[string]string{}
	d["unknown command"] = "Server didn't recognize the command received from the client."
	d["internal server error"] = "An internal error occured within the server. If the error persists please contact us at support@zello.com."
	d["invalid json"] = "The command received included malformed JSON."
	d["invalid request"] = "The server couldn't recognize command format."
	d["not authorized"] = "Username, password or token are not valid."
	d["not logged in"] = "Server received a command before successful logon."
	d["not enough params"] = "The command doesn't include some of the required attributes."
	d["server closed connection"] = "The connection to Zello network was closed. You can try re-connecting."
	d["channel is not ready"] = "Channel you are trying to talk to is not yet connected. Wait for channel online status before sending a message"
	d["listen only connection"] = "The client tried to send a message over listen-only connection."
	d["failed to start stream"] = "Unable to start the stream for unknown reason. You can try again later."
	d["failed to stop stream"] = "Unable to stop the stream for unknown reason. This error is safe to ignore."
	d["failed to send data"] = "An error occured while trying to send stream data packet."
	d["invalid audio packet"] = "Malformed audio packet is received."
}

// Description from error code
func Description(code string) string {
	if desc, ok := d[code]; ok {
		return desc
	}
	return "Unrecognised error code."
}
