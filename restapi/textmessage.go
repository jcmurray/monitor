// cSpell.language:en-GB
// cSpell:disable

package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/texts"
)

const ()

func (w *APIWorker) textMessage(resp http.ResponseWriter, req *http.Request) {
	w.log.Debugf("In: textMessage(resp http.ResponseWriter, req *http.Request")

	resp.Header().Set("Content-Type", "application/json")

	tmr := protocolapp.TextMessageRequest{}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&tmr); err != nil {
		w.log.Errorf("Unmarshal error: %s", err)

		resp.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(resp).Encode(protocolapp.TextMessageResponse{
			Error: "Invalid request",
		})
		return
	}
	defer req.Body.Close()

	message, _ := json.Marshal(protocolapp.InternalTextMessageRequest{
		For:     tmr.For,
		Message: tmr.Message,
	})

	tmw := w.findTextWorker()
	tmw.TextMessageEvent(message)

	resp.WriteHeader(http.StatusCreated)
	json.NewEncoder(resp).Encode(protocolapp.TextMessageResponse{
		Message: fmt.Sprintf("Sending message '%s' for user '%s'", tmr.Message, tmr.For),
	})
}

// FindNetWorker find Net worker
func (w *APIWorker) findTextWorker() *texts.TextMessageWorker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *texts.TextMessageWorker:
			return (*w.workers)[i].(*texts.TextMessageWorker)
		}
	}
	return nil
}
