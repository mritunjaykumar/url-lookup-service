FROM golang

ENV AD /go/src/alertDispatcher
ENV GOBIN /go/bin

ADD . $AD

WORKDIR $AD

RUN go install ./runner.go

ENTRYPOINT $GOBIN/runner
EXPOSE 8082
