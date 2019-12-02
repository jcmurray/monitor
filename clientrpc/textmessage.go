// cSpell.language:en-GB
// cSpell:disable

package clientrpc

import (
	context "context"
	"encoding/json"
	fmt "fmt"

	"github.com/jcmurray/monitor/clientapi"
	"github.com/jcmurray/monitor/protocolapp"
	"github.com/jcmurray/monitor/texts"
)

const ()

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

// findTextWorker find Text worker
func (w *RPCWorker) findTextWorker() *texts.TextMessageWorker {
	for i := range *w.workers {
		switch (*w.workers)[i].(type) {
		case *texts.TextMessageWorker:
			return (*w.workers)[i].(*texts.TextMessageWorker)
		}
	}
	return nil
}
