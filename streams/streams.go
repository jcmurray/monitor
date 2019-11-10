// cSpell.language:en-GB
// cSpell:disable

package streams

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"sync"

	"github.com/jcmurray/monitor/audiodecoder"
	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
)

const ()

type streamsInfo map[int]*StreamInfo

// StreamInfo persist stream information
type StreamInfo struct {
	Codec          string
	PacketDuration int
	StreamID       int
	Channel        string
	From           string
	CodecHeader    string
	For            string
	SampleRate     int
	FrameSizeMs    int
}

// StreamWorker stream worker
type StreamWorker struct {
	sync.Mutex
	command       chan int
	log           *log.Entry
	id            int
	label         string
	workers       *worker.Workers
	activeStreams streamsInfo
}

// NewStreamWorker create a new Streamworker
func NewStreamWorker(workers *worker.Workers, id int, label string) *StreamWorker {
	return &StreamWorker{
		command:       make(chan int, 10),
		id:            id,
		label:         label,
		log:           log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers:       workers,
		activeStreams: make(streamsInfo),
	}
}

// Run is main function of this worker
func (w *StreamWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	nw := w.findNetWorker()
	streamStartChannel := nw.Subscribe(w.id, protocolapp.OnStreamStartEvent, w.label).Channel
	streamStopChannel := nw.Subscribe(w.id, protocolapp.OnStreamStopEvent, w.label).Channel
	errorChannel := nw.Subscribe(w.id, protocolapp.OnErrorEvent, w.label).Channel
	streamDataChannel := nw.Subscribe(w.id, protocolapp.OnStreamDataEvent, w.label).Channel

waitloop:
	for {
		w.log.Tracef("Entering Select")
		select {
		case errorMessage := <-errorChannel:
			w.log.Debugf("Response: %s", string(errorMessage.([]byte)))

		case streamStart := <-streamStartChannel:

			c := protocolapp.NewOnStreamStart()
			err := json.Unmarshal(streamStart.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			codecHeader, _ := base64.StdEncoding.DecodeString(c.CodecHeader)

			delete(w.activeStreams, c.StreamID)
			w.activeStreams[c.StreamID] = &StreamInfo{
				StreamID:       c.StreamID,
				Channel:        c.Channel,
				From:           c.From,
				For:            c.For,
				Codec:          c.Codec,
				PacketDuration: c.PacketDuration,
				CodecHeader:    c.CodecHeader,
				SampleRate:     int(binary.LittleEndian.Uint16(codecHeader[0:2])),
				FrameSizeMs:    int(codecHeader[3]),
			}

			w.log.Infof("Stream id %d Started - from '%s' on '%s' for '%s'", c.StreamID, c.From, c.Channel, c.For)

		case streamStop := <-streamStopChannel:
			c := protocolapp.NewOnStreamStop()
			err := json.Unmarshal(streamStop.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			if si, ok := w.activeStreams[int(c.StreamID)]; ok {
				w.log.Infof("Stream id %d Stopped - from '%s' on '%s' for '%s'", c.StreamID, si.From, si.Channel, si.For)
				delete(w.activeStreams, c.StreamID)
			}
			continue

		case streamData := <-streamDataChannel:
			message := streamData.([]byte)
			streamID := binary.BigEndian.Uint32(message[1:5])
			packetID := binary.BigEndian.Uint32(message[5:9])
			data := message[9:]

			if _, ok := w.activeStreams[int(streamID)]; ok {
				w.log.Tracef("Start of Opus Packet %d received on stream ID %d: %#v ...", packetID, streamID, data[:6])
				au := w.findAudioWorker()
				au.Data(data)
			} else {
				w.log.Errorf("Unrecognised Audio StreamId %d", streamID)
			}
			continue

		case streamCommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", streamCommand)
				switch streamCommand {
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

	nw.UnSubscribe(w.id, protocolapp.OnStreamDataEvent)
	nw.UnSubscribe(w.id, protocolapp.OnStreamStartEvent)
	nw.UnSubscribe(w.id, protocolapp.OnStreamStopEvent)
	nw.UnSubscribe(w.id, network.SubscriptionTypeConection)

	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *StreamWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *StreamWorker) Terminate() {
	w.Command(worker.Terminate)
}

// FindNetWorker find Net worker
func (w *StreamWorker) findNetWorker() *network.Networker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			return (*w.workers)[i].(*network.Networker)
		}
	}
	return nil
}

// FindAudioWorker find Audio worker
func (w *StreamWorker) findAudioWorker() *audiodecoder.AudioWorker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *audiodecoder.AudioWorker:
			return (*w.workers)[i].(*audiodecoder.AudioWorker)
		}
	}
	return nil
}

// Label return label of worker
func (w *StreamWorker) Label() string {
	return w.label
}

// ID return label of worker
func (w *StreamWorker) ID() int {
	return w.id
}

// Subscriptions return a copy of current scubscriptions
func (w *StreamWorker) Subscriptions() []*worker.Subscription {
	return make([]*worker.Subscription, 0)
}
