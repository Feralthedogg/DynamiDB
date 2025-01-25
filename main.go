// main.go

package main

import (
	"log"
	"net"
)

func main() {
	cache := NewLRUCache(1000)
	slabMgr := NewMultiSlabManager()
	ttlMgr := NewTTLManager(cache)

	listener, err := net.Listen("tcp", ":11212")
	if err != nil {
		log.Fatalf("Failed to listen on port 11212: %v\n", err)
	}
	defer listener.Close()

	log.Println("DynamiDB listening on :11212")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go handleConnection(conn, cache, slabMgr, ttlMgr)
	}
}
