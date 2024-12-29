package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func handleConnection(conn net.Conn, cache *LRUCache, slabMgr *MultiSlabManager, ttlMgr *TTLManager) {
	defer func() {
		log.Printf("Connection closed from %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()

	log.Printf("New connection from %s\n", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from %s: %v\n", conn.RemoteAddr().String(), err)
				fmt.Fprintf(conn, "CLIENT_ERROR %v\r\n", err)
			}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		cmd := strings.ToLower(parts[0])

		
		log.Printf("Received command from %s: %q\n", conn.RemoteAddr().String(), line)

		switch cmd {
		case "set":
			if len(parts) < 4 {
				log.Printf("CLIENT_ERROR: invalid set command (not enough parts) from %s\n", conn.RemoteAddr().String())
				fmt.Fprintf(conn, "CLIENT_ERROR invalid set command\r\n")
				continue
			}
			key := parts[1]
			expireSeconds, err := strconv.Atoi(parts[2])
			if err != nil {
				log.Printf("CLIENT_ERROR: invalid expire time %q from %s\n", parts[2], conn.RemoteAddr().String())
				fmt.Fprintf(conn, "CLIENT_ERROR invalid expire time\r\n")
				continue
			}
			valSize, err := strconv.Atoi(parts[3])
			if err != nil {
				log.Printf("CLIENT_ERROR: invalid value size %q from %s\n", parts[3], conn.RemoteAddr().String())
				fmt.Fprintf(conn, "CLIENT_ERROR invalid value size\r\n")
				continue
			}

			log.Printf("SET command: key=%s, expireSeconds=%d, valSize=%d (from %s)\n",
				key, expireSeconds, valSize, conn.RemoteAddr().String())

			valueBuf := make([]byte, valSize)
			_, err = io.ReadFull(reader, valueBuf)
			if err != nil {
				log.Printf("CLIENT_ERROR: could not read value (size=%d) from %s: %v\n", valSize, conn.RemoteAddr().String(), err)
				fmt.Fprintf(conn, "CLIENT_ERROR could not read value\r\n")
				continue
			}
			reader.ReadString('\n')

			allocatedValue := slabMgr.Allocate(valSize)
			copy(allocatedValue, valueBuf)

			cache.Set(key, allocatedValue)

			if expireSeconds > 0 {
				ttlMgr.SetExpire(key, time.Now().Add(time.Duration(expireSeconds)*time.Second))
			}

			fmt.Fprintf(conn, "STORED\r\n")

		case "get":
			if len(parts) < 2 {
				log.Printf("CLIENT_ERROR: invalid get command from %s\n", conn.RemoteAddr().String())
				fmt.Fprintf(conn, "CLIENT_ERROR invalid get command\r\n")
				continue
			}
			key := parts[1]

			log.Printf("GET command: key=%s (from %s)\n", key, conn.RemoteAddr().String())

			value, found := cache.Get(key)
			if !found {
				fmt.Fprintf(conn, "END\r\n")
				continue
			}

			if ttlMgr.IsExpired(key) {
				log.Printf("Key=%s was expired; removing and returning END (from %s)\n", key, conn.RemoteAddr().String())
				cache.Delete(key)
				fmt.Fprintf(conn, "END\r\n")
				continue
			}

			fmt.Fprintf(conn, "VALUE %s %d\r\n", key, len(value))
			conn.Write(value)
			fmt.Fprintf(conn, "\r\nEND\r\n")

		case "delete":
			if len(parts) < 2 {
				log.Printf("CLIENT_ERROR: invalid delete command from %s\n", conn.RemoteAddr().String())
				fmt.Fprintf(conn, "CLIENT_ERROR invalid delete command\r\n")
				continue
			}
			key := parts[1]
			log.Printf("DELETE command: key=%s (from %s)\n", key, conn.RemoteAddr().String())

			cache.Delete(key)
			ttlMgr.DeleteExpire(key)
			fmt.Fprintf(conn, "DELETED\r\n")

		case "quit":
			log.Printf("QUIT command from %s\n", conn.RemoteAddr().String())
			fmt.Fprintf(conn, "BYE\r\n")
			return

		default:
			log.Printf("ERROR: unknown command %q from %s\n", cmd, conn.RemoteAddr().String())
			fmt.Fprintf(conn, "ERROR\r\n")
		}
	}
}

