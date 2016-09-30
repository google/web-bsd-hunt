FROM golang

RUN mkdir -p /go/bin
COPY bin/server-game  /go/bin

RUN chmod 0755 /go/bin/server-game

COPY bin/server-game-run  /go/bin
RUN chmod 0755 /go/bin/server-game-run

RUN apt-get update && apt-get -y install bsdgames

ENTRYPOINT [ "/go/bin/server-game-run", \
	"--server-host",		"0.0.0.0", \
	"--server-port",		"8080", \
	"--huntd-well-known-host",	"localhost", \
	"--huntd-well-known-port",	"4444", \
	"--rpc-type",			"jsonrpc" \
]

EXPOSE 8080
