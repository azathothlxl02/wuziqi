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

type UndoRequestMsg struct {
	Undo bool `json:"undo"`
}

type UndoAcceptMsg struct {
	UndoAccept bool `json:"undoAccept"`
}

type UndoRejectMsg struct {
	UndoReject bool `json:"undoReject"`
}

func sendUndoRequest(conn net.Conn) error {
	return json.NewEncoder(conn).Encode(UndoRequestMsg{Undo: true})
}

func sendUndoAccept(conn net.Conn) error {
	return json.NewEncoder(conn).Encode(UndoAcceptMsg{UndoAccept: true})
}

func sendUndoReject(conn net.Conn) error {
	return json.NewEncoder(conn).Encode(UndoRejectMsg{UndoReject: true})
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
	addr := net.JoinHostPort(room.IP, strconv.Itoa(room.Port))
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
	fmt.Printf("[SEND] Row=%d Col=%d\n", row, col)
	return json.NewEncoder(conn).Encode(NetMsg{Row: row, Col: col})
}

func recvMessage(conn net.Conn) (int, int, string, error) {
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	var raw json.RawMessage
	if err := json.NewDecoder(conn).Decode(&raw); err != nil {
		return 0, 0, "PEER_LEFT", err
	}

	var undoReq UndoRequestMsg
	if json.Unmarshal(raw, &undoReq) == nil && undoReq.Undo {
		return 0, 0, "UNDO_REQUEST", nil
	}
	var undoAcc UndoAcceptMsg
	if json.Unmarshal(raw, &undoAcc) == nil && undoAcc.UndoAccept {
		return 0, 0, "UNDO_ACCEPT", nil
	}
	var undoRej UndoRejectMsg
	if json.Unmarshal(raw, &undoRej) == nil && undoRej.UndoReject {
		return 0, 0, "UNDO_REJECT", nil
	}

	var move NetMsg
	if json.Unmarshal(raw, &move) == nil {
		return move.Row, move.Col, "MOVE", nil
	}

	return 0, 0, "PEER_LEFT", nil
}
