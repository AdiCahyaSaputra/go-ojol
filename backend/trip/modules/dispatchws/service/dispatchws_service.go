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

type OfferRetrier interface {
	RetryFindDriver(ctx context.Context, customerUserID string) error
}

type DispatchWSService interface {
	HandleConn(userID string, conn *websocket.Conn)
	HandleCustomerConn(userID string, conn *websocket.Conn)
	SetOfferRetrier(retrier OfferRetrier)
	Notify(userID string, msg dto.ServerMessage) bool
	NotifyMany(userIDs []string, msg dto.ServerMessage)
}

type wsConnection struct {
	socket     *websocket.Conn
	writeMutex sync.Mutex
}

type dispatchWSService struct {
	locations           *drivergeo.Store
	retrier             OfferRetrier
	mutex               sync.Mutex
	driverConnections   map[string]*wsConnection
	customerConnections map[string]*wsConnection
}

func NewDispatchWSService(locations *drivergeo.Store) DispatchWSService {
	return &dispatchWSService{
		locations:           locations,
		driverConnections:   make(map[string]*wsConnection),
		customerConnections: make(map[string]*wsConnection),
	}
}

func (s *dispatchWSService) SetOfferRetrier(retrier OfferRetrier) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.retrier = retrier
}

func (s *dispatchWSService) Notify(userID string, msg dto.ServerMessage) bool {
	s.mutex.Lock()
	connection := s.driverConnections[userID]
	if connection == nil {
		connection = s.customerConnections[userID]
	}
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

func (s *dispatchWSService) writeJSON(connection *wsConnection, msg dto.ServerMessage) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()
	_ = connection.socket.SetWriteDeadline(time.Now().Add(writeWait))
	return connection.socket.WriteJSON(msg)
}

func (s *dispatchWSService) HandleConn(userID string, conn *websocket.Conn) {
	connection := &wsConnection{socket: conn}
	if old := s.registerDriver(userID, connection); old != nil {
		_ = old.socket.Close()
	}

	defer func() {
		if s.unregisterDriver(userID, connection) && s.locations != nil {
			_ = s.locations.RemoveStandby(context.Background(), userID)
		}
		_ = conn.Close()
	}()

	s.runPingLoop(conn, connection, func() {
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
	})
}

func (s *dispatchWSService) HandleCustomerConn(userID string, conn *websocket.Conn) {
	connection := &wsConnection{socket: conn}
	if old := s.registerCustomer(userID, connection); old != nil {
		_ = old.socket.Close()
	}

	defer func() {
		_ = s.unregisterCustomer(userID, connection)
		_ = conn.Close()
	}()

	s.runPingLoop(conn, connection, func() {
		for {
			var msg dto.ClientMessage
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}

			switch msg.Type {
			case dto.TypeRetry:
				s.mutex.Lock()
				retrier := s.retrier
				s.mutex.Unlock()
				if retrier == nil {
					_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_RETRY_UNAVAILABLE})
					continue
				}
				if err := retrier.RetryFindDriver(context.Background(), userID); err != nil {
					_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeError, Message: err.Error()})
				}
			default:
				_ = s.writeJSON(connection, dto.ServerMessage{Type: dto.TypeError, Message: dto.MESSAGE_UNKNOWN_MESSAGE_TYPE})
			}
		}
	})
}

func (s *dispatchWSService) runPingLoop(conn *websocket.Conn, connection *wsConnection, readLoop func()) {
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

	readLoop()
}

func (s *dispatchWSService) registerDriver(userID string, connection *wsConnection) *wsConnection {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	old := s.driverConnections[userID]
	s.driverConnections[userID] = connection
	return old
}

func (s *dispatchWSService) unregisterDriver(userID string, connection *wsConnection) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.driverConnections[userID] != connection {
		return false
	}
	delete(s.driverConnections, userID)
	return true
}

func (s *dispatchWSService) registerCustomer(userID string, connection *wsConnection) *wsConnection {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	old := s.customerConnections[userID]
	s.customerConnections[userID] = connection
	return old
}

func (s *dispatchWSService) unregisterCustomer(userID string, connection *wsConnection) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.customerConnections[userID] != connection {
		return false
	}
	delete(s.customerConnections, userID)
	return true
}

func validLatLng(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
