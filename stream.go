package tesla

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"strconv"
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
	StreamingURL           = "wss://streaming.vn.teslamotors.com/streaming/"
	ErrClient              = errors.New("client_error")
	ErrDisconnect          = errors.New("disconnect")
	ErrVehicleDisconnected = errors.New("vehicle_disconnected")
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

func (c *Client) streamConnect(vehicleID uint64, params string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.DialContext(
		context.Background(),
		StreamingURL,
		nil)
	if err != nil {
		return nil, err
	}

	token, err := c.Token()
	if err != nil {
		return nil, err
	}

	subMsg := Subscribe{
		Type:  "data:subscribe_oauth",
		Token: token.AccessToken,
		Value: params,
		Tag:   strconv.FormatUint(vehicleID, 10),
	}

	subData, err := json.Marshal(subMsg)
	if err != nil {
		return nil, err
	}

	err = conn.WriteMessage(websocket.TextMessage, subData)

	return conn, err
}

func (c *Client) Stream(vehicleID uint64, ch chan Message, params ...string) error {
	var dataError error
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	if len(params) == 0 {
		params = streamingCols[1:]
	}

	conn, err := c.streamConnect(vehicleID, strings.Join(params, ","))
	if err != nil {
		return err
	}
	defer conn.Close()

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				dataError = ErrDisconnect
				return
			}

			m := Message{}
			if err := json.Unmarshal(message, &m); err != nil {
				continue
			}

			switch m.Type {
			case "control:hello":
			case "data:update":
				ch <- m
			case "data:error":
				switch m.ErrorType {
				case "client_error":
					dataError = ErrClient
				case "vehicle_disconnected":
					dataError = ErrVehicleDisconnected
				}
				return
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
