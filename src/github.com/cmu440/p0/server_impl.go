// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"net"
	"strconv"
)

type multiEchoServer struct {
	// TODO: implement this!
	conns            map[string]net.Conn
	messages         map[string]chan string
	readClientChan   chan string
	addClientChan    chan net.Conn
	removeClientChan chan net.Conn
	countClientChan  chan chan int
	closeChan        chan int
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	conns := make(map[string]net.Conn, 1)
	messages := make(map[string]chan string, 1)
	readClientChan := make(chan string, 1)
	addClientChan := make(chan net.Conn, 1)
	removeClientChan := make(chan net.Conn, 1)
	countClientChan := make(chan chan int, 1)
	closeChan := make(chan int, 1)
	return &multiEchoServer{conns, messages, readClientChan, addClientChan,
		removeClientChan, countClientChan, closeChan}
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	go mes.handleListener(listener)
	return err
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	mes.closeChan <- 1
}

func (mes *multiEchoServer) Count() int {
	// TODO: implement this!
	countChan := make(chan int, 1)
	mes.countClientChan <- countChan
	return <-countChan
}

func (mes *multiEchoServer) handleListener(listener net.Listener) {
	//go mes.handleClientMap(mes.conns, mes.messages, mes.addClientChan, mes.removeClientChan,
	//	mes.readClientChan, mes.countClientChan, mes.closeChan)
	go mes.handleClientMap()
	for {
		conn, err := listener.Accept()
		if err != nil {
			//fmt.Printf("listenner accept error\n")
			break
		} else {
			mes.addClientChan <- conn
			go mes.handleConnRead(conn)
		}
	}
}

func (mes *multiEchoServer) handleConnWrite(conn net.Conn, message chan string) {
	for {
		msg := <-message
		_, err := conn.Write([]byte(msg))
		if err != nil {
			// 感觉有问题，好像不符合要求。要求是如果客户端读的很慢，将buffer写满了，则写
			// 到一个queue里面，queue最多存放100条message，超出则丢弃
			conn.Close()
			mes.removeClientChan <- conn
		}
	}
}

func (mes *multiEchoServer) handleConnRead(conn net.Conn) {
	reader := bufio.NewReader(conn)
	// message包括\n
	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			conn.Close()
			mes.removeClientChan <- conn
			break
		}
		receiveStr := string(message)
		mes.readClientChan <- receiveStr
	}
}

func (mes *multiEchoServer) handleClientMap() {
	for {
		select {
		case conn := <-mes.addClientChan:
			mes.conns[conn.RemoteAddr().String()] = conn
			// 下面这句的意思是生成能容下100条string的channel
			message := make(chan string, 100)
			mes.messages[conn.RemoteAddr().String()] = message
			go mes.handleConnWrite(conn, message)
		case conn := <-mes.removeClientChan:
			key := conn.RemoteAddr().String()
			delete(mes.conns, key)
			delete(mes.messages, key)
		case count := <-mes.countClientChan:
			count <- len(mes.conns)
		case receiveStr := <-mes.readClientChan:
			for _, message := range mes.messages {
				// 不太懂一个channel 的len是什么
				if len(message) < 100 {
					message <- receiveStr
				}
			}
		case value := <-mes.closeChan:
			value = value
			for _, conn := range mes.conns {
				conn.Close()
				mes.removeClientChan <- conn
			}
			break
		}
	}
}

// TODO: add additional methods/functions below!
