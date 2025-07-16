package src

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const BroadcastPort = 55556

type NetMsg struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type RoomInfo struct {
	IP   string
	Port int
}

func HostGame() (net.Conn, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	go broadcastRoom(port)
	conn, err := ln.Accept()
	return conn, err

}

func broadcastRoom(port int) {
	bcastAddr := &net.UDPAddr{IP: net.IPv4bcast, Port: BroadcastPort}
	conn, _ := net.DialUDP("udp", nil, bcastAddr)
	defer conn.Close()
	msg := fmt.Sprintf("%s:%d", localIP(), port)
	for {
		conn.Write([]byte(msg))
		time.Sleep(1 * time.Second)
	}
}

func DiscoverRooms(timeout time.Duration) ([]RoomInfo, error) {
	sock, err := net.ListenUDP("udp", &net.UDPAddr{Port: BroadcastPort})
	if err != nil {
		return nil, err
	}
	defer sock.Close()
	sock.SetDeadline(time.Now().Add(timeout))

	rooms := map[string]RoomInfo{}
	buf := make([]byte, 64)
	for {
		n, addr, err := sock.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				break
			}
			return nil, err
		}
		parts := strings.Split(string(buf[:n]), ":")
		if len(parts) != 2 {
			continue
		}
		port, _ := strconv.Atoi(parts[1])
		rooms[addr.IP.String()] = RoomInfo{IP: addr.IP.String(), Port: port}
	}

	out := make([]RoomInfo, 0, len(rooms))
	for _, r := range rooms {
		out = append(out, r)
	}
	return out, nil
}
func JoinRoom(room RoomInfo) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", room.IP, room.Port)
	return net.Dial("tcp", addr)
}
func localIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}

func sendMove(conn net.Conn, row, col int) error {
	return json.NewEncoder(conn).Encode(NetMsg{Row: row, Col: col})
}

func recvMove(conn net.Conn) (int, int, error) {
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	var msg NetMsg
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		return 0, 0, err
	}

	// fmt.Printf("[RECV] received move: (%d, %d)\n", msg.Row, msg.Col)
	return msg.Row, msg.Col, nil
}
