FROM golang:1.8
MAINTAINER yunlzheng

EXPOSE 9174

COPY . /go/src/github.com/yunlzheng/prometheus-pusher

 RUN cd /go/src/github.com/yunlzheng/prometheus-pusher \
  && GOPATH=/go go get\
  && GOPATH=/go go build -o /bin/prometheus_pusher \
  && rm -rf /go/bin /go/pkg /var/cache/apk/*

ENTRYPOINT [ "/bin/prometheus_pusher" ]