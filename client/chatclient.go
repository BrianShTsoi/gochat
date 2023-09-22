// chat_client.go

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	go receiveMessages(conn)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		text := scanner.Text()

		_, err := conn.Write([]byte(text + "\n"))
		if err != nil {
			fmt.Println(err)
			return
		}

		if text == "/quit" {
			return
		}
	}
}

func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Server closed the connection.")
			return
		}
		fmt.Print(message)
	}
}
