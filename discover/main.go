package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/mkrull/z0rc/registry"
	"github.com/nu7hatch/gouuid"

	"goji.io"
	"goji.io/pat"
)

var storage StorageBackend
var pidFile = flag.String("pidfile", "", "PID file. No pid file created if empty")

func init() {
	storage = NewInMemoryStore()
	flag.Parse()
	writePid()
}

func writePid() {
	if *pidFile != "" {
		pid := os.Getpid()
		err := ioutil.WriteFile(*pidFile, []byte(strconv.Itoa(pid)), 0444)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
}

func newDiscoveryGroup(w http.ResponseWriter, r *http.Request) {
	u, _ := uuid.NewV4()
	o := []byte("{}")

	storage.Put(u.String(), o)

	fmt.Fprint(w, u)
}

func discoveryData(w http.ResponseWriter, r *http.Request) {
	u := pat.Param(r, "uuid")
	o, ok := storage.Get(u)

	if ok {
		fmt.Fprint(w, string(o))
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func registerToDiscoveryGroup(w http.ResponseWriter, r *http.Request) {
	u := pat.Param(r, "uuid")
	o, ok := storage.Get(u)

	if ok {
		register, err := registry.RegisterFromBytes(o)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		node, err := registry.NodeInfoFromBytes(payload)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		register.AddNode(node)

		b, err := register.Bytes()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		storage.Put(u, b)

		fmt.Fprint(w, string(b))
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func dumpStore(w http.ResponseWriter, r *http.Request) {
	o, err := storage.Dump()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprint(w, string(o))
}

func main() {
	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/discover/new"), newDiscoveryGroup)
	mux.HandleFunc(pat.Get("/discover/dump"), dumpStore)
	mux.HandleFunc(pat.Get("/discover/:uuid"), discoveryData)
	mux.HandleFunc(pat.Post("/discover/:uuid/register"), registerToDiscoveryGroup)

	http.ListenAndServe("localhost:8000", mux)
}
