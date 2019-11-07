
# Simple Zello Golang client using the Zello Websocket API

This project is licensed under the terms of the MIT license.

It provides a simple example of a Golang application that uses the
[Zello Websocket API](https://github.com/zelloptt/zello-channel-api/blob/master/API.md). It's simply a toy that I wrote to understand how Zello actually worked and hopefully you can find some educational use in it.

# This sample application 

Supports:

- Zello Websockets API Version 1.0
- A subset of the API messages is implemented. Namely:
  - Sending 'logon' message
  - Receiving 'response' message
  - Receiving 'on_channel_status' messages
  - Receiving 'on_stream_start' messages
  - Receiving 'stream data' messages -- raw audio in OPUS packets
  - Receiving 'on_stream_stop' messages
  - Receiving 'on_error' messages
  - Receiving 'on_image' messages
  - Receiving 'image data' messages -- raw images in JPEG format
  - Receiving 'on_text_messages'
  - Receiving 'on_location_messages'
- It does **NOT** support the sending of the following messages. This is simple because the application is designed to listen to traffic on Zello channels and not to originate any voice or other traffic.
  - Sending 'start_stream' messages
  - Sending 'stream data' messages
  - Sending 'stop_stream' messages
  - Sending 'send_image' messages
  - Sending 'send_text_message' messages
  - Sending 'send_location' messages
- It can optionally same image data received to files.
- It supports playing of the received audio streams directly to the computer's speakers using the [PortAudio](http://www.portaudio.com) package.
- It uses Golang Modules to identify prerequisite packages.

The application itself is written as a set of concurrent GoRoutines, one each for:

- WebSocket network communications
- Managing the authentication of the application to the Zello infrastructure
- Managing starting and stopping of received audio streams
- Managing receipt of Images
- Managing receipt of Text Messages
- Managing receipt of Location data
- Managing the decoding of audio data and sending it to the sound card.

Configuration is via a YAML configuration file which is read and parsed using the Golang  `viper` package.

## Installation

The application uses a Golang wrapper to talk to the [PortAudio](http://www.portaudio.com) libraries and you need to ensure the PortAudio package itself is correctly installed on your computer. PortAudio is used to play audio streams directly to the PCs sound card. There are binary and source distributions available on the PortAudio web site and some simple installations using a package manager for your specific platform:

For example on MacOS, where I've developed and tested the application, it's available from **HomeBrew**:

```shell
brew install portaudio
```

On Ubuntu:

```shell
sudo apt-get install portaudio19-dev
```

I can't vouch for any other platforms.

Zello uses the OPUS Audio Codec to transmit audio packets over WebSockets. OPUS packets are highly compressed and ideal for VoIP applications. This application decodes the OPUS packets into a PCM stream, ( signed, 16-bit integers, mono at 16000 samples per second). It does this using a Golang wrapper but the underlying [OPUS](https://opus-codec.org) libraries that need to be installed on your computer. There are binary and source distributions available on the OPUS web site and some simple installations using a package manager for your specific platform.

For example on MacOS, where I've developed and tested the application, it's available from **HomeBrew**:

```shell
brew install opus
brew install opus-tools
```

On Ubuntu:

```shell
sudo apt-get install opus-tools
```
I can't vouch for any other platforms.

Once these prerequisites are installed you can install the application itself as:

```shell
git clone https://github.com/jcmurray/monitor.git
make
```

## Quickstart

The `monitor` application supports just two command line options. All other options are specified in 
a configuration file ( default file `config.yaml` in the directory where you run `monitor`)

```shell
$ ./monitor --help
Usage of ./monitor:
  -c, --config string     Name of configuration to use, without the .yaml or .json etc. suffix (default "config")
  -l, --loglevel string   Log level to set: info, warn, error, debug, trace, fatal, panic (default "info")
```

So, to run the application using the default configuration `config` just do this. It will read the configuration from `config.yaml`

```shell
$ ./monitor
INFO[2019-11-07 11:04:17] Logging level being set to INFO               Component=Main
INFO[2019-11-07 11:04:17]                                               Component=Main
INFO[2019-11-07 11:04:17] Starting network worker                       Component=Main
INFO[2019-11-07 11:04:17] Hostname: zello.io, Port: 443                 ID=5947 Label="Network Worker"
WARN[2019-11-07 11:04:17] Attempt to send data on disconnected websocket  ID=5947 Label="Network Worker"
INFO[2019-11-07 11:04:19] Stream id 30002 Started - from 'Jay 1956' on 'Network Radios' for ''  ID=2762 Label="Stream Worker"
INFO[2019-11-07 11:04:19] Channel 'Network Radios' online, 118 users ( NO Images / NO Texting / YES Locations )  ID=3367 Label="Status Worker"
^C
INFO[2019-11-07 11:04:24] Completed                                     Component=Main
$ 
```

To stop the application I just entered CTRL-C to initiate the close process.

To use a different configuration file do the following. In this case the configuration is `config-MFPC` which will read from `config-MFPC.yaml` to connect to my own private Zello channel. The error message shows that this channel was not available to the account I was running the app under.

```shell
$ ./monitor --config config-MFPC
INFO[2019-11-07 11:06:43] Logging level being set to INFO               Component=Main
INFO[2019-11-07 11:06:43]                                               Component=Main
INFO[2019-11-07 11:06:43] Starting network worker                       Component=Main
INFO[2019-11-07 11:06:43] Hostname: zello.io, Port: 443                 ID=4814 Label="Network Worker"
WARN[2019-11-07 11:06:43] Attempt to send data on disconnected websocket  ID=4814 Label="Network Worker"
INFO[2019-11-07 11:06:44] Channel 'Murray Family Private Channel' offline, 0 users ( NO Images / NO Texting / NO Locations )  ID=2061 Label="Status Worker"
INFO[2019-11-07 11:06:44] Completed
```

If you're interested in what is going on under the scenes try the 'debug' and 'trace' logging options.For example:

```shell
$ ./monitor -l debug
INFO[2019-11-07 11:11:48] Logging level being set to DEBUG              Component=Main
INFO[2019-11-07 11:11:48]                                               Component=Main
INFO[2019-11-07 11:11:48] Starting network worker                       Component=Main
DEBU[2019-11-07 11:11:48] Worker Started                                ID=9542 Label="Auth Worker"
DEBU[2019-11-07 11:11:48] Entering Select                               ID=9542 Label="Auth Worker"
DEBU[2019-11-07 11:11:48] Received command 3                            ID=9542 Label="Auth Worker"
DEBU[2019-11-07 11:11:48] Worker Started                                ID=7456 Label="Network Worker"
INFO[2019-11-07 11:11:48] Hostname: zello.io, Port: 443                 ID=7456 Label="Network Worker"
DEBU[2019-11-07 11:11:48] resetRetryCounters(): w.retryCount = 8, retryInterval 2, inRetryProcess false, thisInterval 86400,  ID=7456 Label="Network Worker"
DEBU[2019-11-07 11:11:48] Worker Started                                ID=1164 Label="Audio Worker"
DEBU[2019-11-07 11:11:48] Worker Started                                ID=2359 Label="Image Worker"
DEBU[2019-11-07 11:11:48] Entering Select                               ID=7456 Label="Network Worker"
DEBU[2019-11-07 11:11:48] Worker Started                                ID=6584 Label="Stream Worker"
DEBU[2019-11-07 11:11:48] Worker Started                                ID=5883 Label="Status Worker"
DEBU[2019-11-07 11:11:48] Entering Select                               ID=5883 Label="Status Worker"
DEBU[2019-11-07 11:11:48] Worker Started                                ID=6392 Label="Text Message Worker"
DEBU[2019-11-07 11:11:48] Entering Select                               ID=6392 Label="Text Message Worker"
WARN[2019-11-07 11:11:48] Attempt to send data on disconnected websocket  ID=7456 Label="Network Worker"
DEBU[2019-11-07 11:11:48] Entering Select                               ID=9542 Label="Auth Worker"
DEBU[2019-11-07 11:11:48] Worker Started                                ID=5424 Label="Location Worker"
DEBU[2019-11-07 11:11:48] Entering Select                               ID=5424 Label="Location Worker"
DEBU[2019-11-07 11:11:48] Entering Select                               ID=2359 Label="Image Worker"
DEBU[2019-11-07 11:11:48] Received command 0                            ID=7456 Label="Network Worker"
DEBU[2019-11-07 11:11:48] Connecting                                    ID=7456 Label="Network Worker"
DEBU[2019-11-07 11:11:48] Connecting to: wss://zello.io:443/ws          ID=7456 Label="Network Worker"
DEBU[2019-11-07 11:11:48] Audio loop setup calculated buffer length for PCM 'out' buffer: 1920  ID=1164 Label="Audio Worker"
...
```

## Configuration file details

The minimal configuration file that you can use must have the four values shown below set.

```yaml
logon:
  channel: <Channel Name>
  username: <Zello user name>
  password: <Zello username password>
  auth_token: <Zello developer Authentication Token>
```

In addition to your Zello user name and password you need a Zello API token since the API is still in beta. All you need to do is go to the [Zello Developer Console](https://developers.zello.com), sign up for an account and follow the instructions to create a token. Tokens expire after a month or so, so if it expires just go and create a new one.

An example file might look like this (the password and token are obviously not real ones):

```yaml
logon:
  channel: Network Radios
  username: NR1314
  password: dsdjsjkJLjLljJljkl
  auth_token: eyJhbGciOiJSUzI1NiIsInR5cCI6I ... 942F4/PBauE4g==
```

The full configuration file will all values that have sensible defaults looks like this:

```yaml
loglevel: info ## one of info, warn, error, debug, trace, fatal, panic - defaults to info
server:
  host: zello.io ## Zello server (default)
  port: 443 ## Port on Zello Server ( secure Websockets port ) (default)
logon:
  channel: Network Radios ## Name of Zello Channel
  username: NR1314 ## Zello username
  password: dsdjsjkJLjLljJljkl ## password for Zello username
  auth_token: eyJhbGciOiJSUzI1NiIsInR5cCI6I ... 942F4/PBauE4g== ## Zello authentication token
  listen_only: true ## true/false - Tell Zello server I only want to listen on this connection (default true)
location:
  what3words: true ## true/false - optionally resolve locations to What3Words location strings (default false)
  what3wordsapikey: XXXXXXXX ## you need a What3Words developer key to use this ( default 'DEADBEEF')
image:
  logging: true ## true/false - write received image files to the current directory ( default false)
audio: ## used to calculate the required size of the audio PCM buffer ( 1920 bytes of signed, 16-bit integers)
  framerate: 60 ## Zello uses 60ms OPUS Frames -- Recommend not to change!!! (default 60)
  samplerate: 16000 ## Zello uses 16000/s Frame rate -- Recommend not to change!!! (default 16000)
  channels: 1 ## Zello uses a single channel (mono) -- Recommend not to change!!! (default 1)
  framesperpacket: 2 ## Zello uses 2 OPUS Frames per packet -- Recommend not to change!!! (default 2)

```

If you're interested in using the What3Words location setting get a [What3Words API Key](https://developer.what3words.com/public-api) from their developer site.