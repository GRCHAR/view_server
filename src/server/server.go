package server

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type Event struct {
	Kind     uint8 `json:"id"`
	When     time.Time
	Mask     uint16 `json:"mask"`
	Reserved uint16 `json:"reserved"`

	Keycode uint16 `json:"keycode"`
	Rawcode uint16 `json:"rawcode"`
	Keychar rune   `json:"keychar"`

	Button uint16 `json:"button"`
	Clicks uint16 `json:"clicks"`

	X int16 `json:"x"`
	Y int16 `json:"y"`

	Amount    uint16 `json:"amount"`
	Rotation  int32  `json:"rotation"`
	Direction uint8  `json:"direction"`
}

type operateStep struct {
	ClientKey string
	Event     Event
}

type client struct {
	addr      net.UDPAddr
	clientKey string
}

type serverControl struct {
	clientMap   map[string]client
	operateChan chan operateStep
	udpConn     net.UDPConn
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

	//tcpListener, err := net.Listen("tcp", ":8002")
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//go func() {
	//	for {
	//		conn, err := tcpListener.Accept()
	//		if err != nil {
	//			log.Println(err)
	//			continue
	//		}
	//		go control.clientLogin(conn)
	//	}
	//}()

	udpAddr, err := net.ResolveUDPAddr("udp", ":8001")
	if err != nil {
		log.Println(err)
		return
	}

	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(listener.LocalAddr(), "Server is running...")
	go func() {
		for {
			buf := make([]byte, 1024)
			n, addr, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Println(err)
			}
			if n < 200 {
				log.Println("addr:", addr, string(buf[:n]))
				go control.clientLogin(addr, string(buf[:n]))
			}
			go control.sendDataToOperateChannel(buf[:n])

		}
	}()
	go control.sendOperateToClient()
	control.udpConn = *listener
}

func (control *serverControl) sendDataToOperateChannel(data []byte) {
	step := operateStep{
		ClientKey: "",
		Event:     Event{},
	}
	json.Unmarshal(data, &step)
	//log.Println(step)
	control.operateChan <- step
}

func (control *serverControl) clientLogin(addr *net.UDPAddr, key string) {
	//buffer := make([]byte, 1024)
	//for {
	//	n, err := conn.Read(buffer)
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//
	//	if n < 1024 {
	//		break
	//	}
	//}
	newClient := client{
		addr:      *addr,
		clientKey: key,
	}
	log.Println(key, "客户端连接")
	control.clientMap[newClient.clientKey] = newClient
	//controlClient := control.clientMap[newClient.clientKey]
	//go controlClient.readOperateFromClient(control)
}

//func (operateClient *client) readOperateFromClient(control *serverControl) {
//	for {
//		buffer := bytes.Buffer{}
//		for {
//			buf := make([]byte, 256)
//			n, err := operateClient.conn.Read(buf)
//			if err != nil {
//				log.Println(err)
//				return
//			}
//
//			buffer.Write(buf[:n])
//			if n < 256 {
//				break
//			}
//			//log.Println(string(buf))
//		}
//
//		if buffer.Len() == 0 {
//			continue
//		}
//		step := operateStep{}
//		log.Println("操作", buffer.String())
//		err := json.Unmarshal(buffer.Bytes(), &step)
//		if err != nil {
//			log.Println(err)
//		}
//		control.operateChan <- step
//	}
//
//}

func (control *serverControl) sendOperateToClient() {
	for {
		select {
		case step := <-control.operateChan:
			stepByte, err := json.Marshal(step)
			if err != nil {
				log.Println(err)
				continue
			}

			if client, ok := control.clientMap[step.ClientKey]; ok {
				addr := client.addr
				raddr := addr.IP.String() + ":" + strconv.Itoa(addr.Port)
				log.Println("发送地址：", raddr)

				if err != nil {

					log.Println(raddr, err)
					continue
				}
				_, err := control.udpConn.WriteToUDP(stepByte, &client.addr)
				//_, err = conn.Write(stepByte)
				if err != nil {
					log.Println(raddr, err)
					continue
				}
				log.Println("发送：" + string(stepByte))
			}

		}
	}
}
