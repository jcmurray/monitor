// cSpell.language:en-GB
// cSpell:disable

package main

import (
	"math/rand"
	"time"

	"github.com/jcmurray/monitor/worker"
)

// cSpell.language:en-GB
// cSpell:disable

var (
	myrand *rand.Rand
)

func init() {
	myrand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// UtcTimeDate returns current time in UTC
func UtcTimeDate() time.Time {
	loc, _ := time.LoadLocation("UTC")
	return time.Now().In(loc)
}

// UnixMilli returns milliseconds from Unix epoch 0
func UnixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// NewID random ID that doesn't already exist
func NewID(ws worker.Workers) int {
	var proposedID int
check:
	for {
		proposedID = myrand.Intn(10000)
		for i := range ws {
			if ws[i] == proposedID {
				continue check
			}
		}
		return proposedID
	}
}

//
//// FindAuthWorker find Auth worker
//func FindAuthWorker() interface{} {
//	for i := range workers {
//		switch workers[i].(type) {
//		case *authenticate.AuthWorker:
//			return workers[i]
//		}
//	}
//	return nil
//}
//
