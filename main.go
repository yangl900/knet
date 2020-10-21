package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	var listeningAddress = "0.0.0.0:3333"
	var connID = 0

	l, err := net.Listen("tcp", listeningAddress)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	fmt.Println("Listening on " + listeningAddress)
	for {
		conn, err := l.Accept()

		fmt.Printf("Accepted conn %d from %s\n", connID, conn.RemoteAddr())
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, connID)
		connID++
	}
}

func handleConnection(conn net.Conn, id int) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	// Close the connection when you're done with it.
	defer conn.Close()
	count := 0
	for {
		time.Sleep(time.Second * 1)

		count++
		_, err := conn.Write([]byte(fmt.Sprintf("[%d] message %d\n", id, count)))
		if err != nil {
			fmt.Printf("[%d] Connection closed: %v\n", id, err)
			break
		}
	}
}
