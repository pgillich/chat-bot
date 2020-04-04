// Package test is a helper for automatic tests
package test

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pgillich/chat-bot/api"
	"github.com/pgillich/chat-bot/config"
	"github.com/pgillich/chat-bot/internal/db"
)

const defaultLogLevel = log.WarnLevel

// MessagePair contains the incoming message and the expected response
type MessagePair struct {
	In        string
	User      db.User
	Responses []api.ResponseWithDelay
}

// GetLogLevel returns the default log level for tests
func GetLogLevel() string {
	return defaultLogLevel.String()
}

// GetMessagePairsJohnDoe returns conversation of John Doe
func GetMessagePairsJohnDoe(to string) []MessagePair { // nolint:dupl
	bornOnText := "1976.04.24."
	bornOn, _ := time.Parse("2006.01.02.", bornOnText) // nolint:errcheck

	return []MessagePair{
		{"Hello", db.User{UID: to},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "Hi"}, Delay: config.DefaultDelay},
				{Response: api.ResponseMessage{To: to, Text: "What's your name?"}, Delay: config.DefaultDelay},
			},
		},
		{"John Doe", db.User{UID: to, Name: "John Doe"},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "When were you born?"}, Delay: config.DefaultDelay},
			},
		},
		{bornOnText, db.User{UID: to, Name: "John Doe", BornOn: &bornOn},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "Where were you born?"}, Delay: config.DefaultDelay},
			},
		},
		{"Mucsajröcsöge", db.User{UID: to, Name: "John Doe", BornOn: &bornOn, BornAt: "Mucsajröcsöge"},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "Hello John Doe from Mucsajröcsöge!"}, Delay: config.DefaultDelay},
				// nolint:misspell,lll
				{Response: api.ResponseMessage{To: to, Text: "You are 43 years old. Hey, that's still younger than the Millenium Falcon!"}, Delay: 0},
			},
		},
	}
}

// GetMessagePairsJaneDoe returns conversation of Jane Doe
func GetMessagePairsJaneDoe(to string) []MessagePair { // nolint:dupl
	bornOnText := "1994.12.04."
	bornOn, _ := time.Parse("2006.01.02.", bornOnText) // nolint:errcheck

	return []MessagePair{
		{"Hi.", db.User{UID: to},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "Hi"}, Delay: config.DefaultDelay},
				{Response: api.ResponseMessage{To: to, Text: "What's your name?"}, Delay: config.DefaultDelay},
			},
		},
		{"My name is Jane Doe.", db.User{UID: to, Name: "Jane Doe"},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "When were you born?"}, Delay: config.DefaultDelay},
			},
		},
		{bornOnText, db.User{UID: to, Name: "Jane Doe", BornOn: &bornOn},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "Where were you born?"}, Delay: config.DefaultDelay},
			},
		},
		{"Salakszentmotoros", db.User{UID: to, Name: "Jane Doe", BornOn: &bornOn, BornAt: "Salakszentmotoros"},
			[]api.ResponseWithDelay{
				{Response: api.ResponseMessage{To: to, Text: "Hello Jane Doe from Salakszentmotoros!"}, Delay: config.DefaultDelay},
				{Response: api.ResponseMessage{To: to, Text: "You are still so young!"}, Delay: 0},
			},
		},
	}
}
