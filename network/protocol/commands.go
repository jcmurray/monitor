// cSpell.language:en-GB
// cSpell:disable

package protocol

// Commands we can zend to or receive from Zello
const (
	LogonRequest    string = "logon"
	OnChannelStatus string = "on_channel_status"
	OnError         string = "on_error"
	OnStreamStart   string = "on_stream_start"
	OnStreamStop    string = "on_stream_stop"
	OnImage         string = "on_image"
	OnTextMessage   string = "on_text_message"
	OnLocation      string = "on_location"
)
