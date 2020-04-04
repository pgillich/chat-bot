// Package engine implements the business logic
package engine

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/garyburd/redigo/redis"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pgillich/chat-bot/api"
	"github.com/pgillich/chat-bot/config"
	"github.com/pgillich/chat-bot/internal/db"
	"github.com/pgillich/chat-bot/internal/logger"
	"github.com/pgillich/chat-bot/internal/queue"
)

// SimpleClaim is a simple JWT claim
type SimpleClaim struct {
	*jwt.StandardClaims
	TokenType string
}

// App is the service, called by automatic test, too
func App(idleConnsClosed chan struct{},
	subscriber queue.RedisSubscriber, dbHandler db.DbHandler,
	httpClient *http.Client, rsaKeyPath string, clientEndpoint string,
	logLevel string,
) *http.ServeMux {
	logger.Init(logLevel)

	if err := subscriber.Connect(); err != nil {
		logger.Get().Panic("cannot connect to Redis", err)
	}

	if err := dbHandler.Connect(); err != nil {
		logger.Get().Panic("cannot connect to DB", err)
	}

	go Worker(idleConnsClosed, subscriber, dbHandler, httpClient, rsaKeyPath, clientEndpoint)

	serverMux := http.NewServeMux()

	serverMux.Handle("/metrics", promhttp.Handler())

	return serverMux
}

// Worker is the main func of the engine
func Worker(idleConnsClosed chan struct{}, subscriber queue.RedisSubscriber,
	dbHandler db.DbHandler, httpClient *http.Client,
	rsaKeyPath string, clientEndpoint string,
) {
	defer subscriber.Close()

	signBytes, err := ioutil.ReadFile(rsaKeyPath) // nolint:gosec
	if err != nil {
		logger.Get().Panic("cannot open private key, ", err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		logger.Get().Panic("cannot parse private key, ", err)
	}

	token := makeAutorefreshToken(signKey)

	for {
		select {
		case <-idleConnsClosed:
			return
		default:
			switch msg := subscriber.Receive().(type) {
			case redis.Message:
				sendResponses(httpClient, clientEndpoint, token.GetToken(), makeResponses(dbHandler, msg))
			case redis.Subscription:
				// We don't need to listen to subscription messages,
			case error:
				if strings.Contains(msg.Error(), "use of closed network connection") {
					logger.Get().Info("Closing connection, no more messages")
					return
				}

				logger.Get().Warningf("cannot receive from request channel, %s", msg)
			}
		}
	}
}

// AutorefreshToken provides token, refreshed regularly
type AutorefreshToken struct {
	signKey *rsa.PrivateKey
	token   string

	mx *sync.Mutex
}

func makeAutorefreshToken(signKey *rsa.PrivateKey) *AutorefreshToken {
	token, err := createToken(signKey)
	if err != nil {
		logger.Get().Panic("cannot make token, ", err)
	}

	autoToken := &AutorefreshToken{
		signKey: signKey,
		token:   token,
		mx:      &sync.Mutex{},
	}

	logger.Get().Debugf("New JWT token: %s", token)

	go func() {
		ticker := time.Tick(config.TokenRefreshDuration) // nolint:staticcheck
		for range ticker {
			autoToken.mx.Lock()

			var err error
			autoToken.token, err = createToken(signKey)
			if err != nil {
				logger.Get().Panic("cannot make token, ", err)
			}
			logger.Get().Debugf("New JWT token: %s", token)

			autoToken.mx.Unlock()
		}
	}()

	return autoToken
}

// GetToken returns a JWT token
// token is refreshed regularly
func (autoToken *AutorefreshToken) GetToken() string {
	autoToken.mx.Lock()
	defer autoToken.mx.Unlock()

	return autoToken.token
}

func createToken(signKey *rsa.PrivateKey) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Claims = &SimpleClaim{
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(config.TokenExpirationDuration).Unix(),
			Subject:   "chat-bot",
		},
		"level1",
	}

	return t.SignedString(signKey)
}

