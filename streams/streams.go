// cSpell.language:en-GB
// cSpell:disable

package streams

import (
	"encoding/json"
	"sync"

	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
)

const ()

// StreamWorker stream worker
type StreamWorker struct {
	sync.Mutex
	command chan int
	log     *log.Entry
	id      int
	label   string
	workers *worker.Workers
}

// NewStreamWorker create a new Streamworker
func NewStreamWorker(workers *worker.Workers, id int, label string) *StreamWorker {
	return &StreamWorker{
		command: make(chan int),
		id:      id,
		label:   label,
		log:     log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers: workers,
	}
}

// Run is main function of this worker
func (w *StreamWorker) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	w.log.Infof("Worker Started")

	nw := w.findNetWorker()
	streamStartChannel := nw.Subscribe(w.id, protocolapp.OnStreamStartEvent, w.label).Channel
	streamStopChannel := nw.Subscribe(w.id, protocolapp.OnStreamStopEvent, w.label).Channel
	errorChannel := nw.Subscribe(w.id, protocolapp.OnErrorEvent, w.label).Channel

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case errorMessage := <-errorChannel:
			w.log.Debugf("Response: %s", string(errorMessage.([]byte)))

		case streamStart := <-streamStartChannel:

			c := protocolapp.NewOnStreamStart()
			err := json.Unmarshal(streamStart.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			if c.For == "" || c.For == "false" {
				c.For = "*UNKNOWN USER*"
			}

			w.log.Infof("Stream id %d Started - from '%s' on '%s' for '%s'", c.StreamID, c.From, c.Channel, c.For)

		case streamStop := <-streamStopChannel:
			c := protocolapp.NewOnStreamStop()
			err := json.Unmarshal(streamStop.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			w.log.Infof("Stream id %d Stopped", c.StreamID)

		case streamCommand, more := <-w.command:
			if more {
				w.log.Infof("Received command %d", streamCommand)
				switch streamCommand {
				case worker.Terminate:
					w.log.Infof("Terminating")
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

	nw.UnSubscribe(w.id, protocolapp.OnStreamStartEvent)
	nw.UnSubscribe(w.id, protocolapp.OnStreamStopEvent)
	nw.UnSubscribe(w.id, network.SubscriptionTypeConection)

	w.log.Info("Finished")
}

// Command sent to this worker
func (w *StreamWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *StreamWorker) Terminate() {
	w.Command(worker.Terminate)
}

// FindNetWorker find Net worker
func (w *StreamWorker) findNetWorker() *network.Networker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			return (*w.workers)[i].(*network.Networker)
		}
	}
	return nil
}
