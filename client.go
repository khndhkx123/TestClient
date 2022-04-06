package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	ip          = flag.String("ip", "127.0.0.1", "server IP")
	connections = flag.Int("conn", 1, "number of tcp connections")
)

type C_JoinChannel struct {
	ChannelId int32  `json:"channelname"`
	Playerid  string `json:"playerid"`
}

type C_ChatMsg struct {
	Playerid string `json:"playerid"`
	Msg      string `json:"msg"`
}

func main() {
	flag.Parse()

	addr := *ip + ":3333"
	log.Printf("Connect to %s", addr)

	var conns []net.Conn
	for i := 0; i < *connections; i++ {
		c, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			fmt.Println("failed to connect", i, err)
			i--
			continue
		}
		conns = append(conns, c)
		time.Sleep(time.Millisecond)
	}

	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()

	log.Printf("Connection %d DONE !", len(conns))

	tts := time.Second
	if *connections > 100 {
		tts = time.Millisecond * 5
	}

	joinpkt := C_JoinChannel{}
	joinpkt.ChannelId = 0

	chatpkt := C_ChatMsg{}

	for i := 0; i < len(conns); i++ {
		time.Sleep(tts)
		conn := conns[i]
		joinpkt.Playerid = strconv.Itoa(i)
		sendpkt := MakeSendBuffer(100, joinpkt)
		Send(conn, sendpkt)
	}

	log.Printf("%d Client Join Channel Done !", len(conns))

	for {
		for i := 0; i < len(conns); i++ {
			time.Sleep(tts)
			conn := conns[i]
			chatpkt.Playerid = strconv.Itoa(i)
			chatpkt.Msg = "Hello World !!!"
			sendpkt := MakeSendBuffer(1000, chatpkt)
			Send(conn, sendpkt)
		}
	}
}

func Send(conn net.Conn, data []byte) {
	if conn != nil {
		sent, err := conn.Write(data)
		if err != nil {
			log.Println("SendPacket ERROR :", err)
		} else {
			if sent != len(data) {
				log.Println("[Sent diffrent size] : SENT =", sent, "BufferSize =", len(data))
			}
		}
	}
}

func MakeSendBuffer[T any](pktid uint16, data T) []byte {
	sendData, err := json.Marshal(&data)
	if err != nil {
		log.Println("MakeSendBuffer : Marshal Error", err)
	}
	sendBuffer := make([]byte, 4)

	pktsize := len(sendData) + 4

	binary.LittleEndian.PutUint16(sendBuffer, uint16(pktsize))
	binary.LittleEndian.PutUint16(sendBuffer[2:], pktid)

	sendBuffer = append(sendBuffer, sendData...)

	return sendBuffer
}
