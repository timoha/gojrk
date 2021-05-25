package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/timoha/gojrk/jrk"
)

var device = flag.String("device", "", "filepath to device")
var password = flag.String("password", "", "password to be used as last component for actions webhook")
var addr = flag.String("addr", ":8080", "address to serve from")
var certFile = flag.String("cert", "", "location of cert file")
var keyFile = flag.String("key", "", "location of key file")

//go:embed index.html
var index []byte

type ActionRequest struct {
	Handler Handler
	Session Session
}

type Handler struct {
	Name string
}

type Session struct {
	ID     string
	Params map[string]interface{}
}

type ActionResponse struct {
	Session Session
}

type ResponseJSON struct {
}

func actionsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ar ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(ar)

	target, ok := handlers[ar.Handler.Name]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown handler %q", ar.Handler.Name), http.StatusBadRequest)
		return
	}

	j, err := jrk.New(*device)
	if err != nil {
		http.Error(w, fmt.Sprint("connecting to jrk controller:", err), http.StatusInternalServerError)
		return
	}
	defer j.Close()

	if err := j.SetTarget(target); err != nil {
		http.Error(w, fmt.Sprint("setting new target value", err), http.StatusInternalServerError)
		return
	}

	resp := ActionResponse{
		Session: ar.Session,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprint("marshalling json response", err), http.StatusInternalServerError)
		return
	}
}

var handlers = map[string]int{
	"handler_left":   3000,
	"handler_center": 1700,
	"handler_right":  751,
}

func main() {
	flag.Parse()

	if *password == "" {
		// running on local network
		http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Write(index)
		})
	}
	http.HandleFunc("/actions/"+*password, actionsHandlerFunc)

	var err error
	if *certFile != "" && *keyFile != "" {
		err = http.ListenAndServeTLS(*addr, *certFile, *keyFile, nil)
	} else {
		err = http.ListenAndServe(*addr, nil)
	}
	log.Fatal(err)
}
