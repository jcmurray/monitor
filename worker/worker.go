// cSpell.language:en-GB
// cSpell:disable

package worker

import "sync"

// Commands to a worker
const (
	Connect = iota
	Disconnect
	Terminate
	Logon
	Logoff
)

// IF Interface
type IF interface {
	ID() (int, error)
	Label() (string, error)
	Run(wg *sync.WaitGroup)
	Command(int)
	Subscribe(int, string) *Subscription
	UnSubscribe(int)
}

// Workers object
type Workers []interface{}

// Subscription to a messahe on a channel
type Subscription struct {
	ID      int
	Type    string
	Label   string
	Channel chan interface{}
}

// NewSubscription create a new subscription
func NewSubscription(sID int, sType string, label string) *Subscription {
	return &Subscription{
		ID:      sID,
		Type:    sType,
		Label:   label,
		Channel: make(chan interface{}, 2),
	}
}
