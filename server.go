package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	ics "github.com/arran4/golang-ical"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

// CalendarConfig is structure for one canlendar event
type CalendarConfig struct {
	ID                  string `json:"id"`
	URL                 string `json:"url"`
	FetchInterval       int    `json:"fetchInterval"`
	MaximumEntries      int    `json:"maximumEntries"`
	MaximumNumberOfDays int    `json:"maximumNumberOfDays"`
	BroadcastPastEvents bool   `json:"broadcastPastEvents"`
}

func handleAddCalendar(s socketio.Conn, config CalendarConfig) string {
	go func(s socketio.Conn, config CalendarConfig) {
		fmt.Printf("start fetch calendar: %s\n", config.URL)
	}(s, config)
	return ""
}

func parseICal(r io.Reader) error {
	cal, err := ics.ParseCalendar(r)
	if err != nil {
		return err
	}
	for _, c := range cal.Components {
		fmt.Println(c)
	}
	return nil
}

func handleConnSocket(s socketio.Conn) error {
	s.SetContext("")
	fmt.Printf("connected %s's %s\n", s.ID(), s.Namespace())
	return nil
}

func handleEventSocket(s socketio.Conn, msg map[string]interface{}) string {
	fmt.Printf("event %s's %s: %v\n", s.ID(), s.Namespace(), msg)
	return ""
}

func main() {
	f, err := os.Open("US_Holidays.ics")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := parseICal(f); err != nil {
		log.Fatal(err)
	}
}

func start() {
	server := socketio.NewServer(&engineio.Options{Transports: []transport.Transport{websocket.Default}})

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Printf("connected: %+v\n", s)
		return nil
	})
	for ns, events := range map[string][]string{
		"calendar":           {"ADD_CALENDAR"},
		"newsfeed":           {"ADD_FEED"},
		"updatenotification": {"CONFIG", "MODULES"},
	} {
		server.OnConnect("/"+ns, handleConnSocket)
		for _, e := range events {
			fmt.Printf("register event callback: %s-%s\n", ns, e)
			if e == "ADD_CALENDAR" {
				server.OnEvent("/"+ns, e, handleAddCalendar)
			} else {
				server.OnEvent("/"+ns, e, handleEventSocket)
			}
		}
	}

	go server.Serve()

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)
	serveMux.Handle("/socket.io/socket.io.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename := filepath.Join("/Users/qiwa/workspace/opensource/MagicMirror", r.URL.Path)
		log.Println(filename)
		w.Header().Add("Cache-Control", "no-store")
		http.ServeFile(w, r, filename)
	}))
	serveMux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename := filepath.Join("/Users/qiwa/workspace/opensource/MagicMirror", r.URL.Path)
		log.Println(filename)
		w.Header().Add("Cache-Control", "no-store")
		w.Header().Add("Access-Control-Allow-Origin", "http://www.calendarlabs.com")
		http.ServeFile(w, r, filename)
	}))

	log.Println("Starting server...")
	log.Panic(http.ListenAndServe(":8080", serveMux))
}
