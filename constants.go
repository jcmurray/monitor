package main

// cSpell.language:en-GB
// cSpell:disable

// Version: Version string for the app
const (
	ApplicationVersion      = "0.1.3"
	DefaultHostname         = "zello.io"
	DefaultPort             = 443
	DefaultPath             = "/ws"
	DefaultProtocol         = "wss"
	DefaultListenOnly       = true
	DefaulW3WAPIKey         = "DEADBEEF"
	DefaultUseW3W           = false
	DefaultImageLogging     = false
	DefaultFrameRate        = 60
	DefaultSampleRate       = 16000
	DefaultChannels         = 1
	DefaultFramesPerPacket  = 2
	DefaultEnableAudio      = true
	DefaultRPCServerEnabled = false
	DefaultRPCServerPort    = 9998
)

// LogLevelStrings for config file
var LogLevelStrings = make([]string, logLevelTraceEndMarker)

// Log level 'enum'
const (
	LogLevelTrace          = iota // Something very low level
	LogLevelDebug                 // Useful debugging information
	LogLevelInfo                  // Something noteworthy happened
	LogLevelWarn                  // You should probably take a look at this
	LogLevelError                 // Something failed but I'm not quitting
	LogLevelFatal                 // Calls os.Exit(1) after logging
	LogLevelPanic                 // Calls panic() after logging I'm bailing
	logLevelTraceEndMarker        // --- end marker for later make([]string)
)

func init() {
	LogLevelStrings[LogLevelTrace] = "TRACE"
	LogLevelStrings[LogLevelDebug] = "DEBUG"
	LogLevelStrings[LogLevelInfo] = "INFO"
	LogLevelStrings[LogLevelWarn] = "WARN"
	LogLevelStrings[LogLevelError] = "ERROR"
	LogLevelStrings[LogLevelFatal] = "FATAL"
	LogLevelStrings[LogLevelPanic] = "PANIC"
}
