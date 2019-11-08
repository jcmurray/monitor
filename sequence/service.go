// cSpell.language:en-GB
// cSpell:disable

package sequence

import "sync"

const ()

type expectedResponseEntry struct {
	id       int
	sequence int
	command  string
}

var (
	sequenceLock         sync.Mutex
	nextSequenceNumber   int
	expectedResponseList []expectedResponseEntry
)

func init() {
	nextSequenceNumber = 1
}

// GetNextSequenceNumber returns next sequence number
func GetNextSequenceNumber(id int, command string) int {
	sequenceLock.Lock()
	defer sequenceLock.Unlock()
	sequenceNumber := nextSequenceNumber
	expectedResponseList = append(expectedResponseList,
		expectedResponseEntry{
			id:       id,
			sequence: sequenceNumber,
			command:  command,
		})
	nextSequenceNumber++
	return sequenceNumber
}

// IsSeqExpected checks for a response for specific sequence number being expected
func IsSeqExpected(id int, seq int) bool {
	sequenceLock.Lock()
	defer sequenceLock.Unlock()
	for i := range expectedResponseList {
		if (expectedResponseList[i].id == id) && (expectedResponseList[i].sequence == seq) {
			return true
		}
	}
	return false
}

// IsCommandExpected checks for a response for specific command being expected
func IsCommandExpected(id int, command string) bool {
	sequenceLock.Lock()
	defer sequenceLock.Unlock()
	for i := range expectedResponseList {
		if (expectedResponseList[i].id == id) &&
			(expectedResponseList[i].command == command) {
			return true
		}
	}
	return false
}

// IsCommandSeqExpected checks for a response for specific command and Sequence being expected
func IsCommandSeqExpected(id int, seq int, command string) bool {
	sequenceLock.Lock()
	defer sequenceLock.Unlock()
	for i := range expectedResponseList {
		if (expectedResponseList[i].id == id) &&
			(expectedResponseList[i].sequence == seq) &&
			(expectedResponseList[i].command == command) {
			return true
		}
	}
	return false
}

// IsAnyExpected checks for any response being expected
func IsAnyExpected(id int) bool {
	sequenceLock.Lock()
	defer sequenceLock.Unlock()
	for i := range expectedResponseList {
		if expectedResponseList[i].id == id {
			return true
		}
	}
	return false
}

// RemoveEntry checks for a response sequence number being expected
func RemoveEntry(id int, seq int) {
	sequenceLock.Lock()
	defer sequenceLock.Unlock()
	for i := range expectedResponseList {
		if (expectedResponseList[i].id == id) && (expectedResponseList[i].sequence == seq) {
			expectedResponseList = append(expectedResponseList[0:i], expectedResponseList[i+1:]...)
			return
		}
	}
}
