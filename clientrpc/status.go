// cSpell.language:en-GB
// cSpell:disable

package clientrpc

import (
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jcmurray/monitor/audiodecoder"
	"github.com/jcmurray/monitor/authenticate"
	"github.com/jcmurray/monitor/channelstatus"
	"github.com/jcmurray/monitor/clientapi"
	"github.com/jcmurray/monitor/images"
	"github.com/jcmurray/monitor/locations"
	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/streams"
	"github.com/jcmurray/monitor/texts"
	"github.com/jcmurray/monitor/worker"
)

const ()

// Status rpc entry point
func (w *RPCWorker) Status(empty *empty.Empty, stream clientapi.ClientService_StatusServer) error {

	var detail *clientapi.WorkerDetails

	for i := range *w.workers {
		switch t := (*w.workers)[i].(type) {
		case *network.Networker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *authenticate.AuthWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *channelstatus.StatusWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *streams.StreamWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *images.ImageWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *locations.LocationWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *texts.TextMessageWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *audiodecoder.AudioWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}

		case *RPCWorker:
			var subs []*clientapi.Subscription
			for _, s := range t.Subscriptions() {
				subs = append(subs, &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				})
			}
			detail = &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			}
		}
		if err := stream.Send(detail); err != nil {
			return err
		}
	}
	return nil
}

// Command sent to this worker
func (w *RPCWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *RPCWorker) Terminate() {
	w.Command(worker.Terminate)
}

// Label return label of worker
func (w *RPCWorker) Label() string {
	return w.label
}

// ID return label of worker
func (w *RPCWorker) ID() int {
	return w.id
}

// Subscriptions return a copy of current scubscriptions
func (w *RPCWorker) Subscriptions() []*worker.Subscription {
	return make([]*worker.Subscription, 0)
}
