// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	gogotypes "github.com/gogo/protobuf/types"
	"time"

	"github.com/onosproject/onos-lib-go/pkg/errors"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/utils"
	"github.com/onosproject/onos-p4-sdk/pkg/store/topo"
)

var log = logging.GetLogger()

const (
	defaultTimeout = 30 * time.Second
)

// NewController returns a new P4RT target  controller
func NewController(topo topo.Store) *controller.Controller {
	c := controller.NewController("target")
	c.Watch(&TopoWatcher{
		topo: topo,
	})
	c.Reconcile(&Reconciler{
		topo: topo,
	})
	return c
}

// Reconciler reconciles P4RT connections
type Reconciler struct {
	topo topo.Store
}

func (r *Reconciler) createServiceEntity(ctx context.Context, targetID topoapi.ID) (controller.Result, error) {
	serviceEntityID := utils.GetServiceID(targetID)
	log.Infow("Creating service entity", "service entity ID", serviceEntityID)
	object := &topoapi.Object{
		ID:   serviceEntityID,
		Type: topoapi.Object_ENTITY,
		Obj: &topoapi.Object_Entity{
			Entity: &topoapi.Entity{
				KindID: topoapi.ServiceKind,
			},
		},
		Aspects: make(map[string]*gogotypes.Any),
		Labels:  map[string]string{},
	}

	serviceAspect := &topoapi.Service{
		TargetID:        string(targetID),
		Mastershipstate: &topoapi.P4RTMastershipState{},
	}

	err := object.SetAspect(serviceAspect)
	if err != nil {
		return controller.Result{}, err
	}

	err = r.topo.Create(ctx, object)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Warnw("Creating service entity failed", "service entity ID", serviceEntityID, "error", err)
		return controller.Result{}, err
	}

	return controller.Result{}, nil

}

func (r *Reconciler) deleteServiceEntity(ctx context.Context, targetID topoapi.ID) (controller.Result, error) {
	serviceEntityID := utils.GetServiceID(targetID)
	log.Infow("Deleting service entity", "service entity ID", serviceEntityID, "targetID", targetID)
	serviceEntity, err := r.topo.Get(ctx, serviceEntityID)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed retrieving service entity", "service entity ID", serviceEntityID, "targetID", targetID, "error", err)
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}

	err = r.topo.Delete(ctx, serviceEntity)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed deleting service entity", "service entity ID", serviceEntityID, "targetID", targetID, "error", err)
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}
	return controller.Result{}, nil

}

// Reconcile reconciles a connection for a P4RT target
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	targetID := id.Value.(topoapi.ID)
	log.Infow("Reconciling service entity for target", "Target ID", targetID)
	_, err := r.topo.Get(ctx, targetID)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorf("Failed reconciling service entity for target", "Target ID", targetID, "error", err)
			return controller.Result{}, err
		}
		return r.deleteServiceEntity(ctx, targetID)
	}

	return r.createServiceEntity(ctx, targetID)

}
