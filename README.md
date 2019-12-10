
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
- It does **NOT** support the sending of the following messages. This is simply because the application was originally designed to listen to traffic on Zello channels and not to originate any voice or other traffic.
  - Sending 'start_stream' messages
  - Sending 'stream data' messages
  - Sending 'stop_stream' messages
  - Sending 'send_image' messages
  - Sending 'send_location' messages
- It **does** support sending of the following messages. In order to do this I added an API using [gRPC](https://grpc.io) so that clients could send these message types.
  - Sending 'send_text_message' messages
- It can optionally save image data received to files.
- It supports playing of the received audio streams directly to the computer's speakers using the [PortAudio](http://www.portaudio.com) package.
- It uses Golang Modules to identify prerequisite packages.
- It supports a [gRPC](https://grpc.io) API to allow clients to:
  - Request information about the status of the server.
  - Send text messages on the open Zello channel.

The application itself is written as a set of concurrent GoRoutines, one each for:

- WebSocket network communications
- Managing the authentication of the application to the Zello infrastructure
- Managing starting and stopping of received audio streams
- Managing receipt of Images
- Managing receipt of Text Messages
- Managing receipt of Location data
- Managing the decoding of audio data and sending it to the sound card.
- Managing the [gRPC](https://grpc.io) API.

Configuration is via a YAML configuration file which is read and parsed using the Golang  `viper` package.

## Installation

### PortAudio

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

### Opus Codec

Zello uses the OPUS Audio Codec to transmit audio packets over WebSockets. OPUS packets are highly compressed and ideal for VoIP applications. This application decodes the OPUS packets into a PCM stream, ( signed, 16-bit integers, mono at 16000 samples per second). It does this using a Golang wrapper but the underlying [OPUS](https://opus-codec.org) libraries that need to be installed on your computer. There are binary and source distributions available on the OPUS web site and some simple installations using a package manager for your specific platform.

For example on MacOS, where I've developed and tested the application, it's available from **HomeBrew**:

```shell
brew install opus
brew install opus-tools
```

On Ubuntu:

```shell
sudo apt-get install pkg-config libopus-dev libopusfile-dev
```

I can't vouch for any other platforms.

### gRPC and Protocol Buffers

An API is exposed using [gRPC](https://grpc.io) which allows clients written in a number f languages to interact with the server. You first need to install the [Google Protocol Buffers](https://developers.google.com/protocol-buffers/) compiler and library. You can find instructions for your favourite language [here](https://developers.google.com/protocol-buffers/). 

For example on MacOS, where I've developed and tested the application, it's available from **HomeBrew**:

```shell
brew install protobuf
```

There are pre-built binaries available for other platforms and it can also be build from source in the standard **Autoconf** way. Check [here](https://github.com/protocolbuffers/protobuf/blob/master/src/README.md) for details.

Since this application is written in **Go** you need to install a **Go** specific plugin for **protobuf**. The easiest way is:

```shell
go get -u github.com/golang/protobuf/protoc-gen-go
```

### Installing and Building the application

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
rpc:
  apienabled: false ## true/false - enable or disable the gRPC API ( default false )
  apiport: 9998 ## Port the application will listen on for gRPC API **requests**
```

If you're interested in using the What3Words location setting get a [What3Words API Key](https://developer.what3words.com/public-api) from their developer site.

## gRPC Client Examples

### Go Example

The following is a simple **Go** client application that sends a text message over the gRPC API to the main application which then sends it out on the Zello channel it's connected to.

```Go
package main

import (
	"context"
	"time"

	clientapi "<< path of package where Protocol Buffer stubs have been built >>"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)


func main() {
  var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:9998", opts...)
	if err != nil {
		log.Info(err)
		return
	}
	defer conn.Close()
	c := clientapi.NewClientServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t, err := c.SendTextMessage(ctx, &clientapi.TextMessage{
		Message: "Hello World!",
		For:     "",
	})
	if err != nil {
		log.Infof("%#v", err)
		return
	}
	log.Infof("Status %v", t.Success)
	log.Infof("Message %s", t.Message)
}
```

### Java Example

In the `examples` directory there is a simple **Java** client application that sends a text message over the gRPC API to the main application which then sends it out on the Zello channel it's connected to as well as demonstrating **blocking** and **async streaming** use of the **gRPC** API. It uses **Maven** to orchestrate the build process so you may need to install it from [here](https://maven.apache.org).

Build it like this:

```shell
$ cd examples/javaclient/monitor-client-app
$ mvn -DskipTests package
...
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Total time:  17.623 s
[INFO] Finished at: 2019-12-09T23:03:55Z
[INFO] ------------------------------------------------------------------------
$ 
```

Run it like this:

```shell
$ java -jar target/monitor-client-app-1.0-SNAPSHOT-jar-with-dependencies.jar
2019-12-09 23:05:12 INFO  ClientApp:41 - Will try to send message: Hello World!, to:
2019-12-09 23:05:13 INFO  ClientApp:52 - Response: true, message: Text messaage for '' received: Hello World!
2019-12-09 23:05:13 INFO  ClientApp:56 - Querying status of server using Sync API
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Network Worker, Id: 9890
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Auth Worker, Id: 2635
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Stream Worker, Id: 932
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Status Worker, Id: 1438
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Image Worker, Id: 6144
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Text Message Worker, Id: 2810
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Location Worker, Id: 1944
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: Audio Worker, Id: 7562
2019-12-09 23:05:13 INFO  ClientApp:62 - Name: RPC API Worker, Id: 6021
2019-12-09 23:05:13 INFO  ClientApp:70 - Querying status of server using Async API
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Network Worker, Id: 9890
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 1944, Label: Location Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 1944, Label: Location Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 2810, Label: Text Message Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 2810, Label: Text Message Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 2810, Label: Text Message Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 1438, Label: Status Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 1438, Label: Status Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 6144, Label: Image Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 6144, Label: Image Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 6144, Label: Image Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 932, Label: Stream Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 932, Label: Stream Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 932, Label: Stream Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 932, Label: Stream Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 2635, Label: Auth Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 2635, Label: Auth Worker
2019-12-09 23:05:13 INFO  ClientApp:80 - 	Id: 2635, Label: Auth Worker
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Auth Worker, Id: 2635
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Stream Worker, Id: 932
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Status Worker, Id: 1438
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Image Worker, Id: 6144
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Text Message Worker, Id: 2810
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Location Worker, Id: 1944
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: Audio Worker, Id: 7562
2019-12-09 23:05:13 INFO  ClientApp:77 - Name: RPC API Worker, Id: 6021
2019-12-09 23:05:13 INFO  ClientApp:92 - Completed
$
```
### Nodejs Example

In the `examples` directory there is a simple **Nodejs** client application that sends a text message over the gRPC API to the main application which then sends it out on the Zello channel it's connected to as well as demonstrating **async streaming** use of the **gRPC** API. It uses **Npm** to orchestrate the build process so you may need to install it as well as **Nodejs** using the instructions [here][(https://maven.apache.org](https://nodejs.org/en/)).

Build it like this:

```shell
$ cd examples/nodeclient
$ npm install
audited 178 packages in 1.198s
found 0 vulnerabilities
$ 
```

Run it like this:

```shell
$ node ./monitor_client.js
Response:  {
  success: true,
  message: "Text messaage for '' received: Hello World!"
}
data called:  {
  worker_subscription: [
    { id: 7932, type: 'on_error', label: 'Location Worker' },
    { id: 7932, type: 'on_location', label: 'Location Worker' },
    { id: 6728, type: 'on_error', label: 'Text Message Worker' },
    { id: 6728, type: 'xxx_on_response', label: 'Text Message Worker' },
    { id: 6728, type: 'on_text_message', label: 'Text Message Worker' },
    { id: 9496, type: 'on_error', label: 'Status Worker' },
    { id: 9496, type: 'on_channel_status', label: 'Status Worker' },
    { id: 3396, type: 'xxx_on_image_data', label: 'Image Worker' },
    { id: 3396, type: 'on_error', label: 'Image Worker' },
    { id: 3396, type: 'on_image', label: 'Image Worker' },
    { id: 5071, type: 'xxx_on_stream_data', label: 'Stream Worker' },
    { id: 5071, type: 'on_error', label: 'Stream Worker' },
    { id: 5071, type: 'on_stream_stop', label: 'Stream Worker' },
    { id: 5071, type: 'on_stream_start', label: 'Stream Worker' },
    { id: 2086, type: 'on_error', label: 'Auth Worker' },
    { id: 2086, type: 'xxx_on_response', label: 'Auth Worker' },
    { id: 2086, type: 'connection', label: 'Auth Worker' }
  ],
  id: 8245,
  name: 'Network Worker'
}
data called:  { worker_subscription: [], id: 2086, name: 'Auth Worker' }
data called:  { worker_subscription: [], id: 5071, name: 'Stream Worker' }
data called:  { worker_subscription: [], id: 9496, name: 'Status Worker' }
data called:  { worker_subscription: [], id: 3396, name: 'Image Worker' }
data called:  { worker_subscription: [], id: 6728, name: 'Text Message Worker' }
data called:  { worker_subscription: [], id: 7932, name: 'Location Worker' }
data called:  { worker_subscription: [], id: 81, name: 'Audio Worker' }
data called:  { worker_subscription: [], id: 212, name: 'RPC API Worker' }
status received:  {
  code: 0,
  details: '',
  metadata: Metadata { _internal_repr: {}, flags: 0 }
}
end received
```