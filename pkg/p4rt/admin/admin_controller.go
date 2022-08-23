// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-lib-go/pkg/env"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/connection"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/mastership"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/node"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/service"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/target"
	controllerutils "github.com/onosproject/onos-p4-sdk/pkg/controller/utils"
	"github.com/onosproject/onos-p4-sdk/pkg/southbound"
	"github.com/onosproject/onos-p4-sdk/pkg/store/topo"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
)

var log = logging.GetLogger()

// Config is a manager
type Config struct {
	CAPath      string
	KeyPath     string
	CertPath    string
	TopoAddress string
}

// Controller p4rt controller
type Controller struct {
	Config    Config
	conns     southbound.ConnManager
	topoStore topo.Store
}

// StartController creates a new P4RTController instance and starts it
func StartController(cfg Config) *Controller {
	p4rtController := &Controller{
		Config: cfg,
	}
	p4rtController.run()
	return p4rtController
}

// run runs p4rt controller
func (c *Controller) run() {
	log.Infow("Starting P4RT Controller ")
	if err := c.start(); err != nil {
		log.Fatalw("Unable to run P4Rt controller", "error", err)
	}
}

// Client returns a master client for the given target
func (c *Controller) Client(ctx context.Context, targetID topoapi.ID) (Client, *topoapi.Object, *topoapi.Object, error) {

	targetEntity, err := c.topoStore.Get(ctx, targetID)
	if err != nil {
		return nil, nil, nil, err
	}
	serviceEntity, err := c.topoStore.Get(ctx, controllerutils.GetServiceID(targetID))
	if err != nil {
		return nil, nil, nil, err
	}

	serviceAspect := &topoapi.Service{}
	err = serviceEntity.GetAspect(serviceAspect)
	if err != nil {
		return nil, nil, nil, err
	}

	mastershipState := serviceAspect.GetMastershipstate()

	controllerID := controllerutils.GetControllerID()
	controllerEntity, err := c.topoStore.Get(ctx, controllerID)
	if err != nil {
		return nil, nil, nil, err
	}

	controllerInfo := &topoapi.ControllerInfo{}
	err = controllerEntity.GetAspect(controllerInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	if mastershipState.ConnectionID == "" {
		return nil, nil, nil, errors.NewNotFound("Not found master connection  for target %s", targetID)
	}
	relation, err := c.topoStore.Get(ctx, topoapi.ID(mastershipState.ConnectionID))
	if err != nil {
		return nil, nil, nil, err
	}
	if relation.GetRelation().SrcEntityID != controllerID {
		return nil, nil, nil, errors.NewNotFound("Not found master connection  for target %s", targetID)
	}

	p4rtServerInfo := &topoapi.P4RTServerInfo{}
	err = targetEntity.GetAspect(p4rtServerInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	conn, found := c.conns.Get(ctx, southbound.ConnID(relation.ID))
	if !found {
		return nil, nil, nil, errors.NewNotFound("connection not found for target", targetID)
	}

	return &adminClient{
		targetID: targetEntity.ID,
		conn:     conn,
		deviceID: p4rtServerInfo.DeviceID,
		role:     env.GetServiceName(),
		electionID: &p4api.Uint128{
			Low:  mastershipState.Term,
			High: 0,
		},
	}, targetEntity, serviceEntity, nil
}

func (c *Controller) start() error {
	opts, err := certs.HandleCertPaths(c.Config.CAPath, c.Config.KeyPath, c.Config.CertPath, true)
	if err != nil {
		return err
	}

	conns := southbound.NewConnManager()
	c.conns = conns

	topoStore, err := topo.NewStore(c.Config.TopoAddress, opts...)

	if err != nil {
		return err
	}
	c.topoStore = topoStore
	// Starts node controller
	err = c.startNodeController(topoStore)
	if err != nil {
		return err
	}
	// Starts connection controller
	err = c.startConnController(topoStore, conns)
	if err != nil {
		return err
	}

	// Starts target controller
	err = c.startTargetController(topoStore, conns)
	if err != nil {
		return err
	}
	// Starts mastership controller

	err = c.startServiceController(topoStore)
	if err != nil {
		return err
	}
	err = c.startMastershipController(topoStore, conns)
	if err != nil {
		return err
	}
	log.Info("P4RT controller is running")
	return nil
}

// startNodeController starts node controller
func (c *Controller) startNodeController(topo topo.Store) error {
	nodeController := node.NewController(topo)
	return nodeController.Start()
}

// startConnController starts connection controller
func (c *Controller) startConnController(topo topo.Store, conns southbound.ConnManager) error {
	connController := connection.NewController(topo, conns)
	return connController.Start()
}

// startTargetController starts target controller
func (c *Controller) startTargetController(topo topo.Store, conns southbound.ConnManager) error {
	targetController := target.NewController(topo, conns)
	return targetController.Start()
}

// startMastershipController starts mastership controller
func (c *Controller) startMastershipController(topo topo.Store, conns southbound.ConnManager) error {
	mastershipController := mastership.NewController(topo, conns)
	return mastershipController.Start()
}

func (c *Controller) startServiceController(topo topo.Store) error {
	serviceController := service.NewController(topo)
	return serviceController.Start()
}
