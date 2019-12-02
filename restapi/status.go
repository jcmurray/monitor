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
		switch t := (*w.workers)[i].(type) {
		case *network.Networker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *authenticate.AuthWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *channelstatus.StatusWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *streams.StreamWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *images.ImageWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *locations.LocationWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *texts.TextMessageWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *audiodecoder.AudioWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
		case *APIWorker:
			report = append(report, workerDetails{
				ID:            t.ID(),
				Name:          t.Label(),
				Subscriptions: t.Subscriptions()})
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

// Subscriptions return a copy of current scubscriptions
func (w *APIWorker) Subscriptions() []*worker.Subscription {
	return make([]*worker.Subscription, 0)
}
