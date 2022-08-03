// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"time"

	"github.com/onosproject/onos-lib-go/pkg/errors"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-p4-sdk/pkg/southbound"
	"github.com/onosproject/onos-p4-sdk/pkg/store/topo"
)

var log = logging.GetLogger()

const (
	defaultTimeout = 30 * time.Second
)

// NewController returns a new P4RT target  controller
func NewController(topo topo.Store, conns southbound.ConnManager) *controller.Controller {
	c := controller.NewController("target")
	c.Watch(&TopoWatcher{
		topo: topo,
	})
	c.Reconcile(&Reconciler{
		conns: conns,
		topo:  topo,
	})
	return c
}

// Reconciler reconciles P4RT connections
type Reconciler struct {
	conns southbound.ConnManager
	topo  topo.Store
}

// Reconcile reconciles a connection for a P4RT target
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	targetID := id.Value.(topoapi.ID)
	log.Infow("Reconciling Target", "Target ID", targetID)
	target, err := r.topo.Get(ctx, targetID)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorf("Failed reconciling Target", "Target ID", targetID, "error", err)
			return controller.Result{}, err
		}
		return r.disconnect(ctx, targetID)
	}
	return r.connect(ctx, target)
}

func (r *Reconciler) connect(ctx context.Context, target *topoapi.Object) (controller.Result, error) {
	log.Infof("Connecting to Target '%s'", target.ID)
	if err := r.conns.Connect(ctx, target); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Errorw("Failed connecting to Target", "Target ID", target.ID, "error", err)
			return controller.Result{}, err
		}
		log.Warnw("Failed connecting to Target", "Target ID", target.ID, "error", err)
		return controller.Result{}, nil
	}
	return controller.Result{}, nil
}

func (r *Reconciler) disconnect(ctx context.Context, targetID topoapi.ID) (controller.Result, error) {
	log.Infow("Disconnecting from Target", "target ID", targetID)
	if err := r.conns.Disconnect(ctx, targetID); err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed disconnecting from Target", "Target ID", targetID, "error", err)
			return controller.Result{}, err
		}
		log.Warnw("Failed disconnecting from Target", "Target ID", targetID, "error", err)
		return controller.Result{}, nil
	}
	return controller.Result{}, nil
}
