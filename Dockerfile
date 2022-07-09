FROM golang:1.18.1-stretch as builder

COPY src /go/src
WORKDIR /go/src

RUN go build .

WORKDIR /go/proxy/cmd/minetest-go
RUN go build -buildvcs=false .

## plugins & libs:
#COPY plugins /go/plugin_src
#COPY libs /go/libs
#COPY plugin_installer.sh /go/plugin_installer.sh
#RUN sh /go/plugin_installer.sh

# for now on
COPY ./plugins /go/srv/plugins

COPY ./config.json /go/proxy/cmd/minetest-go

#EXPOSE 30000/udp listens on port 30000 but not exposed

CMD [ "/go/proxy/cmd/minetest-go/minetest-go"]
