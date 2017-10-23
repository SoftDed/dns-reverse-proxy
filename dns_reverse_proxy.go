package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v2"
)

var (
	configFile = flag.String("config", "/usr/local/etc/dns-proxy.yaml", "The path to the file containing the list of routes and allow transfers")
	bind       = flag.String("bind", "127.0.0.1:53", "IP:PORT bind proxy")
	config     *Config
	mutex      = &sync.RWMutex{}
)

type Config struct {
	Routes        map[string]string
	Transfers     map[string][]string
	DefaultServer string
}

func getConf() *Config {
	mutex.Lock()
	defer mutex.Unlock()

	var newConfig *Config
	yamlFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &newConfig)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return newConfig
}

func main() {
	flag.Parse()
	config = getConf()

	udpServer := &dns.Server{Addr: *bind, Net: "udp"}
	tcpServer := &dns.Server{Addr: *bind, Net: "tcp"}
	dns.HandleFunc(".", route)
	go func() {
		if err := udpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		if err := tcpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// SIGHUP reload config
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	go func() {
		for {
			<-reloadSignal
			config = getConf()
		}
	}()

	// Wait for SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	udpServer.Shutdown()
	tcpServer.Shutdown()
}

func route(w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) == 0 || !allowed(w, req) {
		dns.HandleFailed(w, req)
		return
	}

	mutex.RLock()
	defer mutex.RUnlock()
	if addr, ok := config.Routes[req.Question[0].Name]; ok {
		proxy(addr, w, req)
		return
	}
	proxy(config.DefaultServer, w, req)
}

func isTransfer(req *dns.Msg) bool {
	for _, q := range req.Question {
		switch q.Qtype {
		case dns.TypeIXFR, dns.TypeAXFR:
			return true
		}
	}
	return false
}

func allowed(w dns.ResponseWriter, req *dns.Msg) bool {
	if !isTransfer(req) {
		return true
	}

	mutex.RLock()
	defer mutex.RUnlock()
	remote, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	for _, ip := range config.Transfers[req.Question[0].Name] {
		if ip == remote {
			return true
		}
	}
	return false
}

func proxy(addr string, w dns.ResponseWriter, req *dns.Msg) {
	transport := "udp"
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		transport = "tcp"
	}
	if isTransfer(req) {
		if transport != "tcp" {
			dns.HandleFailed(w, req)
			return
		}
		t := new(dns.Transfer)
		c, err := t.In(req, addr)
		if err != nil {
			dns.HandleFailed(w, req)
			return
		}
		if err = t.Out(w, req, c); err != nil {
			dns.HandleFailed(w, req)
			return
		}
		return
	}
	c := &dns.Client{Net: transport}
	resp, _, err := c.Exchange(req, addr)
	if err != nil {
		dns.HandleFailed(w, req)
		return
	}
	w.WriteMsg(resp)
}
