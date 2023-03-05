package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

func usage(str string) {
	fmt.Fprint(os.Stderr, str+`Usage: log_tools ACTION file.log [ OPTIONS ]

	ACTION := { filter | help }
	OPTIONS (filter) := { default | info | verbose | warn | error }
`)

}

func main() {
	if len(os.Args) < 2 {
		usage("Please specify a subcommand\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "filter":
		if len(os.Args) < 4 { //prog, cmd, file, filter...
			usage("Filter subcommand requires at least 4 arguments\n")
			os.Exit(1)
		}

		f, err := os.Open(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening logfile: %s!\n", err)
			os.Exit(1)
		}

		defer f.Close()

		// make filter
		filter := make(map[string]struct{})
		for _, v := range os.Args[3:] {
			filter[v] = struct{}{}
		}

		r := csv.NewReader(f)
		w := csv.NewWriter(os.Stdout)

		// scan though all entries
		var rec []string
		for err == nil {
			rec, err = r.Read()
			if err != nil {
				break
			}

			//check if filterd
			if len(rec[0]) < 1 {
				break
			}

			if _, ok := filter[rec[0]]; ok {
				w.Write(rec)
				w.Flush()
			}
		}

		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "Error encounterd: %s\n", err)
		}

	case "help":
		usage("")
		os.Exit(0)

	default:
		usage("")
		os.Exit(1)
	}
}
