package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/mkrull/z0rc/registry"

	"goji.io"
	"goji.io/pat"
)

var hostname = flag.String("hostname", "localhost", "hostname to be used to register the node")
var port = flag.Int("port", 8001, "port to be used to run the node's interface")
var discoverURL = flag.String("discover", "http://localhost:8000/discover", "discovery url to register node")

type node struct {
	sync.Mutex
	id      string
	cluster registry.Register
}

func (n *node) register() {
	resp, err := http.Get(*discoverURL + "/new")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Id: ", string(id))

	n.id = string(id)

	nodeInfo := registry.NodeInfo{
		FQDN: *hostname,
		Port: *port,
	}

	nodeBytes, err := json.Marshal(nodeInfo)

	buff := bytes.NewBufferString(string(nodeBytes))

	log.Println(*discoverURL + "/" + n.id + "/" + "register")

	resp, err = http.Post(*discoverURL+"/"+n.id+"/"+"register", "application/json", buff)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Res: ", string(res))
}

var nodeInfo *node

func init() {
	flag.Parse()
	nodeInfo = &node{}

	nodeInfo.register()
}

func replicationHandler(w http.ResponseWriter, r *http.Request) {
	u := pat.Param(r, "uuid")
	fmt.Fprintf(w, "replication for cluster %s", u)
}

func main() {
	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/replicate/:uuid"), replicationHandler)

	log.Println(*hostname + ":" + strconv.Itoa(*port))
	http.ListenAndServe(*hostname+":"+strconv.Itoa(*port), mux)
}
