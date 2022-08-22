// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package connection

import (
	"context"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/utils"
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

// NewController returns a new p4rt connection  controller
func NewController(topo topo.Store, conns southbound.ConnManager) *controller.Controller {
	c := controller.NewController("connection")
	c.Watch(&ConnWatcher{
		conns: conns,
	})
	c.Watch(&TopoWatcher{
		topo: topo,
	})
	c.Reconcile(&Reconciler{
		conns: conns,
		topo:  topo,
	})
	return c
}

// Reconciler reconciles gNMI connections
type Reconciler struct {
	conns southbound.ConnManager
	topo  topo.Store
}

// Reconcile reconciles a connection for a p4rt target
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	connID := id.Value.(southbound.ConnID)
	log.Infow("Reconciling Connection Relation", "connection ID", connID)
	conn, ok := r.conns.Get(ctx, connID)
	if !ok {
		return r.deleteRelation(ctx, connID)
	}
	return r.createRelation(ctx, conn)
}

func (r *Reconciler) createRelation(ctx context.Context, conn southbound.Conn) (controller.Result, error) {
	_, err := r.topo.Get(ctx, topoapi.ID(conn.ID()))
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed creating CONNECTION relation", "connection ID", conn.ID(), "error", err)
			return controller.Result{}, err
		}
		log.Infow("Creating CONNECTION relation", "connection relation ID", conn.ID())
		relation := &topoapi.Object{
			ID:   topoapi.ID(conn.ID()),
			Type: topoapi.Object_RELATION,
			Obj: &topoapi.Object_Relation{
				Relation: &topoapi.Relation{
					KindID:      topoapi.ConnectionKind,
					SrcEntityID: utils.GetControllerID(),
					TgtEntityID: utils.GetServiceID(conn.TargetID()),
				},
			},
		}
		err = r.topo.Create(ctx, relation)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Errorw("Failed creating CONNECTION relation", "connection relation ID", conn.ID(), "error", err)
				return controller.Result{}, err
			}
			log.Warnf("Failed creating CONNECTION relation", "connection relation ID", conn.ID(), "error", err)
			return controller.Result{}, nil
		}
		log.Infow("Connection Relation is created successfully", "connection relation ID", conn.ID())
	}
	return controller.Result{}, nil
}

func (r *Reconciler) deleteRelation(ctx context.Context, connID southbound.ConnID) (controller.Result, error) {
	relation, err := r.topo.Get(ctx, topoapi.ID(connID))
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed reconciling CONNECTION relation", "connection relation ID", connID, "error", err)
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}
	log.Infow("Deleting CONNECTION relation", "connection relation ID", connID)
	err = r.topo.Delete(ctx, relation)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed deleting CONNECTION relation", "connection relation ID", connID, "error", err)
			return controller.Result{}, err
		}
		log.Warnf("Failed deleting CONNECTION relation", "connection relation ID", connID, "error", err)
		return controller.Result{}, nil
	}
	return controller.Result{}, nil
}
