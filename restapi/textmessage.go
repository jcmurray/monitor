// cSpell.language:en-GB
// cSpell:disable

package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const ()

// TextMessageRequest comes via a POST request
type TextMessageRequest struct {
	For     string `json:"for,omitempty"`
	Message string `json:"message,omitempty"`
}

// TextMessageResponse to POST request
type TextMessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func (w *APIWorker) textMessage(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	tmr := TextMessageRequest{}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&tmr); err != nil {
		w.log.Errorf("Unmarshal error: %s", err)

		resp.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(resp).Encode(TextMessageResponse{
			Error: "Invalid request",
		})
		return
	}
	defer req.Body.Close()

	resp.WriteHeader(http.StatusCreated)
	json.NewEncoder(resp).Encode(TextMessageResponse{
		Message: fmt.Sprintf("Sending message '%s' for user '%s'", tmr.Message, tmr.For),
	})
}
