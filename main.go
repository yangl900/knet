package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	go startTCPServer("0.0.0.0:8005", 5)
	go startTCPServer("0.0.0.0:8600", 600)
	for {
		time.Sleep(time.Hour * 1)
	}
}

func startTCPServer(addr string, interval int) {
	var connID = 0

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	fmt.Println("Listening on " + addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		fmt.Printf("[%s] Accepted conn %d from %s\n", conn.LocalAddr(), connID, conn.RemoteAddr())

		tcpConn := conn.(*net.TCPConn)
		err = tcpConn.SetKeepAlive(false)
		if err != nil {
			fmt.Println("Failed to set keep alive: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, connID, interval)
		connID++
	}
}

func handleConnection(conn net.Conn, id, interval int) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Println("Error reading:", err.Error())
	}

	conn.Write([]byte(fmt.Sprintf("Hello! Current time is %s. I'm going to send message every %d seconds. No KeepAlive in this TCP connection.\n", time.Now().UTC().Format(time.RFC3339), interval)))

	// Close the connection when you're done with it.
	defer conn.Close()
	count := 0
	for {
		time.Sleep(time.Second * time.Duration(interval))

		count++
		msg := fmt.Sprintf("[%s->%s][%d] message %d timestamp %s\n", conn.LocalAddr(), conn.RemoteAddr(), id, count, time.Now().UTC().Format(time.RFC3339))
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Printf("[%s][%d] Connection closed at count %d: %v\n", conn.LocalAddr(), id, count, err)
			break
		}

		fmt.Printf(msg)
	}
}
