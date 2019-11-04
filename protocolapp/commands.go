// cSpell.language:en-GB
// cSpell:disable

package protocolapp

// Commands we can zend to or receive from Zello
const (
	LogonRequest         string = "logon"
	OnChannelStatusEvent string = "on_channel_status"
	OnErrorEvent         string = "on_error"
	OnStreamStartEvent   string = "on_stream_start"
	OnStreamStopEvent    string = "on_stream_stop"
	OnImageEvent         string = "on_image"
	OnTextMessageEvent   string = "on_text_message"
	OnLocationEvent      string = "on_location"
	OnResponseEvent      string = "response"
)
