package main

import (
	"flag"
	"fmt"
	"github.com/henrikssn/stored/overlay"
	"github.com/henrikssn/stored/route"
	"github.com/henrikssn/stored/server"
	"github.com/henrikssn/stored/store"
	"log"
	"net"
	"os"
)

var (
	laddr       = flag.String("l", "127.0.0.1:8046", "The address to bind to.")
	caddr       = flag.String("c", "", "The address of the leader.")
	showVersion = flag.Bool("v", false, "print doozerd's version string")
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

	if *laddr == "" {
		fmt.Fprintln(os.Stderr, "require a listen address")
		flag.Usage()
		os.Exit(1)
	}
	log.Println("Listening on", *laddr)

	tsock, err := net.Listen("tcp", *laddr)
	if err != nil {
		panic(err)
	}
	defer tsock.Close()

	ludp, err := net.ResolveUDPAddr("udp", *laddr)
	if err != nil {
		panic(err)
	}
	log.Println("Listening on", ludp.String())

	usock, err := net.ListenUDP("udp", ludp)
	if err != nil {
		panic(err)
	}
	defer usock.Close()

	cudp, err := net.ResolveUDPAddr("udp", *caddr)
	if err != nil {
		panic(err)
	}

	// Start a router
	r := route.NewRouter(usock).Start()
	store := store.New()

	//Start an overlay
	o := overlay.New(r, store, ludp, cudp)
	if *caddr == "" {
		o.SetCoord()
	}
	o.Start()

	server := server.New(o)

	go server.ListenAndServe(tsock)

	quit := make(chan int)
	<-quit // Wait to be told to exit.
}