func sendResponses(httpClient *http.Client, clientEndpoint string, token string,
	messages []api.ResponseWithDelay,
) {
	go func() {
		for _, messageWithDelay := range messages {
			time.Sleep(messageWithDelay.Delay)
			logger.Get().Infof("SEND %s", messageWithDelay)

			reqBody, _ := json.Marshal(messageWithDelay.Response) // nolint:errcheck

			req, err := http.NewRequest("POST", clientEndpoint, bytes.NewReader(reqBody))
			if err != nil {
				logger.Get().Warning("cannot create POST to client", err)
				return
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

			resp, err := httpClient.Do(req)
			if err != nil {
				logger.Get().Warning("cannot send POST to client", err)
				return
			}
			if resp.Body != nil {
				defer resp.Body.Close() // nolint:errcheck
			}
		}
	}()
}

func makeResponses(dbHandler db.DbHandler, request redis.Message) []api.ResponseWithDelay {
	var requestMessage api.RequestMessage
	if err := json.Unmarshal(request.Data, &requestMessage); err != nil {
		return []api.ResponseWithDelay{
			newResponseWithDelay(requestMessage.From, "invalid request format", config.DefaultDelay),
		}
	}

	logger.Get().Info("RECEIVED", requestMessage)

	if len(requestMessage.From) == 0 {
		return []api.ResponseWithDelay{
			newResponseWithDelay(requestMessage.From, "empty user name", config.DefaultDelay),
		}
	}

	id := requestMessage.From
	if len(id) == 0 {
		return []api.ResponseWithDelay{
			newResponseWithDelay(requestMessage.From, "missing user ID", config.DefaultDelay),
		}
	}

	user, err := dbHandler.GetOrCreateUser(id)
	if err != nil {
		return []api.ResponseWithDelay{
			newResponseWithDelay(requestMessage.From, "cannot create user, "+err.Error(), config.DefaultDelay),
		}
	}

	var responses []api.ResponseWithDelay

	requestText := strings.TrimSpace(requestMessage.Text)
	if requestText == "" {
		return []api.ResponseWithDelay{
			newResponseWithDelay(requestMessage.From, "Well...", config.DefaultDelay),
		}
	}

	user, responses = makeStatefulResponses(user, requestText)

	if err := dbHandler.Update(user); err != nil {
		return []api.ResponseWithDelay{
			newResponseWithDelay(requestMessage.From, "cannot update user, "+err.Error(), config.DefaultDelay),
		}
	}

	return responses
}

func newResponseWithDelay(to string, text string, delay time.Duration) api.ResponseWithDelay {
	return api.ResponseWithDelay{
		Response: api.ResponseMessage{
			To:   to,
			Text: text,
		},
		Delay: delay,
	}
}

var (
	reFirstHi    = regexp.MustCompile(`(?i)^Hi[.]*$`)                            // nolint:gochecknoglobals
	reFirstHello = regexp.MustCompile(`(?i)^Hello[.]*$`)                         // nolint:gochecknoglobals
	reMyNameIs   = regexp.MustCompile(`(?i)^My name is[.\s]*(.*[^.^\s])[.\s]*$`) // nolint:gochecknoglobals
	reLocation   = regexp.MustCompile(`(?i)(?i)^(.*[^.^\s])[.\s]*$`)             // nolint:gochecknoglobals
)

func extractName(text string) string {
	parts := reMyNameIs.FindStringSubmatch(text)
	if len(parts) == 2 {
		return parts[1]
	}

	return text
}

func extractLocation(text string) string {
	parts := reLocation.FindStringSubmatch(text)
	if len(parts) == 2 {
		return parts[1]
	}

	return text
}

func longDelay() time.Duration {
	return time.Millisecond *
		time.Duration(config.LongDelayMinMillis+rand.Int31n(config.LongDelayMaxMillis-config.LongDelayMinMillis))
}

func makeStatefulResponses(user db.User, text string) (db.User, []api.ResponseWithDelay) { // nolint:gocritic
	var responses []api.ResponseWithDelay

	to := user.UID
	loggerUser := logger.Get().WithField("USER", user.UID)

	if reFirstHi.MatchString(text) || reFirstHello.MatchString(text) { // First question
		user = db.User{Model: user.Model, UID: user.UID}
		responses = []api.ResponseWithDelay{
			newResponseWithDelay(to, "Hi", config.DefaultDelay),
			newResponseWithDelay(to, "What's your name?", config.DefaultDelay),
		}
	} else if len(user.Name) == 0 { // User name expected
		user.Name = extractName(text)
		loggerUser.Info("NAME", user.Name)

		responses = []api.ResponseWithDelay{
			newResponseWithDelay(to, "When were you born?", config.DefaultDelay),
		}
	} else if user.BornOn == nil || user.BornOn.IsZero() { // Born date expected
		date, err := time.Parse("2006.01.02.", text)
		if err != nil {
			responses = []api.ResponseWithDelay{
				newResponseWithDelay(to, "Please specify your born date in YYYY.mmm.dd. format: "+text, config.DefaultDelay),
			}
		} else {
			user.BornOn = &date
			loggerUser.Info("BORN_ON", user.BornOn)

			responses = []api.ResponseWithDelay{
				newResponseWithDelay(to, "Where were you born?", config.DefaultDelay),
			}
		}
	} else if len(user.BornAt) == 0 { // Born location expected
		user.BornAt = extractLocation(text)
		loggerUser.Info("BORN_ON", user.BornAt)

		responses = []api.ResponseWithDelay{
			newResponseWithDelay(to, fmt.Sprintf("Hello %s from %s!", user.Name, user.BornAt), config.DefaultDelay),
		}

		age := int(time.Since(*user.BornOn).Seconds() / 31556952)
		if age < 30 {
			responses = append(responses, newResponseWithDelay(to, "You are still so young!", longDelay()))
		} else /* if user.Name == "John Doe" */ {
			responses = append(responses, newResponseWithDelay(to,
				fmt.Sprintf("You are %d years old. Hey, that's still younger than the Millenium Falcon!", // nolint:misspell
					age), longDelay()))
		}
	} else {
		responses = []api.ResponseWithDelay{}

		loggerUser.Info("NO RULE")
	}

	return user, responses
}
