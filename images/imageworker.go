// cSpell.language:en-GB
// cSpell:disable

package images

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/jcmurray/monitor/network"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const ()

type imagesInfo map[int]*ImageInfo

// ImageInfo persists image information
type ImageInfo struct {
	Channel           string
	From              string
	For               string
	MessageID         int
	Type              string
	Height            int
	Width             int
	Source            string
	thumbnailReceived bool
	fullImageReceived bool
}

// ImageWorker stream worker
type ImageWorker struct {
	sync.Mutex
	command      chan int
	log          *log.Entry
	id           int
	label        string
	workers      *worker.Workers
	activeImages imagesInfo
}

// NewImageWorker create a new ImageWorker
func NewImageWorker(workers *worker.Workers, id int, label string) *ImageWorker {
	return &ImageWorker{
		command:      make(chan int, 10),
		id:           id,
		label:        label,
		log:          log.WithFields(log.Fields{"Label": label, "ID": id}),
		workers:      workers,
		activeImages: make(imagesInfo),
	}
}

// Run is main function of this worker
func (w *ImageWorker) Run(wg *sync.WaitGroup, term *chan int) {
	defer wg.Done()
	w.log.Debugf("Worker Started")

	nw := w.findNetWorker()
	imageChannel := nw.Subscribe(w.id, protocolapp.OnImageEvent, w.label).Channel
	errorChannel := nw.Subscribe(w.id, protocolapp.OnErrorEvent, w.label).Channel
	imageDataChannel := nw.Subscribe(w.id, protocolapp.OnImageDataEvent, w.label).Channel

waitloop:
	for {
		w.log.Debugf("Entering Select")
		select {
		case errorMessage := <-errorChannel:
			w.log.Debugf("Response: %s", string(errorMessage.([]byte)))

		case imageMessage := <-imageChannel:

			c := protocolapp.NewOnImage()
			err := json.Unmarshal(imageMessage.([]byte), c)
			if err != nil {
				w.log.Errorf("Unmarshal error: %s", err)
				continue
			}

			delete(w.activeImages, c.MessageID)
			w.activeImages[c.MessageID] = &ImageInfo{
				Channel:           c.Channel,
				From:              c.From,
				For:               c.For,
				MessageID:         c.MessageID,
				Type:              c.Type,
				Height:            c.Height,
				Width:             c.Width,
				Source:            c.Source,
				thumbnailReceived: false,
				fullImageReceived: false,
			}

			w.log.Infof("Message id %d Started - from '%s' on '%s' for '%s'", c.MessageID, c.From, c.Channel, c.For)

		case imageData := <-imageDataChannel:
			message := imageData.([]byte)
			messageID := binary.BigEndian.Uint32(message[1:5])
			imageType := binary.BigEndian.Uint32(message[5:9])
			data := message[9:]

			if ai, ok := w.activeImages[int(messageID)]; ok {
				switch imageType {
				case 1:
					w.log.Debugf("Full Image received on message ID %d", messageID)
					w.saveFullImageFile(ai, data)
					ai.fullImageReceived = true
				case 2:
					w.log.Debugf("Full Image received on message ID %d", messageID)
					w.saveThumbNailImageFile(ai, data)
					ai.thumbnailReceived = true
				default:
					w.log.Errorf("Unrecognised image packet type: %d, for message ID: %d", imageType, messageID)
				}
				if ai.fullImageReceived && ai.thumbnailReceived {
					delete(w.activeImages, int(messageID))
				}
			} else {
				w.log.Errorf("Unrecognised Image Mesage ID %d", messageID)
			}
			continue

		case imageCommand, more := <-w.command:
			if more {
				w.log.Debugf("Received command %d", imageCommand)
				switch imageCommand {
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

	nw.UnSubscribe(w.id, protocolapp.OnImageDataEvent)
	nw.UnSubscribe(w.id, protocolapp.OnImageEvent)
	nw.UnSubscribe(w.id, protocolapp.OnErrorEvent)

	w.log.Debug("Finished")
}

// Command sent to this worker
func (w *ImageWorker) Command(c int) {
	w.command <- c
}

// Terminate the worker
func (w *ImageWorker) Terminate() {
	w.Command(worker.Terminate)
}

// FindNetWorker find Net worker
func (w *ImageWorker) findNetWorker() *network.Networker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *network.Networker:
			return (*w.workers)[i].(*network.Networker)
		}
	}
	return nil
}

func (w *ImageWorker) saveFullImageFile(ai *ImageInfo, data []byte) {
	if viper.GetBool("image.logging") {
		fileNameFull := fmt.Sprintf("image_full_%d_%s.%s", ai.MessageID, ai.Source, ai.Type)
		w.log.Infof("Logging full      image on message ID %d to file: %s", ai.MessageID, fileNameFull)
		w.saveImageFile(ai, fileNameFull, data)
	}
}

func (w *ImageWorker) saveThumbNailImageFile(ai *ImageInfo, data []byte) {
	if viper.GetBool("image.logging") {
		fileNameThumb := fmt.Sprintf("image_thumb_%d_%s.%s", ai.MessageID, ai.Source, ai.Type)
		w.log.Infof("Logging thumbnail image on message ID %d to file: %s", ai.MessageID, fileNameThumb)
		w.saveImageFile(ai, fileNameThumb, data)
	}
}

func (w *ImageWorker) saveImageFile(ai *ImageInfo, fileName string, data []byte) {
	f, err := os.Create(fileName)
	if err != nil {
		w.log.Errorf("Image file creation error: %s", err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		w.log.Errorf("Error writing image file for message ID %d, %v", ai.MessageID, err)
	}
}

// Label return label of worker
func (w *ImageWorker) Label() string {
	return w.label
}

// ID return label of worker
func (w *ImageWorker) ID() int {
	return w.id
}

// Subscriptions return a copy of current scubscriptions
func (w *ImageWorker) Subscriptions() []*worker.Subscription {
	return make([]*worker.Subscription, 0)
}
