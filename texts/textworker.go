// cSpell.language:en-GB
// cSpell:disable

package texts

import (
	"encoding/json"
	"sync"

	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
)

const ()

// TextMessageWorker stream worker
type TextMessageWorker struct {
	sync.Mutex
	command chan int
	log     *log.Entry
	id      int
	label   string
	workers *worker.Workers
}

// NewTextMessageWorker create a new TextMessageWorker
func NewTextMessageWorker(workers *worker.Workers, id int, label string) *TextMessageWorker {
	return &TextMessageWorker{
		command: make(chan int),
		id:      id,
		label:   label,
		log:     log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers: workers,
	}
}

// Run is main function of this worker
func (w *TextMessageWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	nw := w.findNetWorker()
	textMessageChannel := nw.Subscribe(w.id, protocolapp.OnTextMessageEvent, w.label).Channel
	errorChannel := nw.Subscribe(w.id, protocolapp.OnErrorEvent, w.label).Channel

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case errorMessage := <-errorChannel:
			w.log.Debugf("Response: %s", string(errorMessage.([]byte)))

		case textMessage := <-textMessageChannel:

			c := protocolapp.NewOnTextMessage()
			err := json.Unmarshal(textMessage.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			w.log.Infof("Message id %d Started - from '%s' on '%s' for '%s': %s", c.MessageID, c.From, c.Channel, c.For, c.Text)

		case textMessageCommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", textMessageCommand)
				switch textMessageCommand {
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

	nw.UnSubscribe(w.id, protocolapp.OnTextMessageEvent)
	nw.UnSubscribe(w.id, protocolapp.OnErrorEvent)

	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *TextMessageWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *TextMessageWorker) Terminate() {
	w.Command(worker.Terminate)
}

// FindNetWorker find Net worker
func (w *TextMessageWorker) findNetWorker() *network.Networker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			return (*w.workers)[i].(*network.Networker)
		}
	}
	return nil
}
