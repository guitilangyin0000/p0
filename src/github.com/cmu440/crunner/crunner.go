package main

import (
	"fmt"
	"net"
)

const (
	defaultHost = "localhost"
	defaultPort = 9999
)

// To test your server implementation, you might find it helpful to implement a
// simple 'client runner' program. The program could be very simple, as long as
// it is able to connect with and send messages to your server and is able to
// read and print out the server's echoed response to standard output. Whether or
// not you add any code to this file will not affect your grade.
func main() {
	conn, err := net.Dial("tcp", "127.0.0.1"+":9999")
	defer conn.Close()
	if err != nil {
		fmt.Println("dial error")
		return
	}
	message := "hello\n"
	//n, err := io.WriteString(conn, message)
	n, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("client: write: %s", err)
		return
	}
	fmt.Printf("client: wrote %q (%d bytes)", message, n)
	reply := make([]byte, 512)
	n, err = conn.Read(reply)
	fmt.Printf("read %d bytes: %s from server\n", n, reply)
}
