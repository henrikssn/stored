package main

import (
	"flag"
	"fmt"
	"github.com/henrikssn/stored/endpoint"
	"github.com/henrikssn/stored/router"
	"github.com/henrikssn/stored/store"
	"log"
	"net"
	"net/rpc"
	"os"
)

var (
	tcpAddr     = flag.String("t", ":8081", "The tcp address to bind to for the internal RPC.")
	httpAddr    = flag.String("h", ":8080", "The http address of which to serve the REST API.")
	showVersion = flag.Bool("v", false, "print stored's version string")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *showVersion {
		fmt.Println("stored-0.0.1")
		return
	}

	rpc.Register(store.New())
	rpc.Register(router.New())
	e := endpoint.New()
	e.RegisterInternalRPC()
	go e.Listen(*httpAddr)

	l, err := net.Listen("tcp", *tcpAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go func() {
		for {
			conn, _ := l.Accept()
			go rpc.ServeConn(conn)
		}
	}()

	quit := make(chan int)
	<-quit // Wait to be told to exit.
}
