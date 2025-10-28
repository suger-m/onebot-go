package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	types "onebot-go2/pkg/const"
	"onebot-go2/pkg/event"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSServer struct {
	token        string
	clients      map[*websocket.Conn]bool
	upgrader     websocket.Upgrader
	pendingCalls sync.Map
	dispatcher   *event.Dispatcher
}

func NewWSServer(token string) *WSServer {
	return &WSServer{
		token:      token,
		clients:    make(map[*websocket.Conn]bool),
		dispatcher: event.NewDispatcher(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// GetDispatcher 获取事件分发器
func (s *WSServer) GetDispatcher() *event.Dispatcher {
	return s.dispatcher
}

func (s *WSServer) HandlerWebsocket(c *gin.Context) {

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrader error %v", err)
		return
	}
	s.addClient(conn)
	defer s.removeClient(conn)
	log.Printf("WebSocket connention from %s", conn.RemoteAddr())

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("err reading msg :%v", conn.RemoteAddr())
			break
		}

		var response struct {
			Echo string
			types.APIResponse
		}
		if err := json.Unmarshal(message, &response); err == nil && response.Echo != "" {
			if ch, ok := s.pendingCalls.LoadAndDelete(response.Echo); ok {
				if respChan, ok := ch.(chan *types.APIResponse); ok {
					select {
					case respChan <- &response.APIResponse:
					case <-time.After(1 * time.Second):
						log.Printf("Response channel timeout")
					}
					close(respChan)
				}
			}
			continue
		}
		evt, err := ParseEvent(message)
		if err != nil {
			log.Printf("Parse Event Error %v", err)
			continue
		}

		// 分发事件到注册的处理器
		if err := s.dispatcher.Dispatch(context.Background(), evt); err != nil {
			log.Printf("Error dispatching event: %v", err)
		}
	}
}

func ParseEvent(data []byte) (interface{}, error) {
	var base types.Event
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.PostType {
	case types.PostTypeMessage:
		var event types.MessageEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil

	case types.PostTypeNotice:
		var event types.NoticeEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil

	case types.PostTypeRequest:
		var event types.RequestEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil
	case types.PostTypeMetaEvent:
		var event types.NoticeEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil

	default:
		return nil, fmt.Errorf("unknown post_type: %v", base.PostType)

	}

}

func (s *WSServer) addClient(conn *websocket.Conn) {
	s.clients[conn] = true
}

func (s *WSServer) removeClient(conn *websocket.Conn) {
	delete(s.clients, conn)
	conn.Close()
}
