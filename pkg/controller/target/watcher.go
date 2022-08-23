// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"sync"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"

	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-p4-sdk/pkg/store/topo"

	"github.com/onosproject/onos-p4-sdk/pkg/southbound"
)

const queueSize = 100

// TopoWatcher is a topology watcher
type TopoWatcher struct {
	topo   topo.Store
	cancel context.CancelFunc
	mu     sync.Mutex
}

// Start starts the topo store watcher
func (w *TopoWatcher) Start(ch chan<- controller.ID) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		return nil
	}

	eventCh := make(chan topoapi.Event, queueSize)
	ctx, cancel := context.WithCancel(context.Background())

	err := w.topo.Watch(ctx, eventCh, nil)
	if err != nil {
		cancel()
		return err
	}
	w.cancel = cancel
	go func() {
		for event := range eventCh {
			log.Debugw("Received topo event", "topo object ID", event.Object.ID)
			if entity, ok := event.Object.Obj.(*topoapi.Object_Entity); ok {
				log.Debugw("Event entity", "entity", event.Object)
				if entity.Entity.KindID == topoapi.ServiceKind {
					serviceAspect := &topoapi.Service{}
					err = event.Object.GetAspect(serviceAspect)
					if err == nil {
						ch <- controller.NewID(topoapi.ID(serviceAspect.TargetID))
					}
				}
			}
		}
	}()
	return nil
}

// Stop stops the topology watcher
func (w *TopoWatcher) Stop() {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	w.mu.Unlock()
}

// ConnWatcher is a P4RT connection watcher
type ConnWatcher struct {
	conns  southbound.ConnManager
	cancel context.CancelFunc
	mu     sync.Mutex
	connCh chan southbound.Conn
}

// Start starts the connection watcher
func (c *ConnWatcher) Start(ch chan<- controller.ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		return nil
	}

	c.connCh = make(chan southbound.Conn, queueSize)
	ctx, cancel := context.WithCancel(context.Background())
	err := c.conns.Watch(ctx, c.connCh)
	if err != nil {
		cancel()
		return err
	}
	c.cancel = cancel

	go func() {
		for conn := range c.connCh {
			log.Debugw("Received P4RT Connection event", "connection ID", conn.ID())
			ch <- controller.NewID(conn.TargetID())
		}
		close(ch)
	}()
	return nil
}

// Stop stops the connection watcher
func (c *ConnWatcher) Stop() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.mu.Unlock()
}
