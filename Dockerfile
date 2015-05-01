FROM golang:1.4.2

ADD . /go/src/github.com/captncraig/mosaicChallenge

RUN go install github.com/captncraig/mosaicChallenge/web

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/web

EXPOSE 7777