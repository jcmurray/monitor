// cSpell.language:en-GB
// cSpell:disable

package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/jcmurray/monitor/worker"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	defaultRetryCount         = 8
	defaultRetryInterval      = 2
	defaultMaxRetryInterval   = 5 * 60
	oneDayInSeconds           = 24 * 60 * 60
	subscriptionTypeConection = "connection"
)

// Networker network worker
type Networker struct {
	sync.Mutex
	command        chan int
	data           chan []byte
	hostname       string
	port           int
	log            *log.Entry
	id             int
	label          string
	connected      bool
	webSocket      *websocket.Conn
	url            string
	retryCount     int
	retryInterval  time.Duration // seconds
	inRetryProcess bool
	thisInterval   time.Duration
	message        []byte
	subscriptions  []*worker.Subscription
	workers        *worker.Workers
}

// NewNetworker create a new Networker
func NewNetworker(workers *worker.Workers, id int, label string) *Networker {
	return &Networker{
		connected:      false,
		command:        make(chan int, 2),
		data:           make(chan []byte, 2),
		id:             id,
		label:          label,
		log:            log.WithFields(log.Fields{"Label": label, "ID": id}),
		inRetryProcess: false,
		retryCount:     defaultRetryCount,
		retryInterval:  defaultRetryInterval,
		workers:        workers,
	}
}

// Run is main finction of this worker
func (w *Networker) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if w.isConnected() {
			disconnect(w)
		}
	}()

	w.log.Infof("Worker Started")

	w.subscriptions = make([]*worker.Subscription, 0)

	w.hostname = viper.GetString("server.host")
	w.port = viper.GetInt("server.port")
	w.url = w.urlString(w.hostname, w.port)

	w.log.Infof("Hostname: %s, Port: %d", w.hostname, w.port)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	w.resetRetryCounters()

waitloop:
	for {
		var err error
		if w.isConnected() {
			_, w.message, err = w.webSocket.ReadMessage()
			if ce, ok := err.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived:
					w.setDisconnected()
				case websocket.CloseProtocolError,
					websocket.CloseUnsupportedData,
					websocket.CloseAbnormalClosure,
					websocket.CloseInvalidFramePayloadData,
					websocket.ClosePolicyViolation,
					websocket.CloseMessageTooBig,
					websocket.CloseMandatoryExtension,
					websocket.CloseInternalServerErr,
					websocket.CloseServiceRestart,
					websocket.CloseTryAgainLater,
					websocket.CloseTLSHandshake:
					w.setDisconnected()
					if !w.isRetrying() {
						w.resetRetryCounters()
						w.setRetrying()
					}
					w.thisInterval = w.getPollingInterval()
					w.log.Infof("Retry %d in %d seconds", w.retryCount, w.thisInterval)
				default:
					if w.isRetrying() {
						w.resetRetryCounters()
					}
				}
			}
		}
		if !w.isConnected() {
			w.log.Debugf("Entering Select")
			select {
			case netCommand, more := <-w.command:
				w.log.Debugf("case netCommand, more := <-w.command:")
				if more {
					w.log.Infof("Received command %d", netCommand)
					switch netCommand {
					case worker.Connect:
						w.log.Infof("Connecting")
						err := connect(w)
						if err != nil {
							w.log.Errorf("Connection error: %v", err)
							w.setDisconnected()
							if !w.isRetrying() {
								w.resetRetryCounters()
								w.setRetrying()
							} else {
								w.thisInterval = w.getPollingInterval()
							}
						}
					case worker.Disconnect:
						w.log.Infof("Disconnecting")
						err := disconnect(w)
						if err != nil {
							w.log.Errorf("Disconnection error: %v", err)
						}
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
			case <-ticker.C:
				w.log.Debugf("case <-ticker.C:")

				if w.isConnected() {
					w.log.Debug("Sending Ping")
					err := w.webSocket.WriteMessage(websocket.PingMessage, []byte(""))
					if err != nil {
						w.log.Errorf("Write error: %s", err)
					}
				}
				continue

			case <-time.After(time.Second * w.thisInterval):
				w.log.Debugf("case <-time.After(time.Second * w.thisInterval):")
				if w.isRetrying() {
					if err := connect(w); err != nil {
						w.log.Error("Connection failure")
						w.setDisconnected()
						if !w.isRetrying() {
							w.resetRetryCounters()
							w.setRetrying()
						} else {
							w.thisInterval = w.getPollingInterval()
						}
					}
				}
			}
		}

		if len(w.message) > 0 {
			if w.message[0] == 0x01 {
				w.log.Info("Data message")
			}
		}

		w.sendToSubscribersByType("logon_response", w.message)

	}
	w.log.Info("Finished")
}

