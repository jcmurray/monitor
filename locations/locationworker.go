// cSpell.language:en-GB
// cSpell:disable

package locations

import (
	"encoding/json"
	"sync"

	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	w3w "github.com/jcmurray/what3words"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const ()

// LocationWorker stream worker
type LocationWorker struct {
	sync.Mutex
	command chan int
	log     *log.Entry
	id      int
	label   string
	workers *worker.Workers
}

// NewLocationWorker create a new LocationWorker
func NewLocationWorker(workers *worker.Workers, id int, label string) *LocationWorker {
	return &LocationWorker{
		command: make(chan int, 10),
		id:      id,
		label:   label,
		log:     log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers: workers,
	}
}

// Run is main function of this worker
func (w *LocationWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	nw := w.findNetWorker()
	locationChannel := nw.Subscribe(w.id, protocolapp.OnLocationEvent, w.label).Channel
	errorChannel := nw.Subscribe(w.id, protocolapp.OnErrorEvent, w.label).Channel

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case errorMessage := <-errorChannel:
			w.log.Debugf("Response: %s", string(errorMessage.([]byte)))

		case locationMessage := <-locationChannel:

			c := protocolapp.NewOnLocation()
			err := json.Unmarshal(locationMessage.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			w.log.Infof("Location message (ID: %d) on channel '%s' from '%s' for '%s'", c.MessageID, c.Channel, c.From, c.For)
			w.log.Infof("Location message (ID: %d) Lat: %.7f Lon: %.7f Accuracy: %.f m", c.MessageID, c.Latitude, c.Longitude, c.Accuracy)
			w.log.Infof("Location message (ID: %d) Address: %s", c.MessageID, c.FormattedAddress)

			var what3WordsAddress string

			if viper.GetBool("location.what3words") {
				what3WordsAddress, err = w.what3WordsFromLatLon(c.Latitude, c.Longitude)
				if err == nil {
					log.Infof("Location message (ID: %d) What3WordsAddress: ///%s", c.MessageID, what3WordsAddress)
				}
			}

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

	nw.UnSubscribe(w.id, protocolapp.OnLocationEvent)
	nw.UnSubscribe(w.id, protocolapp.OnErrorEvent)

	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *LocationWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *LocationWorker) Terminate() {
	w.Command(worker.Terminate)
}

// FindNetWorker find Net worker
func (w *LocationWorker) findNetWorker() *network.Networker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			return (*w.workers)[i].(*network.Networker)
		}
	}
	return nil
}

func (w *LocationWorker) what3WordsFromLatLon(lat float64, lon float64) (string, error) {
	api := w3w.NewGeocoder(viper.GetString("location.what3wordsapikey"))
	coords, err := w3w.NewCoordinates(lat, lon)
	if err != nil {
		return "", err
	}
	resp, err := api.ConvertTo3wa(coords)
	if err != nil {
		return "", err
	}
	return resp.Words, nil
}

// Label return label of worker
func (w *LocationWorker) Label() string {
	return w.label
}

// ID return label of worker
func (w *LocationWorker) ID() int {
	return w.id
}
