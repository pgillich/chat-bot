// Package config contains global configs
package config

import (
	"time"
)

const (
	// OptLogLevel is the log level
	OptLogLevel = "log-level"
	// DefaultLogLevel is the default to OptLogLevel
	DefaultLogLevel = "DEBUG"

	// OptServiceHostPort is the host:port listening on
	OptServiceHostPort = "listen"
	// DefaultServiceHostPort is default value to OptServiceHostPort
	DefaultServiceHostPort = ":8088"

	// OptChatPath is the path to chat bot service
	OptChatPath = "service-path"
	// DefaultChatPath is default value to OptChatPath
	DefaultChatPath = "/chat"

	// OptClientEndpoint is the client endpoint
	OptClientEndpoint = "client-endpoint"
	// DefaultClientEndpoint is default value to OptClientEndpoint
	DefaultClientEndpoint = "http://localhost:8089/"

	// OptRsaKey is RSA key for JWT
	OptRsaKey = "rsa-key"
	// DefaultRsaKey is default value to OptRsaKey
	DefaultRsaKey = "rsa/chat_rsa"

	// OptRedisHost is the URL to Redis server
	OptRedisHost = "redis-host"
	// DefaultRedisHost is default value to OptRedisHost
	DefaultRedisHost = ":6379"

	// OptRedisUser  is the Redis user name
	OptRedisUser = "redis-user"
	// DefaultRedisUser is default value to OptRedisUser
	DefaultRedisUser = "chat-bot"

	// OptRedisKey is the queue key
	OptRedisKey = "redis-key"
	// DefaultRedisKey is default value to OptRedisKey
	DefaultRedisKey = "online." + DefaultRedisUser

	// OptRedisRequestChannel is the channel name for sending message to worker
	OptRedisRequestChannel = "redis-channel"
	// DefaultRedisRequestChannel is default value to OptRedisRequestChannel
	DefaultRedisRequestChannel = "requests"

	// OptDbHost is the DB host
	OptDbHost = "db-host"
	// DefaultDbHost is default value to OptDbHost
	DefaultDbHost = "localhost"

	// OptDbName is the DB name
	OptDbName = "db-name"
	// DefaultDbName is default value to OptDbName
	DefaultDbName = "chat_bot"

	// OptDbUser is the DB user
	OptDbUser = "db-user"
	// DefaultDbUser is default value to OptDbUser
	DefaultDbUser = "chat_bot"

	// OptDbPassword is the DB password
	OptDbPassword = "db-password"
	// DefaultDbPassword is  default value to OptDbPassword
	DefaultDbPassword = "bot_chat"

	// DefaultDelayMillis is longer than 1s
	DefaultDelayMillis = 1100
	// DefaultDelay is longer than 1s
	DefaultDelay = DefaultDelayMillis * time.Millisecond
	// LongDelayMinMillis is 2s
	LongDelayMinMillis = /*DefaultDelayMillis +*/ 2000
	// LongDelayMaxMillis is 5s
	LongDelayMaxMillis = 5000

	// TokenExpirationDuration is the expiration duration of JWT token
	TokenExpirationDuration = time.Hour * 2
	// TokenRefreshDuration is the refresh time of JWT token
	TokenRefreshDuration = time.Hour * 1
)
