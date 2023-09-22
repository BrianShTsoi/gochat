package main

import (
	"bufio"
	// "io"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	shutdownCh := make(chan struct{})
	
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

	detectShutdown(shutdownCh, &wg, ln)
}

func detectShutdown(shutdownCh chan struct{}, wg *sync.WaitGroup, ln *net.TCPListener) {
	scanner := bufio.NewScanner(os.Stdin)	
	for scanner.Scan() {
		if scanner.Text() == "/quit" {
			close(shutdownCh)

			fmt.Println("Shutting down server....")
			wg.Wait()
			ln.Close()

			fmt.Println("All connections closed, server shutdown")
			break
		} 
	}
}

func acceptConnection(shutdownCh chan struct{}, wg *sync.WaitGroup, ln *net.TCPListener) {
	defer wg.Done()
	defer fmt.Println("wg done on accepct Connection")
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
	c.Write([]byte("Hello client\n"))
	
	for {
		select {
		case <-shutdownCh:
			c.Write([]byte("Shutting down server...."))
			return
		default:
			buf := make([]byte, 1024)
			c.SetReadDeadline(time.Now().Add(1 * time.Second))
			_, err := c.Read(buf)

			if err != nil {
                if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                    continue
                }
				fmt.Println("Disconnected:", c.RemoteAddr())
				return
			}
			c.Write(buf)
		}
	}
}