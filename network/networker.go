// cSpell.language:en-GB
// cSpell:disable

package network

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// SubscriptionTypeConection for use by other go routines
const (
	defaultRetryCount         = 8
	defaultRetryInterval      = 2
	defaultMaxRetryInterval   = 5 * 60
	oneDayInSeconds           = 24 * 60 * 60
	SubscriptionTypeConection = "connection"
	Connected                 = "connected"
	Disconnected              = "disconnected"
)

// Networker network worker
type Networker struct {
	sync.Mutex
	command           chan int
	hostname          string
	port              int
	log               *log.Entry
	id                int
	label             string
	connected         bool
	webSocket         *websocket.Conn
	url               string
	retryCount        int
	retryInterval     time.Duration // seconds
	inRetryProcess    bool
	thisInterval      time.Duration
	message           []byte
	subscriptionsLock sync.Mutex
	subscriptions     *worker.Subscription
	workers           *worker.Workers
}

// NewNetworker create a new Networker
func NewNetworker(workers *worker.Workers, id int, label string) *Networker {
	return &Networker{
		connected:      false,
		command:        make(chan int, 10),
		id:             id,
		label:          label,
		log:            log.WithFields(log.Fields{"Label": label, "ID": id}),
		inRetryProcess: false,
		retryCount:     defaultRetryCount,
		retryInterval:  defaultRetryInterval,
		workers:        workers,
		subscriptions:  nil,
	}
}

