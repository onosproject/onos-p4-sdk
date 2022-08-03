// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import (
	"context"
	"github.com/google/uuid"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"math"
	"sync"
)

var log = logging.GetLogger()

// ConnManager p4rt connection manager
type ConnManager interface {
	Get(ctx context.Context, connID ConnID) (Conn, bool)
	GetByTarget(ctx context.Context, targetID topoapi.ID) (Client, error)
	Connect(ctx context.Context, target *topoapi.Object) error
	Disconnect(ctx context.Context, targetID topoapi.ID) error
	Watch(ctx context.Context, ch chan<- Conn) error
}

// NewConnManager creates a new p4rt connection manager
func NewConnManager() ConnManager {
	mgr := &connManager{
		targets:  make(map[topoapi.ID]*client),
		conns:    make(map[ConnID]Conn),
		watchers: make(map[uuid.UUID]chan<- Conn),
		eventCh:  make(chan Conn),
	}
	go mgr.processEvents()
	return mgr
}

type connManager struct {
	targets    map[topoapi.ID]*client
	conns      map[ConnID]Conn
	connsMu    sync.RWMutex
	watchers   map[uuid.UUID]chan<- Conn
	watchersMu sync.RWMutex
	eventCh    chan Conn
}

// Get returns a connection based on a given ID
func (m *connManager) Get(ctx context.Context, connID ConnID) (Conn, bool) {
	m.connsMu.RLock()
	defer m.connsMu.RUnlock()
	conn, ok := m.conns[connID]
	return conn, ok
}

// GetByTarget returns a P4RT client based on a given target ID
func (m *connManager) GetByTarget(ctx context.Context, targetID topoapi.ID) (Client, error) {
	m.connsMu.RLock()
	defer m.connsMu.RUnlock()
	if p4rtClient, ok := m.targets[targetID]; ok {
		return p4rtClient, nil
	}
	return nil, errors.NewNotFound("p4rt client for target %s not found", targetID)
}

// Connect makes a gRPC connection to the target
func (m *connManager) Connect(ctx context.Context, target *topoapi.Object) error {
	m.connsMu.RLock()
	p4rtClient, ok := m.targets[target.ID]
	m.connsMu.RUnlock()
	if ok {
		return errors.NewAlreadyExists("target '%s' already exists", target.ID)
	}

	m.connsMu.Lock()
	defer m.connsMu.Unlock()

	p4rtClient, ok = m.targets[target.ID]
	if ok {
		return errors.NewAlreadyExists("target '%s' already exists", target.ID)
	}

	if target.Type != topoapi.Object_ENTITY {
		return errors.NewInvalid("object is not a topo entity %v+", target)
	}

	typeKindID := string(target.GetEntity().KindID)
	if len(typeKindID) == 0 {
		return errors.NewInvalid("target entity %s must have a 'kindID' to work with ", target.ID)
	}

	destination, err := newDestination(target)
	if err != nil {
		log.Warnw("Failed to create a new target %s", "error", err)
		return err
	}

	log.Infow("Connecting to P4RT target", "target ID", destination.TargetID)
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
	}
	p4rtClient, clientConn, err := connect(ctx, *destination, opts...)
	if err != nil {
		log.Warnw("Failed to connect to the P4RT target %s: %s", "target ID", destination.TargetID, "error", err)
		return err
	}

	m.targets[target.ID] = p4rtClient
	streamChannel, err := p4rtClient.p4runtimeClient.StreamChannel(context.Background())
	if err != nil {
		log.Errorw("Cannot open a p4rt stream for connection", "targetID", target.ID, "error", err)
		return err
	}
	p4rtClient.streamClient.streamChannel = streamChannel
	go func() {
		var conn Conn
		state := clientConn.GetState()
		switch state {
		case connectivity.Ready:
			conn = newConn(target.ID, p4rtClient)

			m.addConn(conn)
		}

		for clientConn.WaitForStateChange(context.Background(), state) {
			state = clientConn.GetState()
			log.Infow("Connection state changed for Target", "target ID", target.ID, "state", state)

			// If the channel is active, ensure a connection is added to the manager.
			// If the channel is idle, do not change its state (this can occur when
			// connected or disconnected).
			// In all other states, remove the connection from the manager.
			switch state {
			case connectivity.Ready:
				if conn == nil {
					conn = newConn(target.ID, p4rtClient)
					streamChannel, err := p4rtClient.p4runtimeClient.StreamChannel(context.Background())
					if err != nil {
						log.Warnw("Cannot open a p4rt stream for connection", "targetID", target.ID, "error", err)
						continue
					}
					p4rtClient.streamClient.streamChannel = streamChannel
					m.addConn(conn)
				}

			case connectivity.Idle:
				clientConn.Connect()
			default:
				if conn != nil {
					log.Infow("Removing connection ", "connectionID", conn.ID())
					m.removeConn(conn.ID())
					conn = nil
				}
			}

			// If the channel is shutting down, exit the goroutine.
			switch state {
			case connectivity.Shutdown:
				return
			}
		}
	}()

	return nil
}

