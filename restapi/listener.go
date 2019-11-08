// cSpell.language:en-GB
// cSpell:disable

package restapi

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
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

// Run is main function of this worker
func (w *APIWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/status", w.status)
	myRouter.HandleFunc("/textmessage", w.textMessage).
		Methods("POST")

	server := &http.Server{
		Handler:      myRouter,
		Addr:         fmt.Sprintf(":%d", viper.GetInt("rest.apiport")),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
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
