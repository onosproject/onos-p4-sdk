// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

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
	defaultTimeout            = 30 * time.Second
	defaultExpirationDuration = 30 * time.Second
)

var log = logging.GetLogger()

// NewController returns a new node controller
func NewController(topo topo.Store) *controller.Controller {
	c := controller.NewController("node")
	c.Watch(&TopoWatcher{
		topo: topo,
	})

	c.Reconcile(&Reconciler{
		topo: topo,
	})

	return c
}

// Reconciler is a controller node reconciler
type Reconciler struct {
	topo topo.Store
}

func (r *Reconciler) createControllerEntity(ctx context.Context, controllerID topoapi.ID) error {
	log.Infow("Creating P4RT controller entity", "controller ID", controllerID)
	object := &topoapi.Object{
		ID:   utils.GetControllerID(),
		Type: topoapi.Object_ENTITY,
		Obj: &topoapi.Object_Entity{
			Entity: &topoapi.Entity{
				KindID: topoapi.ControllerKind,
			},
		},
		Aspects: make(map[string]*gogotypes.Any),
		Labels:  map[string]string{},
	}

	expiration := time.Now().Add(defaultExpirationDuration)
	leaseAspect := &topoapi.Lease{
		Expiration: &expiration,
	}
	controllerAspect := &topoapi.ControllerInfo{
		Role: &topoapi.ControllerRole{
			Name: "",
		},
		Type: topoapi.ControllerInfo_P4RUNTIME,
	}

	err := object.SetAspect(leaseAspect)
	if err != nil {
		return err
	}
	err = object.SetAspect(controllerAspect)
	if err != nil {
		return err
	}

	err = r.topo.Create(ctx, object)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Warnw("Creating controller entity failed", "controller ID", controllerID, "error", err)
		return err
	}

	return nil

}

// Reconcile reconciles the P4RT controller entities
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	controllerID := id.Value.(topoapi.ID)
	log.Infow("Reconciling controller entity", "controller ID", controllerID)
	object, err := r.topo.Get(ctx, controllerID)
	if err == nil {
		//  Reconciles a controller entity that’s not local so the controller should requeue
		//  it for the lease expiration time and delete the entity if the lease has not been renewed
		if controllerID != utils.GetControllerID() {
			lease := &topoapi.Lease{}
			_ = object.GetAspect(lease)

			// Check if the lease is expired
			if lease.Expiration.Before(time.Now()) {
				log.Infow("Deleting the expired lease for controller", "controller ID", controllerID)
				err := r.topo.Delete(ctx, object)
				if !errors.IsNotFound(err) {
					log.Warnw("Deleting the expired lease for controller failed", "controller ID", controllerID)
					return controller.Result{}, err
				}
				return controller.Result{}, nil
			}

			// Requeue the object to be reconciled at the expiration time
			return controller.Result{
				RequeueAfter: time.Until(*lease.Expiration),
			}, nil
		}

		// Renew the lease If this is the controller entity for the local node
		if controllerID == utils.GetControllerID() {
			lease := &topoapi.Lease{}

			err := object.GetAspect(lease)
			if err != nil {
				log.Warnw("Failed to renew the lease for controller", "controller ID", controllerID, "error", err)
				return controller.Result{}, err
			}

			remainingTime := time.Until(*lease.GetExpiration())
			// If the remaining time of lease is more than  half the lease duration, no need to renew the lease
			// schedule the next renewal
			if remainingTime > defaultExpirationDuration/2 {
				log.Debugw("No need to renew the lease", "controller ID", controllerID, "remaining lease time (ms)", remainingTime.Milliseconds())
				return controller.Result{
					RequeueAfter: time.Until(lease.Expiration.Add(defaultExpirationDuration / 2 * -1)),
				}, nil
			}

			// Renew the release to trigger the reconciler
			log.Debugw("Renewing the lease for controller", "controller ID", controllerID)
			expiration := time.Now().Add(defaultExpirationDuration)
			lease = &topoapi.Lease{
				Expiration: &expiration,
			}

			err = object.SetAspect(lease)
			if err != nil {
				return controller.Result{}, err
			}
			err = r.topo.Update(ctx, object)
			if err != nil && !errors.IsNotFound(err) {
				log.Warnw("Failed to renew the lease for controller", "controller ID", controllerID, "error", err)
				return controller.Result{}, err
			}
			return controller.Result{}, nil
		}

	} else if !errors.IsNotFound(err) {
		log.Warnw("Failed to renew the lease for controller", "controller ID", controllerID, "error", err)
		return controller.Result{}, err
	}

	// Create the P4RT controller entity
	if err := r.createControllerEntity(ctx, controllerID); err != nil {
		log.Warnw("Failed to create the controller entity", "controller ID", controllerID, "error", err)
		return controller.Result{}, err
	}

	return controller.Result{}, nil

}