// Run is main finction of this worker
func (w *Networker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	defer func() {
		if w.isConnected() {
			disconnect(w)
		}
	}()

	w.log.Debugf("Worker Started")

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
			w.log.Trace("About to: w.webSocket.ReadMessage()")
			_, w.message, err = w.webSocket.ReadMessage()
			if ce, ok := err.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived:
					w.log.Trace("WebSocket Normal Close")
					w.setDisconnected()
					break waitloop
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
					w.log.Trace("WebSocket Abnormal Close")
					w.setDisconnected()
					if !w.isRetrying() {
						w.resetRetryCounters()
						w.setRetrying()
					}
					w.thisInterval = w.getPollingInterval()
					w.log.Infof("Retry %d in %d seconds", w.retryCount, w.thisInterval)

				case 3001:
					w.log.Trace("Exiting due to channel closed")
					w.setDisconnected()
					w.Command(worker.Terminate)
					*term <- 1

				case 3002:
					w.log.Trace("Exiting due to invalid credentials")
					w.setDisconnected()
					w.Command(worker.Terminate)
					*term <- 1

				default:
					w.log.Tracef("Exiting due to unrecognised websocket close code=%s", err)
					w.setDisconnected()
					w.Command(worker.Terminate)
					*term <- 1
				}
			} else {
				w.log.Tracef("WebSocket exited with non-close indication, err=%s", err)
			}
		} else {
			w.log.Debugf("Entering Select")
			select {
			case netCommand, more := <-w.command:
				w.log.Tracef("netCommand, more := <-w.command:")
				if more {
					w.log.Debugf("Received command %d", netCommand)
					switch netCommand {
					case worker.Connect:
						w.log.Debugf("Connecting")
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
						w.log.Debugf("Connected")
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
			case <-ticker.C:
				w.log.Tracef("case <-ticker.C:")
				if w.isConnected() {
					w.log.Debug("Sending Ping")
					err := w.webSocket.WriteMessage(websocket.PingMessage, []byte(""))
					if err != nil {
						w.log.Errorf("Write error: %s", err)
					}
				}
				continue

			case <-time.After(time.Second * w.thisInterval):
				w.log.Tracef("time.Second * w.thisInterval")
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

		if len(w.message) == 0 {
			continue
		}

		if w.message[0] == 0x01 {
			w.log.Trace("Data message")
			w.sendToSubscribersByType(protocolapp.OnStreamDataEvent, w.message)
			continue
		}

		if w.message[0] == 0x02 {
			w.log.Trace("Image message")
			w.sendToSubscribersByType(protocolapp.OnImageDataEvent, w.message)
			continue
		}

		command := protocolapp.Command{}
		err = json.Unmarshal(w.message, &command)
		if err != nil {
			w.log.Errorf("Unmarshal error: %s", err)
			continue
		}

		switch {
		case command.Command == protocolapp.OnErrorEvent:
			w.log.Debugf("Command received: %s", command.Command)
			w.log.Tracef("Error Message: %s", string(w.message))
			w.sendToSubscribersByType(protocolapp.OnErrorEvent, w.message)
			continue
		case command.Command == protocolapp.OnChannelStatusEvent:
			w.log.Debugf("Command received: %s", command.Command)
			w.log.Tracef("Status Message: %s", string(w.message))
			w.sendToSubscribersByType(protocolapp.OnChannelStatusEvent, w.message)
			continue
		case command.Command == protocolapp.OnStreamStartEvent:
			w.log.Debugf("Command received: %s", command.Command)
			w.log.Tracef("Stream Start Message: %s", string(w.message))
			w.sendToSubscribersByType(protocolapp.OnStreamStartEvent, w.message)
			continue
		case command.Command == protocolapp.OnStreamStopEvent:
			w.log.Debugf("Command received: %s", command.Command)
			w.log.Tracef("Stream Stop Message: %s", string(w.message))
			w.sendToSubscribersByType(protocolapp.OnStreamStopEvent, w.message)
			continue
		case command.Command == protocolapp.OnImageEvent:
			w.log.Debugf("Command received: %s", command.Command)
			w.log.Tracef("Image Message: %s", string(w.message))
			w.sendToSubscribersByType(protocolapp.OnImageEvent, w.message)
			continue
		case command.Command == protocolapp.OnTextMessageEvent:
			w.log.Debugf("Command received: %s", command.Command)
			w.log.Tracef("Text Message: %s", string(w.message))
			w.sendToSubscribersByType(protocolapp.OnTextMessageEvent, w.message)
			continue
		case command.Command == protocolapp.OnLocationEvent:
			w.log.Debugf("Command received: %s", command.Command)
			w.log.Tracef("Location Message: %s", string(w.message))
			w.sendToSubscribersByType(protocolapp.OnLocationEvent, w.message)
			continue
		}

		w.sendToSubscribersByType(protocolapp.OnResponseEvent, w.message)

	}
	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *Networker) Command(c int) {
	w.command <- c
}

// Data sent to this worker
func (w *Networker) Data(d []byte) {
	if w.isConnected() {
		err := w.webSocket.WriteMessage(websocket.TextMessage, d)
		if err != nil {
			w.log.Errorf("write: %s", err)
		}
		w.log.Tracef("writen: %v", d)
	} else {
		w.log.Warn("Attempt to send data on disconnected websocket")
	}
}

// Disconnect the websocket
func (w *Networker) Disconnect() {
	w.log.Trace("in: func (w *Networker) Disconnect()")
	disconnect(w)
}

// Terminate the worker
func (w *Networker) Terminate() {
	w.log.Trace("in: func (w *Networker) Terminate()")
	w.Disconnect()
	w.Command(worker.Terminate)
}

func connect(w *Networker) error {
	w.Lock()
	defer w.Unlock()

	w.log.Trace("in: func (w *Networker) connect()")

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
	w.sendToSubscribersByType(SubscriptionTypeConection, []byte(Connected))
	w.webSocket = c
	w.setConnected()
	return nil
}

func disconnect(w *Networker) error {
	w.Lock()
	defer w.Unlock()

	w.log.Trace("in: func (w *Networker) disconnect()")

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
	w.sendToSubscribersByType(SubscriptionTypeConection, []byte(Disconnected))
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

	subscription := worker.NewSubscription(id, sType, label)
	subscription.Next = w.subscriptions
	w.subscriptions = subscription

	return subscription
}

// UnSubscribe from this worker
func (w *Networker) UnSubscribe(id int, sType string) {
	w.subscriptionsLock.Lock()
	defer w.subscriptionsLock.Unlock()

	var located *worker.Subscription = nil
	var prior *worker.Subscription = nil

	for s := w.subscriptions; s != nil; s = s.Next {
		if s.ID == id && s.Type == sType {
			located = s
			break
		}
	}

	if located == nil {
		return
	}

	if located == w.subscriptions {
		w.subscriptions = located.Next
		return
	}

	for s := w.subscriptions; s != nil; s = s.Next {
		if s.Next == located {
			prior = s
			break
		}
	}

	prior.Next = located.Next
}

func (w *Networker) sendToSubscribersByType(sType string, message []byte) {
	w.subscriptionsLock.Lock()
	defer w.subscriptionsLock.Unlock()
	w.log.Tracef("sendToSubscribersByType(): type %s, %v", sType, message)
	for s := w.subscriptions; s != nil; s = s.Next {
		if s.Type == sType {
			s.Channel <- message
			return
		}
	}
}
