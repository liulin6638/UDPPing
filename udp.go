package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
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

func server() {
	// 创建监听
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9999,
	})
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

func client(remote string) {
	// 创建连接
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.ParseIP(remote),
		Port: 9999,
	})
	if err != nil {
		fmt.Println("连接失败!", err)
		return
	}
	defer socket.Close()
	seq := 1
	for {
		packet := NewPacket(seq)
		seq++
		// 发送数据
		buff, err := packet.Encode()
		_, err = socket.Write(buff)
		if err != nil {
			fmt.Println("发送数据失败!", err)
			return
		}

		fmt.Println(time.Now().UTC(), "sent ping")
		// 接收数据
		socket.SetReadDeadline(time.Now().Add(time.Second))
		data := make([]byte, 4096)
		read, _, err := socket.ReadFromUDP(data)
		if err != nil {
			fmt.Println(time.Now().UTC(), "recv ping ", remote, packet.Seq, "time out")
			continue
		}

		response, err := Decode(data[:read])
		if err != nil {
			fmt.Println("解析数据失败!", err)
			os.Exit(0)
		}
		fmt.Println(time.Now().UTC(), "ping ", remote, response.Seq, (time.Now().UnixNano()/1e6 - response.Ts))
		time.Sleep(time.Second)
	}
}

func main() {
	isServer := true
	remoteAddr := ""
	args := os.Args

	// client("192.168.137.1")
	for index, value := range args {
		if index == 1 && value == "-c" {
			isServer = false
		}
		if index == 2 && isServer == false {
			remoteAddr = value
		}
	}

	if isServer {
		server()
	} else {
		client(remoteAddr)
	}

}
