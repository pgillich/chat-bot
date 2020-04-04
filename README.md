# Chat bot

## Introduction

It's a simple chat bot server, without GUI.

## Implementation

It's a simple solution, plans are marked as `TODO` in the source code.

### Framework

Because of simplicity, a HTML framework (Gin, Chi, etc.) was not selected and used.

### Error messages

Errors are checked, but error message sendings were implemented only a few times. If compliance to RFC7807 is important, my error formatter library can be used: <https://github.com/pgillich/errfmt>

### Runtime parameters

It's handled by Cobra+viper pair in `cmd` directory. CLI parameter and environment variable handling were tested, config file handling not.

### Microservice components

In a bigger system, the publisher and subscriber of a message queue (here: Redis) are running in separated components (in different containers), but here they are running as separated services. So, only one binary is compiled and a CLI parameter decides, which service will be run. For simplicity, only one worker instance is running. To run more workers (in more goroutines), the `RedisSubscriber` struct must be refactored.

### Redis

The current implementation uses Redis without authentication. In production, authentication is needed, for example username/password.

Starting:

```sh
sudo systemctl start redis-server.service
```

### JWT

The keys must be generated, for example:

```sh
mkdir -p rsa
cd rsa
ssh-keygen -f chat_rsa
openssl rsa -in chat_rsa -pubout > chat_rsa.pem
cd ..
```
The automatic test uses the private key, too.

The authentication is a simple RSA public key authentication (`chat_rsa.pub`), without username/password.

### Database

Postgres was selected. Post-install steps: 

```sh
sudo -i -u postgres
createuser -P chat_bot # password: bot_chat
createdb -O chat_bot chat_bot
```

It's not secure, because the user cred is stored in Git repo. A secure solution can be giving the user cred by K8s Secret or to have more secure system, Vault or <https://cert-manager.io/docs/installation/kubernetes/> can be used.

### Global variables.

It looks easy, but complex from automatic testing point of view. I tried to avoid it as much as possible.

### Faking

In order to run function tests by `go test` (not only unit tests), network dependencies (Redis, Postgres) must be faked. In Go, faking is more powerful than mocking.

Faking implementation (with the interface) is expensive, but the investment will worth it for long term: it will be possible to write function test quickly.

### Business logic

The current implementation is very simple, in some cases it uses regex. The next step can be tokenizing for making more clever logic.

## Testing

### Test application

Unfortunately, it was not compiled static, so it's not easy to run it in CLI. Example for running it continuously:

```sh
while true; do docker run --rm -it -p "8089:8089"  -v "$PWD/rsa/chat_rsa.pub:/app/chat_rsa.pub" registry.gitlab.com/hellowearemito/go-tester:latest -endpoint "http://172.17.0.1:8088/chat"; done
```

### Running tests

```sh
go vet
golangci-lint run

go test -v -race ./pkg/...
```

The private key, described above, is expected in `rsa`  directory (same to `--rsa-key ../../rsa/chat_rsa` parameter)

## Usage

### Running parameters

It's possible to set by CLI options or environment variables. See the help:

```text
$ ./chat-bot frontend --help
Start chat bot frontend service.

Usage:
  chat-bot frontend [flags]

Flags:
  -h, --help                  help for frontend
      --service-path string   SERVICE_PATH, path to chat bot service (default "/chat")

Global Flags:
      --listen string          LISTEN, host:port listening on (default ":8088")
      --log-level string       LOG_LEVEL, log level (default "DEBUG")
      --redis-channel string   REDIS_CHANNEL, Redis channel name for sending message to worker (default "requests")
      --redis-host string      REDIS_HOST, URL to Redis server (default ":6379")
      --redis-key string       REDIS_KEY, Redis queue key (default "online.chat-bot")
      --redis-user string      REDIS_USER, Redis user name (default "chat-bot")
```

```text
$ ./chat-bot engine --help
Start chat bot engine service.

Usage:
  chat-bot engine [flags]

Flags:
      --client-endpoint string   CLIENT_ENDPOINT, client endpoint (default "http://localhost:8089/")
      --db-host string           DB_HOST, DB host (default "localhost")
      --db-name string           DB_NAME, DB name (default "chat_bot")
      --db-password string       DB_PASSWORD, DB password (default "bot_chat")
      --db-user string           DB_USER, DB user (default "chat_bot")
  -h, --help                     help for engine
      --rsa-key string           RSA_KEY, RSA key for JWT (default "rsa/chat_rsa")

Global Flags:
      --listen string          LISTEN, host:port listening on (default ":8088")
      --log-level string       LOG_LEVEL, log level (default "DEBUG")
      --redis-channel string   REDIS_CHANNEL, Redis channel name for sending message to worker (default "requests")
      --redis-host string      REDIS_HOST, URL to Redis server (default ":6379")
      --redis-key string       REDIS_KEY, Redis queue key (default "online.chat-bot")
      --redis-user string      REDIS_USER, Redis user name (default "chat-bot")
```

## Running

Default parameters are tested on Ubuntu 16.04. Redis and Postgres are running as daemon on the host machine, with default configuration.

Compilation and running:

```sh
go build

./chat-bot service
# In another shell
./chat-bot engine --listen ':8087'
```

### Docker compose

It can be use to run more components on the developer machine (including Redis and Postgres, too).

Image building:

```sh
docker build -t pgillich/chat-bot .
```

Start:

```sh
mkdir -p tmp/postgres
docker-compose up
```

Connecting to Postgres server:

```sh
psql -h localhost -p 14320 -U chat_bot -d chat_bot
```

Connecting to Redis:

```sh
redis-cli -p 16379 monitor
```

### Kubernetes

It's ideal for prod and CI environment. Not implemented.

### CI

It's possible to do more activities, for example:

* build
* static checks
* Docker image build, push
* integration tests
* deployment
* release
