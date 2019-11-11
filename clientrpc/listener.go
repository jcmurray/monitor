// cSpell.language:en-GB
// cSpell:disable

package clientrpc

import (
	context "context"
	"encoding/json"
	fmt "fmt"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jcmurray/monitor/audiodecoder"
	"github.com/jcmurray/monitor/authenticate"
	"github.com/jcmurray/monitor/channelstatus"
	"github.com/jcmurray/monitor/clientapi"
	"github.com/jcmurray/monitor/images"
	"github.com/jcmurray/monitor/locations"
	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/streams"
	"github.com/jcmurray/monitor/texts"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

const ()

// RPCWorker stream worker
type RPCWorker struct {
	sync.Mutex
	command           chan int
	log               *log.Entry
	id                int
	label             string
	workers           *worker.Workers
	subscriptionsLock sync.Mutex
	subscriptions     *worker.Subscription
}

// NewRPCWorker create a new RPCWorker
func NewRPCWorker(workers *worker.Workers, id int, label string) *RPCWorker {
	return &RPCWorker{
		command: make(chan int, 10),
		id:      id,
		label:   label,
		log:     log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers: workers,
	}
}

// Run is main function of this worker
func (w *RPCWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	var opts []grpc.ServerOption

	lis, err := net.Listen("tcp", "localhost:9998")
	if err != nil {
		w.log.Errorf("Listen on RPC address failed: %s", err)
		return
	}

	grpcServer := grpc.NewServer(opts...)
	clientapi.RegisterClientServiceServer(grpcServer, w)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			w.log.Errorf("RPC  gRPC failed to start: %s", err)
		}
	}()

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case rpcAPICommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", rpcAPICommand)
				switch rpcAPICommand {
				case worker.Terminate:
					w.log.Debugf("Terminating")
					_, cancel := context.WithTimeout(context.Background(), time.Second*10)
					defer cancel()
					grpcServer.GracefulStop()
					w.log.Infof("Shutting down grpc messaging server.")
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

// SendTextMessage rpc entry point
func (w *RPCWorker) SendTextMessage(ctx context.Context, t *clientapi.TextMessage) (*clientapi.TextMessageResponse, error) {
	w.log.Infof("in SendTextMessage")
	w.log.Infof("For=%s", t.For)
	w.log.Infof("Message=%s", t.Message)

	message, _ := json.Marshal(protocolapp.InternalTextMessageRequest{
		For:     t.For,
		Message: t.Message,
	})

	tmw := w.findTextWorker()
	tmw.TextMessageEvent(message)

	return &clientapi.TextMessageResponse{
		Success: true,
		Message: fmt.Sprintf("Text messaage for '%s' received: %s", t.For, t.Message),
	}, nil
}

// Status rpc entry point
func (w *RPCWorker) Status(ctx context.Context, t *empty.Empty) (*clientapi.StatusResponse, error) {

	var workers []*clientapi.WorkerDetails
	var subs []*clientapi.Subscription

	for i := range *w.workers {
		switch t := (*w.workers)[i].(type) {
		case *network.Networker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *authenticate.AuthWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *channelstatus.StatusWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *streams.StreamWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *images.ImageWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *locations.LocationWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *texts.TextMessageWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *audiodecoder.AudioWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		case *RPCWorker:
			for _, s := range t.Subscriptions() {
				subscription := &clientapi.Subscription{
					Id:    int32(s.ID),
					Type:  s.Type,
					Label: s.Label,
				}
				subs = append(subs, subscription)
			}
			workers = append(workers, &clientapi.WorkerDetails{
				Id:                 int32(t.ID()),
				Name:               t.Label(),
				WorkerSubscription: subs,
			})
		}
	}
	return &clientapi.StatusResponse{
		Workers: workers,
	}, nil
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

// findNetWorker find Net worker
func (w *RPCWorker) findTextWorker() *texts.TextMessageWorker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *texts.TextMessageWorker:
			return (*w.workers)[i].(*texts.TextMessageWorker)
		}
	}
	return nil
}
