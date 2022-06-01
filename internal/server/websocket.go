package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mattn/go-sqlite3"
	"github.com/nilp0inter/komejirushi/internal/server/commands"
	"github.com/nilp0inter/komejirushi/internal/server/config"
	"github.com/nilp0inter/komejirushi/internal/server/search"
)

var addr = "localhost:8080"

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options

func responder(c *websocket.Conn) chan<- commands.SearchResponse {
	out := make(chan commands.SearchResponse)
	go func() {
		defer close(out)
		for msg := range out {
			err := c.WriteJSON(msg)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}()
	return out
}

func receiveCommands(cfg config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		for {
			var msg commands.ServerCommand
			err := c.ReadJSON(&msg)
			if err != nil {
				log.Println("read error:", err)
				break
			}
			if msg.Command == "search" {
				var s commands.SearchPayload
				err = json.Unmarshal(msg.Payload, &s)
				if err != nil {
					log.Println("search error:", err)
					break
				}
				go search.MakeSearch(cfg, s.Term, responder(c))
				// TODO: Start search pipeline
			} else {
				log.Println("invalid command:", msg.Command)
			}
		}
	}
}

func RunServer(docsets ...string) {

	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{
			Extensions: []string{
				"sqlite3_mod_komejirushi",
			},
		})

	c := config.Config{}
	c.Docsets = make(map[string]*sql.DB)

	for _, path := range docsets {
		hld, err := sql.Open("sqlite3_extended", path)
		if err != nil {
			log.Fatal(err)
		}
		defer hld.Close()
		c.Docsets[path] = hld
	}

	http.HandleFunc("/commands", receiveCommands(c))
	log.Fatal(http.ListenAndServe(addr, nil))
}
