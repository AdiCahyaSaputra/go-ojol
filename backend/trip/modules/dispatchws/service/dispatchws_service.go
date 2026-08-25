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
	Notify(userID string, msg dto.ServerMessage) bool
	NotifyMany(userIDs []string, msg dto.ServerMessage)
}

type driverConnection struct {
	socket     *websocket.Conn
	writeMutex sync.Mutex
}

type dispatchWSService struct {
	locations         *drivergeo.Store
	mutex             sync.Mutex
	driverConnections map[string]*driverConnection
}

func NewDispatchWSService(locations *drivergeo.Store) DispatchWSService {
	return &dispatchWSService{
		locations:         locations,
		driverConnections: make(map[string]*driverConnection),
	}
}

func (s *dispatchWSService) Notify(userID string, msg dto.ServerMessage) bool {
	s.mutex.Lock()
	connection := s.driverConnections[userID]
	s.mutex.Unlock()
	if connection == nil {
		return false
	}
	return s.writeJSON(connection, msg) == nil
}

func (s *dispatchWSService) NotifyMany(userIDs []string, msg dto.ServerMessage) {
	for _, userID := range userIDs {
		_ = s.Notify(userID, msg)
	}
}

func (s *dispatchWSService) writeJSON(connection *driverConnection, msg dto.ServerMessage) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()
	_ = connection.socket.SetWriteDeadline(time.Now().Add(writeWait))
	return connection.socket.WriteJSON(msg)
}

func (s *dispatchWSService) HandleConn(userID string, conn *websocket.Conn) {
	connection := &driverConnection{socket: conn}
	if old := s.register(userID, connection); old != nil {
		_ = old.socket.Close()
	}

	defer func() {
		if s.unregister(userID, connection) && s.locations != nil {
			_ = s.locations.RemoveStandby(context.Background(), userID)
		}
		_ = conn.Close()
	}()

	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				connection.writeMutex.Lock()
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				err := conn.WriteMessage(websocket.PingMessage, nil)
				connection.writeMutex.Unlock()
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
				_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_INVALID_LAT_LONG})
				continue
			}
			if s.locations == nil {
				_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_LOCATION_STORE_UNAVAILABLE})
				continue
			}
			if err := s.locations.SetStandby(context.Background(), userID, msg.Lat, msg.Lng); err != nil {
				_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_FAILED_SAVE_LOC})
				continue
			}
			if msg.Type == dto.TypeStandby {
				_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeStandbyOK})
			}
		default:
			_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_UNKNOWN_MESSAGE_TYPE})
		}
	}
}

func (s *dispatchWSService) register(userID string, connection *driverConnection) *driverConnection {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	old := s.driverConnections[userID]
	s.driverConnections[userID] = connection
	return old
}

func (s *dispatchWSService) unregister(userID string, connection *driverConnection) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.driverConnections[userID] != connection {
		return false
	}
	delete(s.driverConnections, userID)
	return true
}

func validLatLng(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
