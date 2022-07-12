package main

import (
	"net"
	"log"
)

func GetOutboundIP() string {
    conn, err := net.Dial("udp", "1.1.1.1:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}
