// cSpell.language:en-GB
// cSpell:disable

package authenticate

import (
	"encoding/json"
	"sync"

	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const ()

// AuthWorker logon worker
type AuthWorker struct {
	sync.Mutex
	command      chan int
	log          *log.Entry
	id           int
	label        string
	workers      *worker.Workers
	loggedOn     bool
	refreshToken string
}

// NewAuthWorker create a new Logworker
func NewAuthWorker(workers *worker.Workers, id int, label string) *AuthWorker {
	return &AuthWorker{
		command:  make(chan int, 10),
		id:       id,
		label:    label,
		log:      log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers:  workers,
		loggedOn: false,
	}
}

// Run is main function of this worker
func (w *AuthWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	nw := w.findNetWorker()
	connectionChannel := nw.Subscribe(w.id, network.SubscriptionTypeConection, w.label).Channel
	responseChannel := nw.Subscribe(w.id, protocolapp.OnResponseEvent, w.label).Channel
	errorChannel := nw.Subscribe(w.id, protocolapp.OnErrorEvent, w.label).Channel

waitloop:
	for {
		w.log.Tracef("Entering Select")
		select {
		case connected := <-connectionChannel:
			w.log.Tracef("case connected := <-connectionChannel: %s", string(connected.([]byte)))
			cmd := string(connected.([]byte))
			switch cmd {
			case network.Connected:
				w.log.Debugf("Connected message received")
				if !w.isLoggedOn() {
					w.doLogon()
				}
			case network.Disconnected:
				w.log.Debugf("Disconnected message received")
				w.unsetLoggedOn()
			}

		case response := <-responseChannel:
			w.log.Tracef("Response: %s", string(response.([]byte)))
			resp := protocolapp.NewResponse()
			err := json.Unmarshal(response.([]byte), resp)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}
			if resp.Success && resp.Seq == 1 && resp.StreamID == 0 {
				w.refreshToken = resp.RefreshToken
				w.setLoggedOn()
				continue
			}
			if !resp.Success && resp.Seq == 1 && resp.Error == "invalid password" {
				w.log.Error("Logon failure - invalid password - requesting application termination")
				*term <- 1
				break waitloop
			}

			if !resp.Success && resp.Seq == 1 && resp.Error == "invalid username" {
				w.log.Error("Logon failure - invalid username - requesting application termination")
				*term <- 1
				break waitloop
			}

			if !resp.Success && resp.Seq == 1 && resp.Error == "not authorized" {
				w.log.Error("Logon failure - invalid authorisation token - requesting application termination")
				*term <- 1
				break waitloop
			}

		case errorMessage := <-errorChannel:
			w.log.Debugf("Error: %s", string(errorMessage.([]byte)))

		case logonCommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", logonCommand)
				switch logonCommand {
				case worker.Logon:
					if !w.isLoggedOn() {
						w.doLogon()
					}
				case worker.Logoff:
					w.unsetLoggedOn()
				case worker.Terminate:
					w.log.Debugf("Terminating")
					w.unsetLoggedOn()
					break waitloop
				default:
					continue
				}
			} else {
				w.log.Info("Channel closed")
				break waitloop
			}
			continue
		}
	}

	nw.UnSubscribe(w.id, protocolapp.OnResponseEvent)
	nw.UnSubscribe(w.id, protocolapp.OnErrorEvent)
	nw.UnSubscribe(w.id, network.SubscriptionTypeConection)

	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *AuthWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *AuthWorker) Terminate() {
	w.Command(worker.Terminate)
}

// FindNetWorker find Net worker
func (w *AuthWorker) findNetWorker() *network.Networker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			return (*w.workers)[i].(*network.Networker)
		}
	}
	return nil
}

func (w *AuthWorker) isLoggedOn() bool {
	return w.loggedOn
}

func (w *AuthWorker) setLoggedOn() {
	w.Lock()
	defer w.Unlock()
	w.loggedOn = true
}

func (w *AuthWorker) unsetLoggedOn() {
	w.Lock()
	defer w.Unlock()
	w.loggedOn = false
}

func (w *AuthWorker) doLogon() error {
	logon := protocolapp.NewLogon()
	logon.Channel = viper.GetString("logon.channel")

	if w.refreshToken != "" {
		logon.RefreshToken = w.refreshToken
	} else {
		logon.AuthToken = viper.GetString("logon.auth_token")
	}

	logon.Username = viper.GetString("logon.username")
	logon.Password = viper.GetString("logon.password")
	logon.ListenOnly = viper.GetBool("logon.listen_only")

	buff, err := json.Marshal(logon)
	if err != nil {
		w.log.Errorf("Marshal error: %s", err)
		return errors.Annotate(err, "Marshal failure for Zello logon request")
	}

	w.log.Tracef("Sending: %s", buff)

	nw := w.findNetWorker()
	nw.Data([]byte(string(buff)))
	return nil
}
