package p0

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

type multiEchoServer struct {
	ln               net.Listener
	clients          map[string]*multiEchoClient
	addClientChan    chan *multiEchoClient
	removeClientChan chan *multiEchoClient
	readChan         chan []byte
	closeChan        chan bool
	countRequestChan chan chan int
}

type multiEchoClient struct {
	conn      net.Conn
	writeChan chan []byte
	closeChan chan bool
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	mes := new(multiEchoServer)
	mes.clients = make(map[string]*multiEchoClient)
	mes.addClientChan = make(chan *multiEchoClient)
	mes.removeClientChan = make(chan *multiEchoClient)
	mes.readChan = make(chan []byte)
	mes.closeChan = make(chan bool)
	mes.countRequestChan = make(chan chan int)
	return mes
}

func (mes *multiEchoServer) Start(port int) error {
	fmt.Println("Start listening.")
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	mes.ln = ln
	go mes.serverRoutine()
	return err
}

func (mes *multiEchoServer) Close() {
	mes.closeChan <- true
	mes.ln.Close()
}

func (mes *multiEchoServer) Count() int {
	countChan := make(chan int)
	mes.countRequestChan <- countChan
	return <-countChan
}

func (mes *multiEchoServer) serverRoutine() {
	go mes.handleAccept(mes.ln)

	for {
		select {
		case line := <-mes.readChan:
			// 给所有的客户端发信息
			for _, client := range mes.clients {
				if len(client.writeChan) < 100 {
					client.writeChan <- line
				}
			}
		case client := <-mes.addClientChan:
			mes.clients[client.conn.RemoteAddr().String()] = client
			go client.clientRoutine(mes)
			go mes.handleRead(client)
		case client := <-mes.removeClientChan:
			client.closeChan <- true
			delete(mes.clients, client.conn.RemoteAddr().String())
		case countChan := <-mes.countRequestChan:
			countChan <- len(mes.clients)
		case <-mes.closeChan:
			for _, client := range mes.clients {
				client.closeChan <- true
				delete(mes.clients, client.conn.RemoteAddr().String())
			}
			return
		}
	}
}

func (mes *multiEchoServer) handleAccept(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		client := new(multiEchoClient)
		client.conn = conn
		client.writeChan = make(chan []byte, 100)
		client.closeChan = make(chan bool)
		mes.addClientChan <- client
	}
}

func (mes *multiEchoServer) handleRead(client *multiEchoClient) {
	reader := bufio.NewReader(client.conn)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			mes.removeClientChan <- client
			return
		}

		mes.readChan <- line
	}
}

func (client *multiEchoClient) clientRoutine(mes *multiEchoServer) {
	for {
		select {
		case line := <-client.writeChan:
			_, err := client.conn.Write(line)
			if err != nil {
				mes.removeClientChan <- client
				return
			}
		case <-client.closeChan:
			client.conn.Close()
			return
		}
	}
}
