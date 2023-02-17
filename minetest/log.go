package minetest

import (
	"log"
	"os"
)

var logFlags = log.Flags() | log.LstdFlags | log.Lmsgprefix | log.Lshortfile

var MapLogger = func() *log.Logger {
	lpkts, ok := GetConfig("log-map", false)

	if (ConfigVerbose() && !(ok && !lpkts)) || lpkts {
		return log.New(log.Writer(), "[map] ", logFlags)
	}

	return log.New(&nilWriter{}, "", 0)
}()

type nilWriter struct{}

func (nw *nilWriter) Write(b []byte) (n int, err error) {
	return len(b), nil
}

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
	log.SetFlags(logFlags)

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
