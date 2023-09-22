package main

import (
	"bufio"
	"strings"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var clients map[string]net.Conn

func main() {
	clients = make(map[string]net.Conn)
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
			fmt.Println("Shutdown initiated")
			close(shutdownCh)

			wg.Wait()
			ln.Close()

			fmt.Println("All connections closed, server shutdown")
			break
		} 
	}
}

func acceptConnection(shutdownCh chan struct{}, wg *sync.WaitGroup, ln *net.TCPListener) {
	defer wg.Done()
	for {
		select {
		case <-shutdownCh:
			fmt.Println("No longer accepting connections")
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

	clients[c.RemoteAddr().String()] = c

	fmt.Println("Connected:", c.RemoteAddr())
	c.Write([]byte("-----------Connected to Server. Type your message:-----------\n"))
	
	for {
		select {
		case <-shutdownCh:
			c.Write([]byte("Server is being shut down....\n"))
			delete(clients, c.RemoteAddr().String())
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
				delete(clients, c.RemoteAddr().String())
				return
			}

			buf_str := strings.Trim(string(buf), "\x00")
			buf_str = strings.TrimSpace(buf_str)
			if buf_str == "/quit" {
				c.Write([]byte("Leaving server....\n"))
				fmt.Println("Disconnected:", c.RemoteAddr())
				delete(clients, c.RemoteAddr().String())
				return
			}

			msg := fmt.Sprintf("[%s]: %s\n", c.RemoteAddr(), buf_str)
			fmt.Print(msg)
			broadcastMessage(c.RemoteAddr().String(), msg)
		}
	}
}

func broadcastMessage(senderAddr string, msg string) {
	for addr, c := range clients {
		if addr == senderAddr {
			continue
		}
		c.Write([]byte(msg))
	}
}