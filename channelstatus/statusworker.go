// cSpell.language:en-GB
// cSpell:disable

package channelstatus

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"

	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const ()

// StatusWorker stream worker
type StatusWorker struct {
	sync.Mutex
	command chan int
	log     *log.Entry
	id      int
	label   string
	workers *worker.Workers
}

// NewStatusWorker create a new Statusworker
func NewStatusWorker(workers *worker.Workers, id int, label string) *StatusWorker {
	return &StatusWorker{
		command: make(chan int),
		id:      id,
		label:   label,
		log:     log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers: workers,
	}
}

// Run is main function of this worker
func (w *StatusWorker) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	nw := w.findNetWorker()
	statusChannel := nw.Subscribe(w.id, protocolapp.OnChannelStatusEvent, w.label).Channel
	errorChannel := nw.Subscribe(w.id, protocolapp.OnErrorEvent, w.label).Channel

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case errorMessage := <-errorChannel:
			w.log.Debugf("Response: %s", string(errorMessage.([]byte)))

		case channelStatus := <-statusChannel:

			c := protocolapp.NewOnChannelStatus()
			err := json.Unmarshal(channelStatus.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			if w.blockedOnChannel(c) {
				w.log.Errorf("Error - Exiting - User: '%s' Blocked on Channel: '%s'", viper.GetString("logon.username"), c.Channel)
				continue
			}

			var statusMessage bytes.Buffer
			statusMessage.WriteString("Channel '%s' %s, %d users ")
			if c.ImagesSupported {
				statusMessage.WriteString("( YES Images / ")
			} else {
				statusMessage.WriteString("( NO Images / ")
			}
			if c.TextingSupported {
				statusMessage.WriteString("YES Texting / ")
			} else {
				statusMessage.WriteString("NO Texting / ")
			}
			if c.LocationsSupported {
				statusMessage.WriteString("YES Locations )")
			} else {
				statusMessage.WriteString("NO Locations )")
			}

			w.log.Infof(statusMessage.String(), c.Channel, c.Status, c.UsersOnline)

		case streamCommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", streamCommand)
				switch streamCommand {
				case worker.Terminate:
					w.log.Debugf("Terminating")
					break waitloop
				default:
					continue
				}
			} else {
				w.log.Info("Channel closed")
				break waitloop
			}
		}
	}

	nw.UnSubscribe(w.id, protocolapp.OnChannelStatusEvent)
	nw.UnSubscribe(w.id, protocolapp.OnErrorEvent)

	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *StatusWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *StatusWorker) Terminate() {
	w.Command(worker.Terminate)
}

// FindNetWorker find Net worker
func (w *StatusWorker) findNetWorker() *network.Networker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			return (*w.workers)[i].(*network.Networker)
		}
	}
	return nil
}

func (w *StatusWorker) blockedOnChannel(c *protocolapp.OnChannelStatus) bool {
	/*
	   Check for being blocked on channel
	   ==================================

	   If so then bail -- unrecoverable error

	   c.Error == "blocked"
	   c.ErrorType == "unknown"
	   c.Status == "offline"
	   c.UsersOnline == 0
	*/
	if strings.EqualFold(c.ErrorType, "unknown") &&
		strings.EqualFold(c.Error, "blocked") &&
		strings.EqualFold(c.Status, "offline") &&
		c.UsersOnline == 0 {
		return true
	}
	return false
}
