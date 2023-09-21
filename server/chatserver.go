package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var clients map[net.Conn]bool

func main() {
	var wg sync.WaitGroup
	shutdownCh := make(chan struct{})
	clients = make(map[net.Conn]bool)
	
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
    if err != nil {
        fmt.Println("Error resolving address:", err)
        os.Exit(1)
    }
    ln, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        fmt.Println("Error listening:", err)
        os.Exit(1)
    }
	defer ln.Close()

	fmt.Println("Chat server is running on localhost:8080")
	wg.Add(1)
	go acceptConnection(shutdownCh, &wg, ln)

	scanner := bufio.NewScanner(os.Stdin)	
	for scanner.Scan() {
		if scanner.Text() == "/quit" {
			close(shutdownCh)

			fmt.Println("Shutting down server....")
			wg.Wait()
			ln.Close()

			fmt.Println("All connections closed, server shutdown")
		}
	}
	fmt.Println("Returning from main")
}


func acceptConnection(shutdownCh chan struct{}, wg *sync.WaitGroup, ln *net.TCPListener) {
	defer wg.Done()
	for {
		select {
		case <-shutdownCh:
			fmt.Println("No more connections")
			return
		default:
			ln.SetDeadline(time.Now().Add(1 * time.Second))
			c, err := ln.Accept()
			if err != nil {
				if nErr, ok := err.(*net.OpError); ok && nErr.Timeout() {
					continue
				}
				fmt.Println(err)
				os.Exit(1)
			}
			wg.Add(1)
			go handleConnection(c, shutdownCh, wg)
		}
	}
}

func handleConnection(c net.Conn, shutdownCh <-chan struct{}, wg *sync.WaitGroup) {
	defer c.Close()
	defer wg.Done()

	fmt.Println("Connected:", c.RemoteAddr())
	clients[c] = true
	
	for {
		select {
		case <-shutdownCh:
			c.Write([]byte("Shutting down server...."))
			delete(clients, c)
			return
		default:
			buf := make([]byte, 1024)
			_, err := c.Read(buf)
			if err != nil {
				delete(clients, c)
				fmt.Println("Disconnected:", c.RemoteAddr())
				return
			}
			c.Write(buf)
		}
	}
}