// Disconnect disconnects a gRPC connection based on a given target ID
func (m *connManager) Disconnect(ctx context.Context, targetID topoapi.ID) error {
	m.connsMu.Lock()
	clientConn, ok := m.targets[targetID]
	if !ok {
		m.connsMu.Unlock()
		return errors.NewNotFound("target '%s' not found", targetID)
	}
	delete(m.targets, targetID)
	m.connsMu.Unlock()
	return clientConn.Close()
}

func (m *connManager) Watch(ctx context.Context, ch chan<- Conn) error {
	id := uuid.New()
	m.watchersMu.Lock()
	m.connsMu.Lock()
	m.watchers[id] = ch
	m.watchersMu.Unlock()

	go func() {
		for _, conn := range m.conns {
			ch <- conn
		}
		m.connsMu.Unlock()

		<-ctx.Done()
		m.watchersMu.Lock()
		delete(m.watchers, id)
		m.watchersMu.Unlock()
	}()
	return nil
}

func (m *connManager) addConn(conn Conn) {
	m.connsMu.Lock()
	m.conns[conn.ID()] = conn
	m.connsMu.Unlock()
	m.eventCh <- conn
}

func (m *connManager) removeConn(connID ConnID) {
	m.connsMu.Lock()
	if conn, ok := m.conns[connID]; ok {
		delete(m.conns, connID)
		m.connsMu.Unlock()
		m.eventCh <- conn
	} else {
		m.connsMu.Unlock()
	}
}

func (m *connManager) processEvents() {
	for conn := range m.eventCh {
		m.processEvent(conn)
	}
}

func (m *connManager) processEvent(conn Conn) {
	log.Infow("Notifying P4RT connection", "connection ID", conn.ID())
	m.watchersMu.RLock()
	for _, watcher := range m.watchers {
		watcher <- conn
	}
	m.watchersMu.RUnlock()
}

func connect(ctx context.Context, d Destination, opts ...grpc.DialOption) (*client, *grpc.ClientConn, error) {
	switch d.TLS {
	case nil:
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	default:
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(d.TLS)))
	}

	// TODO handle credentials

	gCtx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	addr := ""
	if len(d.Addrs) != 0 {
		addr = d.Addrs[0]
	}
	conn, err := grpc.DialContext(gCtx, addr, opts...)
	if err != nil {
		return nil, nil, errors.NewInternal("Dialer(%s, %v): %v", addr, d.Timeout, err)
	}

	cl := p4v1.NewP4RuntimeClient(conn)

	p4rtClient := &client{
		grpcClient:      conn,
		p4runtimeClient: cl,
		writeClient: &writeClient{
			p4runtimeClient: cl,
		},
		readClient: &readClient{
			p4runtimeClient: cl,
		},
		pipelineConfigClient: &pipelineConfigClient{
			p4runtimeClient: cl,
		},
		streamClient: &streamClient{
			p4runtimeClient: cl,
		},
	}

	return p4rtClient, conn, nil
}
