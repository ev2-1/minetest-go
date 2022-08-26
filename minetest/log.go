package minetest

import (
	"log"
	"os"
)

var logWriter *LogWriter

type LogWriter struct {
	f *os.File
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	n, err = os.Stderr.Write(p)
	if err != nil {
		return
	}

	return lw.f.Write(p)
}

func initLog() {
	log.SetPrefix("[minetest] ")
	log.SetFlags(log.Flags() | log.Lmsgprefix)

	f, err := os.OpenFile(Path("latest.log"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer f.Close()
		select {}
	}()

	logWriter = &LogWriter{f}
	log.SetOutput(logWriter)
}
