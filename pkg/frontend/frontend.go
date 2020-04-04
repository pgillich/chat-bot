// Package frontend is the frontend
package frontend

import (
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pgillich/chat-bot/internal/logger"
	"github.com/pgillich/chat-bot/internal/queue"
)

// App is the service, called by automatic test, too
func App(idleConnsClosed chan struct{},
	publisher queue.RedisPublisher,
	chatPath string,
	logLevel string,
) *http.ServeMux {
	logger.Init(logLevel)

	if err := publisher.Connect(); err != nil {
		logger.Get().Panic("cannot connect to Redis", err)
	}

	serverMux := http.NewServeMux()

	serverMux.Handle("/metrics", promhttp.Handler())
	serverMux.HandleFunc(chatPath, func(w http.ResponseWriter, r *http.Request) {
		Handler(w, r, publisher)
	})

	return serverMux
}

// Handler is the handler of chat bot service
// nolint:interfacer
func Handler(w http.ResponseWriter, r *http.Request,
	publisher queue.RedisPublisher,
) {
	if r.Body == nil {
		logger.Get().Warning("empty body received")
		// TODO error message to user
		w.WriteHeader(http.StatusNoContent)

		return
	}
	defer r.Body.Close() // nolint:errcheck

	if r.Method != http.MethodPost {
		logger.Get().Warning("not POST method received")
		// TODO error message to user
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Get().Warning("bad body received", err)
		// TODO error message to user
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// nolint:gocritic
	logger.Get().Tracef("REQ: %+v\nHEAD: %+v", r, r.Header)
	logger.Get().Debugf("BODY: %s", string(body))

	if err := publisher.Request(body); err != nil {
		logger.Get().Warning("cannot publish", err)
		// TODO error message to user
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
