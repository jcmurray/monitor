// cSpell.language:en-GB
// cSpell:disable

package clientrpc

import (
	context "context"
	"net"
	"sync"
	"time"

	"github.com/jcmurray/monitor/clientapi"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

const ()

// RPCWorker stream worker
type RPCWorker struct {
	clientapi.UnimplementedClientServiceServer
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
