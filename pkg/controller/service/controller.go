// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"time"

	gogotypes "github.com/gogo/protobuf/types"

	"github.com/onosproject/onos-lib-go/pkg/errors"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"

	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/utils"
	"github.com/onosproject/onos-p4-sdk/pkg/store/topo"
)

const (
	defaultTimeout = 30 * time.Second
)

var log = logging.GetLogger()

// NewController returns a new service entity controller
func NewController(topo topo.Store) *controller.Controller {
	c := controller.NewController("service")
	c.Watch(&TopoWatcher{
		topo: topo,
	})

	c.Reconcile(&Reconciler{
		topo: topo,
	})

	return c
}

// Reconciler is a service entity reconciler
type Reconciler struct {
	topo topo.Store
}

func (r *Reconciler) createServiceEntity(ctx context.Context, serviceID topoapi.ID) error {
	log.Infow("Creating P4RT service entity", "service ID", serviceID)
	object := &topoapi.Object{
		ID:   utils.GetServiceID(),
		Type: topoapi.Object_ENTITY,
		Obj: &topoapi.Object_Entity{
			Entity: &topoapi.Entity{
				KindID: topoapi.ServiceKind,
			},
		},
		Aspects: make(map[string]*gogotypes.Any),
		Labels:  map[string]string{},
	}

	err := r.topo.Create(ctx, object)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Warnw("Creating service entity failed", "service ID", serviceID, "error", err)
		return err
	}

	return nil

}

// Reconcile reconciles the P4RT controller entities
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	serviceID := id.Value.(topoapi.ID)
	log.Infow("Reconciling service entity", "service ID", serviceID)
	_, err := r.topo.Get(ctx, serviceID)
	if err == nil {
		log.Infow("Service entity exists already", "serviceID", serviceID)
		return controller.Result{}, nil
	} else if !errors.IsNotFound(err) {
		log.Warnw("Failed to reconcile service entity", "service ID", serviceID, "error", err)
		return controller.Result{}, err
	}

	// Create the service entity
	if err := r.createServiceEntity(ctx, serviceID); err != nil {
		log.Warnw("Failed to create the service entity", "service ID", serviceID, "error", err)
		return controller.Result{}, err
	}

	return controller.Result{}, nil
}
