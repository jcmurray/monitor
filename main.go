// cSpell.language:en-GB
// cSpell:disable

package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/jcmurray/monitor/audiodecoder"
	"github.com/jcmurray/monitor/authenticate"
	"github.com/jcmurray/monitor/channelstatus"
	"github.com/jcmurray/monitor/clientrpc"
	"github.com/jcmurray/monitor/images"
	"github.com/jcmurray/monitor/locations"
	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/streams"
	"github.com/jcmurray/monitor/texts"
	"github.com/jcmurray/monitor/util"
	"github.com/jcmurray/monitor/worker"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	mlog    *log.Entry
	workers worker.Workers
)

func init() {
	clog := new(logrus.TextFormatter)
	log.SetFormatter(clog)
	clog.TimestampFormat = "2006-01-02 15:04:05"
	clog.FullTimestamp = true
	log.SetLevel(log.DebugLevel)
	mlog = log.WithFields(log.Fields{"Component": "Main"})
}

func main() {
	var (
		done             chan struct{}
		terminateRequest chan int
		interrupt        chan os.Signal
		waitGroup        sync.WaitGroup
		once             sync.Once
	)
	done = make(chan struct{})
	terminateRequest = make(chan int, 10)
	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	timeLogger := time.NewTicker(10 * 60 * time.Second)

	var config string
	var loglevel string

	pflag.StringVarP(&config, "config", "c", "config", "Name of configuration to use, without the .yaml or .json etc. suffix")
	pflag.StringVarP(&loglevel, "loglevel", "l", "info", "Log level to set: info, warn, error, debug, trace, fatal, panic")
	pflag.Parse()

	viper.SetConfigName(config)
	viper.AddConfigPath(".")

	viper.SetDefault("loglevel", util.LogLevelStrings[util.LogLevelInfo])
	viper.BindPFlag("loglevel", pflag.Lookup("loglevel"))

	viper.SetDefault("server.host", util.DefaultHostname)
	viper.SetDefault("server.port", util.DefaultPort)
	viper.SetDefault("log.listen_only", util.DefaultListenOnly)

	viper.SetDefault("location.what3wordsapikey", util.DefaulW3WAPIKey)
	viper.SetDefault("location.what3words", util.DefaultUseW3W)

	viper.SetDefault("image.logging", util.DefaultImageLogging)

	viper.SetDefault("audio.enable", util.DefaultEnableAudio)
	viper.SetDefault("audio.framerate", util.DefaultFrameRate)
	viper.SetDefault("audio.samplerate", util.DefaultSampleRate)
	viper.SetDefault("audio.channels", util.DefaultChannels)
	viper.SetDefault("audio.framesperpacket", util.DefaultFramesPerPacket)

	viper.SetDefault("rpc.apienabled", util.DefaultRPCServerEnabled)
	viper.SetDefault("rpc.apiport", util.DefaultRPCServerPort)

	if err := viper.ReadInConfig(); err != nil {
		mlog.Fatalf("Config file error: %s", err)
	}

	configLogLevel := viper.GetString("loglevel")
	foundLogLevel := false
	for i := range util.LogLevelStrings {
		if strings.EqualFold(util.LogLevelStrings[i], configLogLevel) {
			mlog.Infof("Logging level being set to %s", util.LogLevelStrings[i])
			mlog.Info("")
			switch i {
			case util.LogLevelTrace:
				log.SetLevel(log.TraceLevel)
			case util.LogLevelDebug:
				log.SetLevel(log.DebugLevel)
			case util.LogLevelInfo:
				log.SetLevel(log.InfoLevel)
			case util.LogLevelWarn:
				log.SetLevel(log.WarnLevel)
			case util.LogLevelError:
				log.SetLevel(log.ErrorLevel)
			case util.LogLevelFatal:
				log.SetLevel(log.FatalLevel)
			case util.LogLevelPanic:
				log.SetLevel(log.PanicLevel)
			default:
				log.SetLevel(log.InfoLevel)
			}
			foundLogLevel = true
			break
		}
	}
	if !foundLogLevel {
		mlog.Infof("Unknown logging level, defaulting to %s", util.LogLevelStrings[util.LogLevelInfo])
		mlog.Info("")
		log.SetLevel(log.InfoLevel)
	}

	mlog.Info("Starting network worker")

	networker := network.NewNetworker(&workers, util.NewID(workers), "Network Worker")
	workers = append(workers, networker)
	waitGroup.Add(1)
	go networker.Run(&waitGroup, &terminateRequest)
	networker.Command(worker.Connect)

	authworker := authenticate.NewAuthWorker(&workers, util.NewID(workers), "Auth Worker")
	workers = append(workers, authworker)
	waitGroup.Add(1)
	go authworker.Run(&waitGroup, &terminateRequest)
	authworker.Command(worker.Logon)

	streamworker := streams.NewStreamWorker(&workers, util.NewID(workers), "Stream Worker")
	workers = append(workers, streamworker)
	waitGroup.Add(1)
	go streamworker.Run(&waitGroup, &terminateRequest)

	statusworker := channelstatus.NewStatusWorker(&workers, util.NewID(workers), "Status Worker")
	workers = append(workers, statusworker)
	waitGroup.Add(1)
	go statusworker.Run(&waitGroup, &terminateRequest)

	imageworker := images.NewImageWorker(&workers, util.NewID(workers), "Image Worker")
	workers = append(workers, imageworker)
	waitGroup.Add(1)
	go imageworker.Run(&waitGroup, &terminateRequest)

	textworker := texts.NewTextMessageWorker(&workers, util.NewID(workers), "Text Message Worker")
	workers = append(workers, textworker)
	waitGroup.Add(1)
	go textworker.Run(&waitGroup, &terminateRequest)

	locationworker := locations.NewLocationWorker(&workers, util.NewID(workers), "Location Worker")
	workers = append(workers, locationworker)
	waitGroup.Add(1)
	go locationworker.Run(&waitGroup, &terminateRequest)

	audioworker := audiodecoder.NewAudioWorker(&workers, util.NewID(workers), "Audio Worker")
	workers = append(workers, audioworker)
	waitGroup.Add(1)
	go audioworker.Run(&waitGroup, &terminateRequest)

	var rpcapiworker *clientrpc.RPCWorker
	if viper.GetBool("rpc.apienabled") {
		rpcapiworker = clientrpc.NewRPCWorker(&workers, util.NewID(workers), "RPC API Worker")
		workers = append(workers, rpcapiworker)
		waitGroup.Add(1)
		go rpcapiworker.Run(&waitGroup, &terminateRequest)
	}

waitLoop:
	for {
		select {
		case <-terminateRequest:
			mlog.Debug("Received 'terminateRequest' event")
			once.Do(func() {
				mlog.Debug("Closing 'done' channel")
				close(done)
			})

		case <-done:
			mlog.Debug("Received 'done' event")
			audioworker.Terminate()
			mlog.Debug("Terminated audioworker")
			imageworker.Terminate()
			mlog.Debug("Terminated imageworker")
			textworker.Terminate()
			mlog.Debug("Terminated textworker")
			locationworker.Terminate()
			mlog.Debug("Terminated locationworker")
			statusworker.Terminate()
			mlog.Debug("Terminated statusworker")
			streamworker.Terminate()
			mlog.Debug("Terminated streamworker")
			authworker.Terminate()
			mlog.Debug("Terminated authworker")
			networker.Terminate()
			mlog.Debug("Terminated networker")

			if viper.GetBool("rpc.apienabled") {
				rpcapiworker.Terminate()
				mlog.Debug("Terminated rpcapiworker ")
			}
			break waitLoop

		case <-timeLogger.C:
			mlog.Info("")
			mlog.Info("---------------------------------------------------------------")
			mlog.Infof("Current Time: %s", util.UtcTimeDate().Format(time.RFC3339))
			mlog.Info("---------------------------------------------------------------")
			mlog.Info("")

		case <-interrupt:
			mlog.Debug("Interrupt! Waiting for grace period before exiting")

			select {
			case <-done:
			case <-time.After(time.Second):
				mlog.Debug("Grace period timer expired -- exiting")
				terminateRequest <- 1
			}
		}
	}
	mlog.Debug("Finished monitoring events")
	mlog.Debug("Waiting for workers to finish")
	waitGroup.Wait()
	mlog.Info("Completed")
}
