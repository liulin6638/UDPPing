package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type PingPacket struct {
	Seq int32
	Ts  int64
}

func NewPacket(seq int) *PingPacket {
	obj := &PingPacket{}
	obj.Seq = int32(seq)
	obj.Ts = time.Now().UnixNano() / 1e6
	return obj
}

func (obj *PingPacket) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, obj); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decode(b []byte) (*PingPacket, error) {
	buf := bytes.NewBuffer(b)

	obj := &PingPacket{}

	if err := binary.Read(buf, binary.LittleEndian, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func server(port int) {
	// 创建监听
	udpAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:"+strconv.Itoa(port))
	socket, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		fmt.Println("监听失败!", err)
		return
	}
	defer socket.Close()
	fmt.Println("Start Server ", socket.LocalAddr().String())
	for {
		// 读取数据
		data := make([]byte, 4096)
		read, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}
		fmt.Println(read, remoteAddr)

		_, err = socket.WriteToUDP(data[:read], remoteAddr)
		if err != nil {
			return
			fmt.Println("发送数据失败!", err)
		}
	}
}

func client(ip string, port int) {
	// 创建连接
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})
	if err != nil {
		fmt.Println("连接失败!", err)
		return
	}
	defer socket.Close()
	seq := 1
	go func(socket *net.UDPConn) {
		for {
			socket.SetReadDeadline(time.Now().Add(time.Second))
			data := make([]byte, 4096)
			read, _, err := socket.ReadFromUDP(data)
			if err != nil {
				// fmt.Println(time.Now().UTC(), "recv ping ", socket.RemoteAddr().String(), "time out")
				continue
			}

			response, err := Decode(data[:read])
			if err != nil {
				fmt.Println("解析数据失败!", err)
				os.Exit(0)
			}
			fmt.Println(time.Now().UTC(), "ping ", ip, ":", port, response.Seq, (time.Now().UnixNano()/1e6 - response.Ts))
		}
	}(socket)
	for {
		// fmt.Println(time.Now().UTC(), "New loop start write")
		packet := NewPacket(seq)
		seq++
		// 发送数据
		buff, err := packet.Encode()
		_, err = socket.Write(buff)
		if err != nil {
			fmt.Println("发送数据失败!", err)
			return
		}

		// fmt.Println(time.Now().UTC(), "sent ping write over start recv ")
		// 接收数据

		time.Sleep(time.Second)
	}
}

func main() {
	isServer := 0
	serverAddr := ""
	serverPort := 0
	serverAddrIndex := 0
	serverPortIndex := 0
	isHelp := false
	args := os.Args

	for index, value := range args {
		if value == "-h" || value == "?" || value == "--help" {
			isHelp = true
			continue
		}

		if value == "-c" {
			isServer = 1
			serverAddrIndex = index
			continue
		}
		if isServer == 1 && index-1 == serverAddrIndex {
			serverAddr = value
			continue
		}

		if value == "-p" && isServer == 1 {
			serverPortIndex = index
			continue
		}
		if isServer == 1 && index-1 == serverPortIndex {
			serverPort, _ = strconv.Atoi(value)
			continue
		}

		if value == "-s" && isServer == 0 {
			isServer = 2
			serverPortIndex = index
			continue
		}
		if isServer == 2 && index-1 == serverPortIndex {
			serverPort, _ = strconv.Atoi(value)
			continue
		}
	}

	if isHelp {
		fmt.Println("As Client -c serverip:port or -c serverip -p port")
		fmt.Println("As Server -s port")
		return
	}

	if serverPort == 0 && isServer == 1 {
		str := strings.Split(serverAddr, ":")
		serverAddr = str[0]
		serverPort, _ = strconv.Atoi(str[1])
	}

	if isServer == 2 {
		server(serverPort)
	} else if isServer == 1 {
		client(serverAddr, serverPort)
	} else {
		fmt.Println("As Client -c serverip:port or -c serverip -p port")
		fmt.Println("As Server -s port")
		return
	}

}
