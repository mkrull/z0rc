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
	"time"

	"github.com/mkrull/z0rc/registry"

	"goji.io"
	"goji.io/pat"
)

var hostname = flag.String("hostname", "localhost", "hostname to be used to register the node")
var port = flag.Int("port", 8001, "port to be used to run the node's interface")
var discoveryURL = flag.String("discover", "http://localhost:8000/discover", "discovery url to register node")
var discoveryToken = flag.String("token", "", "token of the cluster to join")

type node struct {
	sync.Mutex
	discoveryBaseURL string
	discoveryToken   string
	cluster          *registry.Register
}

func (n *node) heartbeat() {
	go func() {
		for {
			time.Sleep(2 * time.Second)
			for _, c := range n.cluster.Nodes {
				if c.FQDN == *hostname && c.Port == *port {
					continue
				}

				st := struct {
					FQDN string
					Port int
				}{
					FQDN: *hostname,
					Port: *port,
				}
				info, err := json.Marshal(st)

				if err != nil {
					log.Println(err)
					os.Exit(1)
				}

				buff := bytes.NewBuffer(info)
				log.Println("payload send", buff.String())
				resp, err := http.Post("http://"+c.FQDN+":"+strconv.Itoa(c.Port)+"/heartbeat", "application/json", buff)
				if err != nil {
					log.Println("Error:", err)
					c.Dead = true
					continue
				}
				c.Dead = false
				payload, err := ioutil.ReadAll(resp.Body)
				log.Println("payload receive", string(payload))
			}
		}
	}()
}

func (n *node) newToken() {
	resp, err := http.Get(n.discoveryBaseURL + "/new")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Token: ", string(id))

	n.discoveryToken = string(id)
}

func (n *node) register() {
	if n.discoveryToken == "" {
		n.newToken()
	}

	nodeInfo := registry.NodeInfo{
		FQDN: *hostname,
		Port: *port,
	}

	nodeBytes, err := json.Marshal(nodeInfo)

	buff := bytes.NewBufferString(string(nodeBytes))

	log.Println(n.discoveryBaseURL + "/" + n.discoveryToken + "/" + "register")

	resp, err := http.Post(n.discoveryBaseURL+"/"+n.discoveryToken+"/"+"register", "application/json", buff)
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

	register, err := registry.RegisterFromBytes(res)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	n.cluster = register
}

var nodeInfo *node

func init() {
	flag.Parse()
	nodeInfo = &node{
		discoveryBaseURL: *discoveryURL,
		discoveryToken:   *discoveryToken,
	}

	nodeInfo.register()
	nodeInfo.heartbeat()
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error", err)
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println("Heartbeat:", string(payload))

	st := struct {
		FQDN string
		Port int
	}{
		FQDN: *hostname,
		Port: *port,
	}
	info, err := json.Marshal(st)

	fmt.Fprint(w, string(info))
}

func replicationHandler(w http.ResponseWriter, r *http.Request) {
	u := pat.Param(r, "uuid")
	fmt.Fprintf(w, "replication for cluster %s", u)
}

func main() {
	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/replicate/:uuid"), replicationHandler)
	mux.HandleFunc(pat.Post("/heartbeat"), heartbeatHandler)

	log.Println(*hostname + ":" + strconv.Itoa(*port))
	http.ListenAndServe(*hostname+":"+strconv.Itoa(*port), mux)
}
