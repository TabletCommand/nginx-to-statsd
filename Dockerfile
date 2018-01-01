FROM golang:1.8

ENV GOPATH /go/src

RUN mkdir -p /go/{bin/src}

ADD ./main.go /go/src/main.go
ADD ./main_test.go /go/src/main_test.go

WORKDIR /go/src

RUN go get github.com/hpcloud/tail gopkg.in/alexcesaro/statsd.v2
RUN go test && go build -o /go/bin/nginx-to-statsd

CMD /go/bin/nginx-to-statsd -host localhost -port 8125 -prefix nginx.access.log -file /var/log/nginx/access.log
