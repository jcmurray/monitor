// cSpell.language:en-GB
// cSpell:disable

package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/jcmurray/monitor/audiodecoder"
	"github.com/jcmurray/monitor/authenticate"
	"github.com/jcmurray/monitor/channelstatus"
	"github.com/jcmurray/monitor/images"
	"github.com/jcmurray/monitor/locations"
	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/streams"
	"github.com/jcmurray/monitor/texts"
	"github.com/jcmurray/monitor/worker"
)

const ()

type workerDetails struct {
	ID            int                    `json:"id,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Subscriptions []*worker.Subscription `json:"subscriptions,omitempty"`
}

type workerReport []workerDetails

func (w *APIWorker) status(resp http.ResponseWriter, req *http.Request) {

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
			report = append(report, workerDetails{
				ID:   statusworker.ID(),
				Name: statusworker.Label()})
		case *streams.StreamWorker:
			streamworker := (*w.workers)[i].(*streams.StreamWorker)
			report = append(report, workerDetails{
				ID:   streamworker.ID(),
				Name: streamworker.Label()})
		case *images.ImageWorker:
			imageworker := (*w.workers)[i].(*images.ImageWorker)
			report = append(report, workerDetails{
				ID:   imageworker.ID(),
				Name: imageworker.Label()})
		case *locations.LocationWorker:
			locationworker := (*w.workers)[i].(*locations.LocationWorker)
			report = append(report, workerDetails{
				ID:   locationworker.ID(),
				Name: locationworker.Label()})
		case *texts.TextMessageWorker:
			textworker := (*w.workers)[i].(*texts.TextMessageWorker)
			report = append(report, workerDetails{
				ID:   textworker.ID(),
				Name: textworker.Label()})
		case *audiodecoder.AudioWorker:
			audioworker := (*w.workers)[i].(*audiodecoder.AudioWorker)
			report = append(report, workerDetails{
				ID:   audioworker.ID(),
				Name: audioworker.Label()})
		case *APIWorker:
			restapiworker := (*w.workers)[i].(*APIWorker)
			report = append(report, workerDetails{
				ID:   restapiworker.ID(),
				Name: restapiworker.Label()})
		}
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	json.NewEncoder(resp).Encode(report)
}

// Label return label of worker
func (w *APIWorker) Label() string {
	return w.label
}

// ID return label of worker
func (w *APIWorker) ID() int {
	return w.id
}
