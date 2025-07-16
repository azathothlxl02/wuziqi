package src

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

const Port = ":55555"

type NetMsg struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

func hostGame() (net.Conn, error) {
	ln, err := net.Listen("tcp", Port)
	if err != nil {
		return nil, err
	}
	fmt.Println("Waiting for opponent...")
	return ln.Accept()
}

func joinGame(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr+Port)
}

func sendMove(conn net.Conn, row, col int) error {
	return json.NewEncoder(conn).Encode(NetMsg{Row: row, Col: col})
}

func recvMove(conn net.Conn) (int, int, error) {
	var msg NetMsg
	err := json.NewDecoder(conn).Decode(&msg)
	return msg.Row, msg.Col, err
}

func GetLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip := ipnet.IP.To4()
			if ip != nil && strings.HasPrefix(ip.String(), "192.168") || strings.HasPrefix(ip.String(), "10.") {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}
