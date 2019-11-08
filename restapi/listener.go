// cSpell.language:en-GB
// cSpell:disable

package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/jcmurray/monitor/audiodecoder"
	"github.com/jcmurray/monitor/authenticate"
	"github.com/jcmurray/monitor/channelstatus"
	"github.com/jcmurray/monitor/images"
	"github.com/jcmurray/monitor/locations"
	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/streams"
	"github.com/jcmurray/monitor/texts"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const ()

// APIWorker stream worker
type APIWorker struct {
	sync.Mutex
	command chan int
	log     *log.Entry
	id      int
	label   string
	workers *worker.Workers
}

// NewAPIWorker create a new APIWorker
func NewAPIWorker(workers *worker.Workers, id int, label string) *APIWorker {
	return &APIWorker{
		command: make(chan int, 10),
		id:      id,
		label:   label,
		log:     log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers: workers,
	}
}

type workerDetails struct {
	ID            int                    `json:"id,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Subscriptions []*worker.Subscription `json:"subscriptions,omitempty"`
}

type workerReport []workerDetails

func (w *APIWorker) status(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
	report := workerReport{}

	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			networker := (*w.workers)[i].(*network.Networker)
			report = append(report, workerDetails{
				ID:            networker.ID(),
				Name:          networker.Label(),
				Subscriptions: networker.Subscriptions()})
		case *authenticate.AuthWorker:
			authworker := (*w.workers)[i].(*authenticate.AuthWorker)
			report = append(report, workerDetails{
				ID:   authworker.ID(),
				Name: authworker.Label()})
		case *channelstatus.StatusWorker:
			statusworker := (*w.workers)[i].(*channelstatus.StatusWorker)
			report = append(report, workerDetails{ID: statusworker.ID(), Name: statusworker.Label()})
		case *streams.StreamWorker:
			streamworker := (*w.workers)[i].(*streams.StreamWorker)
			report = append(report, workerDetails{ID: streamworker.ID(), Name: streamworker.Label()})
		case *images.ImageWorker:
			imageworker := (*w.workers)[i].(*images.ImageWorker)
			report = append(report, workerDetails{ID: imageworker.ID(), Name: imageworker.Label()})
		case *locations.LocationWorker:
			locationworker := (*w.workers)[i].(*locations.LocationWorker)
			report = append(report, workerDetails{ID: locationworker.ID(), Name: locationworker.Label()})
		case *texts.TextMessageWorker:
			textworker := (*w.workers)[i].(*texts.TextMessageWorker)
			report = append(report, workerDetails{ID: textworker.ID(), Name: textworker.Label()})
		case *audiodecoder.AudioWorker:
			audioworker := (*w.workers)[i].(*audiodecoder.AudioWorker)
			report = append(report, workerDetails{ID: audioworker.ID(), Name: audioworker.Label()})
		case *APIWorker:
			restapiworker := (*w.workers)[i].(*APIWorker)
			report = append(report, workerDetails{ID: restapiworker.ID(), Name: restapiworker.Label()})
		}
	}
	json.NewEncoder(resp).Encode(report)
}

// Run is main function of this worker
func (w *APIWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/status", w.status)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("rest.apiport")),
		Handler: myRouter,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			w.log.Errorf("HTTP ListenAndServe failed: %s", err)
		}
	}()

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case restAPICommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", restAPICommand)
				switch restAPICommand {
				case worker.Terminate:
					w.log.Debugf("Terminating")

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := server.Shutdown(ctx); err != nil {
						w.log.Errorf("HTTP Shutdown failed: %s", err)
					}
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
	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *APIWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *APIWorker) Terminate() {
	w.Command(worker.Terminate)
}

// Label return label of worker
func (w *APIWorker) Label() string {
	return w.label
}

// ID return label of worker
func (w *APIWorker) ID() int {
	return w.id
}