// Command sent to this worker
func (w *Networker) Command(c int) {
	w.command <- c
}

// Data sent to this worker
func (w *Networker) Data(d []byte) {
	err := w.webSocket.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		w.log.Errorf("write: %s", err)
	}
	w.log.Debugf("writen: %v", d)
}

// Disconnect the websocket
func (w *Networker) Disconnect() {
	disconnect(w)
}

// Terminate the worker
func (w *Networker) Terminate() {
	disconnect(w)
	w.Command(worker.Terminate)
}

func connect(w *Networker) error {
	if w.isConnected() {
		w.log.Debugf("State error, unable to connect when already connected")
		return nil
	}
	w.log.Debugf("Connecting to: %s", w.url)
	c, _, err := websocket.DefaultDialer.Dial(w.url, nil)
	if err != nil {
		w.setDisconnected()
		return err
	}
	w.log.Debugf("Connected to %s", w.url)
	w.sendToSubscribersByType(subscriptionTypeConection, []byte("connected"))
	w.webSocket = c
	w.setConnected()
	return nil
}

func disconnect(w *Networker) error {
	if !w.isConnected() {
		w.log.Debugf("State error, unable to disconnect when already disconnected")
		return nil
	}
	w.log.Debugf("Disconnecting from: %s", w.url)

	err := w.webSocket.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		w.log.Errorf("Connection close error: %s", err)
		w.setDisconnected()
		return errors.Annotate(err, "Error on disconnecting Zello WebSocket")
	}
	w.sendToSubscribersByType(subscriptionTypeConection, []byte("disconnected"))
	w.setDisconnected()
	w.log.Debugf("Disconnected from %s", w.url)
	return nil
}

func (w *Networker) setConnected() {
	w.connected = true
}

func (w *Networker) setDisconnected() {
	w.connected = false
}

func (w *Networker) urlString(hostname string, port int) string {
	return fmt.Sprintf("wss://%s:%d/ws", hostname, port)
}

func (w *Networker) isRetrying() bool {
	return w.inRetryProcess
}

func (w *Networker) setRetrying() {
	w.inRetryProcess = true
	w.thisInterval = defaultRetryInterval
	w.log.Debugf("setRetrying(): setRetrying %v, inRetryProcess %d", w.inRetryProcess, w.thisInterval)
}

func (w *Networker) isConnected() bool {
	return w.connected
}

func (w *Networker) resetRetryCounters() {
	w.retryCount = defaultRetryCount
	w.retryInterval = defaultRetryInterval
	w.inRetryProcess = false
	w.thisInterval = oneDayInSeconds
	w.log.Debugf("resetRetryCounters(): w.retryCount = %d, retryInterval %d, inRetryProcess %v, thisInterval %d,",
		w.retryCount, w.retryInterval, w.inRetryProcess, w.thisInterval)
}

func (w *Networker) getPollingInterval() time.Duration {
	if w.retryCount > 0 {
		w.retryCount--
		w.retryInterval *= 2
	} else {
		w.retryInterval = defaultMaxRetryInterval
	}
	w.log.Debugf("getPollingInterval(): retryInterval %d", w.retryInterval)
	return w.retryInterval
}

// Subscribe to this worker
func (w *Networker) Subscribe(id int, sType string, label string) *worker.Subscription {
	w.Lock()
	defer w.Unlock()
	for _, s := range w.subscriptions {
		if s.ID == id && s.Type == sType {
			return s
		}
	}
	subscription := worker.NewSubscription(id, sType, label)

	spew.Dump(subscription)

	w.subscriptions = append(w.subscriptions, subscription)

	return subscription
}

// UnSubscribe from this worker
func (w *Networker) UnSubscribe(id int, sType string) {
	w.Lock()
	defer w.Unlock()
	for i, s := range w.subscriptions {
		if s.ID == id && s.Type == sType {
			w.subscriptions = append(w.subscriptions[:i], w.subscriptions[i+1:]...)
			return
		}
	}
}

func (w *Networker) sendToSubscribersByType(sType string, message []byte) {
	w.Lock()
	defer w.Unlock()
	w.log.Debugf("sendToSubscribersByType(): type %s, %v", sType, message)
	spew.Dump(w.subscriptions)
	for _, s := range w.subscriptions {
		if s.Type == sType {
			w.log.Debugf("sendToSubscribersByType(): Found channel match")
			s.Channel <- message
		}
	}
}
