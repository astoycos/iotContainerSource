FROM golang:1.13

LABEL maintainer="Andrew Stoycos <astoycos@redhat.com>"

WORKDIR /iotContainerSource

COPY go.mod go.sum ./

RUN go mod download 

COPY . . 

RUN GO111MODULE=on go build -o iot ./cmd 

#EXPOSE 5672 5671 

ENTRYPOINT ["./iot"]