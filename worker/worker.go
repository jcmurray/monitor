// cSpell.language:en-GB
// cSpell:disable

package worker

import (
	"sync"
)

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
	ID() int
	Label() string
	Run(wg *sync.WaitGroup)
	Command(int)
	Subscribe(int, string) *Subscription
	UnSubscribe(int)
	Subscriptions() []*Subscription
}

// Workers object
type Workers []interface{}

// Subscription to a message on a channel
type Subscription struct {
	ID      int              `json:"id,omitempty"`
	Type    string           `json:"type,omitempty"`
	Label   string           `json:"label,omitempty"`
	Channel chan interface{} `json:"-"`
	Next    *Subscription    `json:"-"`
}

// NewSubscription create a new subscription
func NewSubscription(sID int, sType string, label string) *Subscription {
	return &Subscription{
		ID:      sID,
		Type:    sType,
		Label:   label,
		Channel: make(chan interface{}, 2),
		Next:    nil,
	}
}
