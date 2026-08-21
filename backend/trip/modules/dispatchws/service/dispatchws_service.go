package service

import (
	"context"
	"sync"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"github.com/gorilla/websocket"
)

const (
	pongWait     = 60 * time.Second
	pingInterval = 30 * time.Second
	writeWait    = 10 * time.Second
)

type DispatchWSService interface {
	HandleConn(userID string, conn *websocket.Conn)
}

type dispatchWSService struct {
	locations *drivergeo.Store
	mu        sync.Mutex
	conns     map[string]*websocket.Conn
}

func NewDispatchWSService(locations *drivergeo.Store) DispatchWSService {
	return &dispatchWSService{
		locations: locations,
		conns:     make(map[string]*websocket.Conn),
	}
}

func (s *dispatchWSService) HandleConn(userID string, conn *websocket.Conn) {
	if old := s.register(userID, conn); old != nil {
		_ = old.Close()
	}

	defer func() {
		if s.unregister(userID, conn) && s.locations != nil {
			_ = s.locations.RemoveStandby(context.Background(), userID)
		}
		_ = conn.Close()
	}()

	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	var writeMu sync.Mutex
	writeJSON := func(msg dto.ServerMessage) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteJSON(msg)
	}

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writeMu.Lock()
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				err := conn.WriteMessage(websocket.PingMessage, nil)
				writeMu.Unlock()
				if err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	for {
		var msg dto.ClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}

		switch msg.Type {
		case dto.TypeStandby, dto.TypeLocation:
			if !validLatLng(msg.Lat, msg.Lng) {
				_ = writeJSON(dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_INVALID_LAT_LONG})
				continue
			}
			if s.locations == nil {
				_ = writeJSON(dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_LOCATION_STORE_UNAVAILABLE})
				continue
			}
			if err := s.locations.SetStandby(context.Background(), userID, msg.Lat, msg.Lng); err != nil {
				_ = writeJSON(dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_FAILED_SAVE_LOC})
				continue
			}
			if msg.Type == dto.TypeStandby {
				_ = writeJSON(dto.ServerMessage{Type: dto.TypeStandbyOK})
			}
		default:
			_ = writeJSON(dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_UNKNOWN_MESSAGE_TYPE})
		}
	}
}

func (s *dispatchWSService) register(userID string, conn *websocket.Conn) *websocket.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.conns[userID]
	s.conns[userID] = conn
	return old
}

func (s *dispatchWSService) unregister(userID string, conn *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns[userID] != conn {
		return false
	}
	delete(s.conns, userID)
	return true
}

func validLatLng(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
