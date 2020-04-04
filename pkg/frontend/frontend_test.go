package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pgillich/chat-bot/api"
	"github.com/pgillich/chat-bot/config"
	"github.com/pgillich/chat-bot/internal/db"
	"github.com/pgillich/chat-bot/internal/queue"
	"github.com/pgillich/chat-bot/internal/test"
	"github.com/pgillich/chat-bot/pkg/engine"
)

func buildServerFrontend(idleConnsClosed chan struct{},
	publisher queue.RedisPublisher,
) *httptest.Server {
	return httptest.NewServer(App(idleConnsClosed,
		publisher,
		config.DefaultChatPath,
		test.GetLogLevel()))
}

func buildServerEngine(idleConnsClosed chan struct{},
	subscriber queue.RedisSubscriber, dbHandler db.DbHandler,
	httpClient *http.Client,
) *httptest.Server {
	return httptest.NewServer(engine.App(idleConnsClosed,
		subscriber, dbHandler, httpClient,
		"../../"+config.DefaultRsaKey, config.DefaultClientEndpoint,
		test.GetLogLevel()))
}

func post(testServer *httptest.Server, postBody string) (*http.Response, error) {
	rootURL, _ := url.Parse(testServer.URL) // nolint:errcheck

	testClient := testServer.Client()

	return testClient.Post("http://"+rootURL.Host+config.DefaultChatPath, "application/json",
		bytes.NewBufferString(postBody))
}

type FakeTransport struct {
	reqBodies []string
	Header    http.Header

	mx sync.Mutex
}

func (fakeTansport *FakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fakeTansport.mx.Lock()
	defer fakeTansport.mx.Unlock()

	reqBody, _ := ioutil.ReadAll(req.Body) // nolint:errcheck

	req.Body.Close() // nolint:errcheck,gosec

	fakeTansport.reqBodies = append(fakeTansport.reqBodies, string(reqBody))

	response := &http.Response{
		Header:        fakeTansport.Header,
		Body:          ioutil.NopCloser(bytes.NewBufferString("")),
		StatusCode:    200,
		Status:        http.StatusText(200),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: 0,
		Request:       req,
	}

	return response, nil
}

// GetReqBodies returns collected request bodies
func (fakeTansport *FakeTransport) GetReqBodies() []string {
	fakeTansport.mx.Lock()
	defer fakeTansport.mx.Unlock()

	return fakeTansport.reqBodies
}

// ResetBodies clears collected request bodies
func (fakeTansport *FakeTransport) ResetReqBodies() {
	fakeTansport.mx.Lock()
	defer fakeTansport.mx.Unlock()

	fakeTansport.reqBodies = []string{}
}

func testE2E(t *testing.T, uid string, messagePairs []test.MessagePair) {
	fakeRedis := &queue.FakeRedis{}
	defer fakeRedis.Close()

	publisher := fakeRedis
	subscriber := fakeRedis

	dbHandler := &db.FakeDbHandler{}
	defer dbHandler.Close()

	fakeTransport := FakeTransport{}
	httpClient := &http.Client{
		Transport: &fakeTransport,
	}

	idleConnsClosed := make(chan struct{}, 1)
	defer close(idleConnsClosed)

	frontendServer := buildServerFrontend(idleConnsClosed, publisher)
	defer frontendServer.Close()

	engineServer := buildServerEngine(idleConnsClosed, subscriber, dbHandler, httpClient)
	defer engineServer.Close()

	for m, messagePair := range messagePairs { // nolint:gocritic
		fakeTransport.ResetReqBodies()

		requestMessage := api.RequestMessage{From: uid, Text: messagePair.In}
		requestBody, _ := json.Marshal(requestMessage) // nolint:errcheck

		resp, err := post(frontendServer, string(requestBody))
		if !assert.NoError(t, err, "POST") {
			return
		}

		resp.Body.Close() // nolint:errcheck,gosec

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Status")

		time.Sleep(config.DefaultDelay)

		for _, expectedResponse := range messagePair.Responses {
			if expectedResponse.Delay == 0 {
				time.Sleep(time.Millisecond * config.LongDelayMaxMillis)
			} else {
				time.Sleep(expectedResponse.Delay)
			}
		}

		reqBodies := fakeTransport.GetReqBodies()
		if len(messagePair.Responses) != len(reqBodies) {
			assert.Equal(t, len(messagePair.Responses), len(reqBodies), fmt.Sprintf("Responses #%d", m))
		}

		for r, expectedResponse := range messagePair.Responses {
			assert.Equal(t, expectedResponse.Response.To, uid, fmt.Sprintf("To #%d/%d", m, r))

			responseBytes, _ := json.Marshal(api.ResponseMessage{To: uid, Text: expectedResponse.Response.Text}) // nolint:errcheck
			assert.Equal(t, string(responseBytes), reqBodies[r], fmt.Sprintf("Text #%d/%d", m, r))
		}
	}
}

func TestJohnDoeE2E(t *testing.T) {
	uid := "001"
	messagePairs := test.GetMessagePairsJohnDoe(uid)

	testE2E(t, uid, messagePairs)
	testE2E(t, uid, messagePairs)
}

func TestJaneDoeE2E(t *testing.T) {
	uid := "002"
	messagePairs := test.GetMessagePairsJaneDoe(uid)

	testE2E(t, uid, messagePairs)
	testE2E(t, uid, messagePairs)
}
