package web_manager

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsConn struct {
	*websocket.Conn

	mu sync.Mutex
}

func (conn *wsConn) Send(str string) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	conn.WriteMessage(websocket.TextMessage, []byte(str))
}

var conns = make(map[*wsConn]struct{})
var connsMu sync.RWMutex

func broadcast(str string) {
	connsMu.RLock()
	for conn := range conns {
		conn.Send(str)
	}
	connsMu.RUnlock()
}

func serveStatic(w http.ResponseWriter, file string) {
	if len(file) != 0 && []byte(file)[0] == []byte("/")[0] {
		file = file[1:]
	}

	switch true {
	case strings.HasSuffix(file, "css"):
		w.Header().Add("Content-Type", "text/css")
	case strings.HasSuffix(file, "html"):
		w.Header().Add("Content-Type", "text/html")
	case strings.HasSuffix(file, "js"):
		w.Header().Add("Content-Type", "application/x-javascript")
	}

	data, err := files.ReadFile(file)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("file '%s' not found!", file)))
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

type serve struct{}

func (*serve) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/packet" {
		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}

		c := &wsConn{Conn: conn}

		// add conn
		connsMu.Lock()
		conns[c] = struct{}{}
		connsMu.Unlock()

		go func() {
			for {
				defer connsMu.Unlock()
				defer delete(conns, c)
				defer connsMu.Lock()
				defer conn.Close()
				_, input, err := conn.ReadMessage()
				if err != nil {
					fmt.Println(err)
					break
				}

				cmd, msg := parseCmd(string(input))

				log.Println("command", cmd, msg)

				c.Send("hello 0")

				// send metadata (like packet types etc.)
				c.Send(pkts)
			}
		}()
	} else {
		serveStatic(rw, r.URL.Path)
	}
}

func parseCmd(str string) (cmd, msg string) {
	split := strings.SplitN(str, " ", 2)

	switch len(split) {
	case 1:
		return split[0], ""
	case 2:
		return split[0], split[1]

	default:
		return
	}
}

func init() {
	http.HandleFunc("/packet", func(w http.ResponseWriter, r *http.Request) {

	})

	go http.ListenAndServe(":8012", &serve{})
}
