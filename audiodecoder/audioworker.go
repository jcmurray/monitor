// cSpell.language:en-GB
// cSpell:disable

package audiodecoder

import (
	"sync"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/gordonklaus/portaudio"
	"github.com/hraban/opus"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	defaultQueueSize    = 20
	defaultQueueGetSize = 5
)

// AudioWorker stream worker
type AudioWorker struct {
	sync.Mutex
	command         chan int
	log             *log.Entry
	id              int
	label           string
	workers         *worker.Workers
	opusBufferQueue *queue.Queue
	enableAudio     bool
}

// NewAudioWorker create a new AudioWorker
func NewAudioWorker(workers *worker.Workers, id int, label string) *AudioWorker {
	return &AudioWorker{
		command:     make(chan int, 10),
		id:          id,
		label:       label,
		log:         log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers:     workers,
		enableAudio: viper.GetBool("audio.enable"),
	}
}

// Run is main function of this worker
func (w *AudioWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()

	w.log.Debugf("Worker Started")

	portaudio.Initialize()

	frameRate := viper.GetInt("audio.framerate")
	sampleRate := viper.GetInt("audio.samplerate")
	channels := viper.GetInt("audio.channels")
	framesPerPacket := viper.GetInt("audio.framesperpacket")
	bufferLength := frameRate * sampleRate * channels * framesPerPacket / 1000

	w.opusBufferQueue = queue.New(defaultQueueSize)
	dec, err := opus.NewDecoder(sampleRate, 1)
	if err != nil {
		w.log.Errorf("Error creating decoder: %s", err)
		return
	}

	pcm := make([]int16, int(bufferLength))

	w.log.Debugf("Audio loop setup calculated buffer length for PCM 'out' buffer: %d", bufferLength)

	if w.enableAudio {
		paStream, err := portaudio.OpenDefaultStream(0, 1, float64(sampleRate), bufferLength, func(out []int16) {
			if w.queueOccupancyWarning(w.opusBufferQueue.Len()) {
				w.log.Warnf("Buffer queue length %d occupancy %.2f%%", w.opusBufferQueue.Len(), w.queueOccupancyPercent(w.opusBufferQueue.Len()))
			}
			bufferList, err := w.opusBufferQueue.Get(defaultQueueGetSize)
			if err != nil {
				w.log.Errorf("PortAudio callback error getting buffer list from queue, %s", err)
				return
			}

			// bufferList is [][]interface{} te be interpretted as [][]byte

			for i := range bufferList {
				buffer := bufferList[i]
				n, err := dec.Decode(buffer.([]byte), pcm)
				if err != nil {
					log.Errorf("Error decoding: %s", err)
					continue
				}
				pcmBuffer := pcm[0:n]
				copy(out, pcmBuffer)
				w.log.Tracef("Audio loop, 'out' buffer length: %d, PCM buffer length: %d", len(out), len(pcmBuffer))
				w.log.Tracef("Start of Buffer sent to PortAudio: %#v ...", pcmBuffer[:10])
			}
		})

		if err != nil {
			w.log.Errorf("Failed to open PortAudio Stream, %s", err)
			return
		}
		w.log.Debug("PortAudio Stream opened")
		paStream.Start()
		w.log.Debug("PortAudio Stream started")
	} else {
		/*
		 * If audio to speaker is not enabled (PortAudio not available for example)
		 * this go routine discards audio buffers sent to us via the opusBufferQueue
		 */
		go func() {
			for {
				_, err := w.opusBufferQueue.Get(defaultQueueGetSize)
				if err != nil {
					w.log.Warnf("Audio buffer queue, %s", err)
					return
				}
				w.log.Tracef("Audio buffer queue entries discarded")
			}
		}()
		defer func() {
			_ = w.opusBufferQueue.Dispose()
		}()
	}

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {

		case audioCommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", audioCommand)
				switch audioCommand {
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
	w.log.Debug("Finished")
}

// Data sent to this worker
func (w *AudioWorker) Data(d []byte) {
	w.opusBufferQueue.Put(d)
}

// Command sent to this worker
func (w *AudioWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *AudioWorker) Terminate() {
	w.Command(worker.Terminate)
}

func (w *AudioWorker) queueOccupancyWarning(level int64) bool {
	return ((defaultQueueSize - level) < defaultQueueGetSize)
}

func (w *AudioWorker) queueOccupancyPercent(level int64) float64 {
	return (100.00 * (float64(level) / float64(defaultQueueSize)))
}
