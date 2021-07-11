package tesla

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Subscribe struct {
	Type  string `json:"msg_type"`
	Token string `json:"token"`
	Value string `json:"value"`
	Tag   string `json:"tag"`
}

type Message struct {
	Type              string `json:"msg_type"` //
	ConnectionTimeout int64  `json:"connection_timeout"`
	Value             string `json:"value"`
	Tag               string `json:"tag"`
	ErrorType         string `json:"error_type"`
}

var (
	StreamingURL             = "wss://streaming.vn.teslamotors.com/streaming/"
	ClientError              = errors.New("client_error")
	DisconnectError          = errors.New("disconnect")
	VehicleDisconnectedError = errors.New("vehicle_disconnected")
)

var streamingCols = []string{
	"timestamp",
	"speed",
	"odometer",
	"soc",
	"elevation",
	"est_heading",
	"est_lat",
	"est_lng",
	"power",
	"shift_state",
	"range",
	"est_range",
	"heading"}

func (c *Client) StreamConnect(vehicleID uint64) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.DialContext(
		context.Background(),
		StreamingURL,
		nil)
	if err != nil {
		return nil, err
	}

	token, err := c.StreamToken()
	if err != nil {
		return nil, err
	}

	subMsg := Subscribe{
		Type:  "data:subscribe_oauth",
		Token: token,
		Value: strings.Join(streamingCols[1:], ","),
		Tag:   fmt.Sprintf("%d", vehicleID),
	}

	subData, _ := json.Marshal(subMsg)
	log.Printf("Sending: %s", string(subData))
	err = conn.WriteMessage(websocket.TextMessage, subData)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func (c *Client) Stream(vehicleID uint64, ch chan Message) error {
	var dataError error
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	conn, err := c.StreamConnect(vehicleID)
	if err != nil {
		return err
	}

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("stream read error: %s", err)
				dataError = DisconnectError
				break
			}
			log.Printf("stream received message: %s", message)

			m := Message{}
			err = json.Unmarshal(message, &m)
			if err != nil {
				log.Printf("unable to unmarshal message (%s): %s", message, m)
				continue
			}

			switch m.Type {
			case "control:hello":
			case "data:update":
				ch <- m
			case "data:error":
				switch m.ErrorType {
				case "client_error":
					dataError = ClientError
				case "vehicle_disconnected":
					dataError = VehicleDisconnectedError
				}
				// close(done)
				return
			default:
				log.Printf("Received unhandled message: %v", m)
			}
		}
	}()

	for {
		select {
		case <-done:
			return dataError
		case <-interrupt:
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return nil
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}
