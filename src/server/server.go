package server

import (
	"bytes"
	"encoding/json"
	hook "github.com/robotn/gohook"
	"io"
	"log"
	"net"
	"sync"
)

type operateStep struct {
	clientKey string
	event     hook.Event
}

type client struct {
	conn      net.Conn
	clientKey string
}

type serverControl struct {
	clientMap   map[string]client
	operateChan chan operateStep
}

func InitServer() {
	control := serverControl{
		clientMap:   make(map[string]client),
		operateChan: make(chan operateStep, 10000),
	}
	control.ListenServer()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func (control *serverControl) ListenServer() {
	listener, err := net.Listen("tcp", ":8001")
	if err != nil {
		log.Println(err)
	}
	log.Println("Server is running...")
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
			}
			go control.clientLogin(conn)

		}
	}()
	go control.sendOperateToClient()
}

func (control *serverControl) clientLogin(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			return
		}
	}
	newClient := client{
		conn:      conn,
		clientKey: string(buffer),
	}
	log.Println(string(buffer), "客户端连接")
	control.clientMap[newClient.clientKey] = newClient
	controlClient := control.clientMap[newClient.clientKey]
	go controlClient.readOperateFromClient(control)
}

func (operateClient *client) readOperateFromClient(control *serverControl) {
	for {
		buffer := bytes.Buffer{}
		for {
			buf := make([]byte, 1024)
			n, err := operateClient.conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				return
			}
			buffer.Write(buf[:n])
			log.Println(string(buf))
		}

		if buffer.Len() == 0 {
			continue
		}
		step := operateStep{}
		log.Println("操作", buffer.String())
		err := json.Unmarshal(buffer.Bytes(), &step)
		if err != nil {
			log.Println(err)
		}
		control.operateChan <- step
	}

}

func (control *serverControl) sendOperateToClient() {
	for {
		select {
		case step := <-control.operateChan:
			stepByte, err := json.Marshal(step)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(string(stepByte))
			control.clientMap[step.clientKey].conn.Write(stepByte)
		}
	}
}
