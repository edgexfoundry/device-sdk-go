package models

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Event struct {
	contract.Event
}

// HasBinaryValue confirms whether an event contains one or more
// readings populated with a BinaryValue payload.
func (e Event) HasBinaryValue() bool {
	if len(e.Readings) > 0 {
		for r := range e.Readings {
			if len(e.Readings[r].BinaryValue) > 0 {
				return true
			}
		}
	}
	return false
}
