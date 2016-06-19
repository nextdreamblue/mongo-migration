FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
ENV GOPATH /app
RUN go get github.com/gizak/termui
RUN go get github.com/urfave/cli
RUN go get gopkg.in/mgo.v2
RUN go build -o mm .
ENTRYPOINT ["./mm"]