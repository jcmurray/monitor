// cSpell.language:en-GB
// cSpell:disable

package authenticate

import (
	"encoding/json"
	"sync"

	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/network/protocol"
	"github.com/jcmurray/monitor/worker"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const ()

// AuthWorker network worker
type AuthWorker struct {
	command chan int
	log     *log.Entry
	id      int
	label   string
	workers *worker.Workers
}

// NewAuthworker create a new Logworker
func NewAuthworker(workers *worker.Workers, id int, label string) *AuthWorker {
	return &AuthWorker{
		command: make(chan int),
		id:      id,
		label:   label,
		log:     log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers: workers,
	}
}

// Run is main function of this worker
func (w *AuthWorker) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	w.log.Infof("Worker Started")

	nw := w.findNetWorker()
	con := nw.Subscribe(w.id, "connection", w.label)
	connectionChannel := con.Channel
	logon := nw.Subscribe(w.id, "logon_response", w.label)
	logonChannel := logon.Channel

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case connected := <-connectionChannel:
			w.log.Debugf("Connected message received: v", connected)
			w.doLogon()

		case logon := <-logonChannel:
			w.log.Debugf("Response: %s", string(logon.([]byte)))

		case logonCommand, more := <-w.command:
			w.log.Debugf("case logonCommand, more := <-w.command:")
			if more {
				w.log.Infof("Received command %d", logonCommand)
				switch logonCommand {
				case worker.Logon:
				case worker.Logoff:
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
			continue
		}

	}
	w.log.Info("Finished")
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

func (w *AuthWorker) doLogon() error {
	logon := protocol.NewLogon()
	logon.Channel = viper.GetString("logon.channel")
	logon.AuthToken = viper.GetString("logon.auth_token")
	logon.Username = viper.GetString("logon.username")
	logon.Password = viper.GetString("logon.password")
	logon.ListenOnly = viper.GetBool("logon.listen_only")

	buff, err := json.Marshal(&logon)
	if err != nil {
		w.log.Errorf("Marshal error: %s", err)
		return errors.Annotate(err, "Marshal failure for Zello logon request")
	}

	w.log.Debugf("Sending: %s", buff)

	nw := w.findNetWorker()
	nw.Data([]byte(string(buff)))
	return nil
}
