package ws_server

import (
	"binance_test/internal/entities"
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	"github.com/rs/zerolog/log"
)

type Server struct {
	clients map[string]*websocket.Conn
}

func CreateServer() *Server {
	return &Server{
		clients: make(map[string]*websocket.Conn),
	}
}

// AddClient добавить клиента
func (s *Server) AddClient(id string, c *websocket.Conn) {
	s.clients[id] = c
	log.Debug().Msgf("Added client: %s", id)
}

// RemoveClient Удалить клиента
func (s *Server) RemoveClient(id string) {
	if _, ok := s.clients[id]; ok {
		delete(s.clients, id)
		log.Debug().Msgf("removed client: %s", id)
	}
}

// Read Чтение WebSocket клиента, просто, чтоб словить событие Close
func (s *Server) Read(c *websocket.Conn) {
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}

// WriteAll Отправить всем
func (s *Server) WriteAll(msg entities.Response) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	for _, client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			log.Error().Err(err).Msg("Write ws error")
		}
	}
}
