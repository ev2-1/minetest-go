package http_media

import (
	"github.com/ev2-1/minetest-go/minetest/log"

	"net"
)

func getOutboundIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		log.Warnf("Couldn't get outbound IP using localhost (127.0.0.1): %s", err)
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
