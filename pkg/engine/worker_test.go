package engine

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pgillich/chat-bot/api"
	"github.com/pgillich/chat-bot/internal/db"
	"github.com/pgillich/chat-bot/internal/logger"
	"github.com/pgillich/chat-bot/internal/test"
)

func TestMain(m *testing.M) {
	logger.Init(test.GetLogLevel())

	exitVal := m.Run()

	os.Exit(exitVal)
}

func testStatefulResponses(t *testing.T, user db.User, messagePairs []test.MessagePair) { // nolint:gocritic
	for m, messagePair := range messagePairs { // nolint:gocritic
		var responses []api.ResponseWithDelay

		user, responses = makeStatefulResponses(user, messagePair.In)

		assert.Equal(t, messagePair.User.UID, user.UID, fmt.Sprintf("UID #%d", m))
		assert.Equal(t, messagePair.User.Name, user.Name, fmt.Sprintf("Name #%d", m))
		assert.Equal(t, messagePair.User.BornAt, user.BornAt, fmt.Sprintf("BornAt #%d", m))
		assert.Equal(t, messagePair.User.BornOn, user.BornOn, fmt.Sprintf("BornOn #%d", m))

		assert.Equal(t, len(messagePair.Responses), len(responses), fmt.Sprintf("Responses #%d", m))

		for r, response := range responses {
			assert.Equal(t, messagePair.Responses[r].Response.To, response.Response.To, fmt.Sprintf("To #%d/%d", m, r))
			assert.Equal(t, messagePair.Responses[r].Response.Text, response.Response.Text, fmt.Sprintf("Text #%d/%d", m, r))

			if messagePair.Responses[r].Delay != 0 {
				assert.Equal(t, messagePair.Responses[r].Delay, response.Delay, fmt.Sprintf("Text #%d/%d", m, r))
			}
		}
	}
}

func TestJohnDoe(t *testing.T) {
	uid := "001"
	user := db.User{UID: uid}

	messagePairs := test.GetMessagePairsJohnDoe(uid)

	testStatefulResponses(t, user, messagePairs)
}

func TestJaneDoe(t *testing.T) {
	uid := "002"
	user := db.User{UID: uid}

	messagePairs := test.GetMessagePairsJaneDoe(uid)

	testStatefulResponses(t, user, messagePairs)
}

func TestNameJohnDoe(t *testing.T) {
	text := "John Doe"
	assert.Equal(t, "John Doe", extractName(text))
}

func TestNameJaneDoe(t *testing.T) {
	text := "My name is Jane Doe."
	assert.Equal(t, "Jane Doe", extractName(text))
}

func TestLocationJohnDoe(t *testing.T) {
	text := "Mucsajröcsöge"
	assert.Equal(t, "Mucsajröcsöge", extractLocation(text))
}

func TestLocationJaneDoe(t *testing.T) {
	text := "Salakszentmotoros."
	assert.Equal(t, "Salakszentmotoros", extractLocation(text))
}
