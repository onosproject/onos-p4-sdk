// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"sync"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"

	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-p4-sdk/pkg/store/topo"
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
			if _, ok := event.Object.Obj.(*topoapi.Object_Entity); ok {
				log.Debugw("Event entity", "entity", event.Object)
				err = event.Object.GetAspect(&topoapi.P4RTServerInfo{})
				if err == nil {
					ch <- controller.NewID(event.Object.ID)
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
