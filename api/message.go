// Package api defines interfaces and schemas
package api

import (
	"time"
)

// RequestMessage is the incoming message
type RequestMessage struct {
	From string `json:"from"`
	Text string `json:"text"`
}

// ResponseMessage is the outgoing message
type ResponseMessage struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

// ResponseWithDelay contains the artificial delay, too
type ResponseWithDelay struct {
	Response ResponseMessage
	Delay    time.Duration
}
