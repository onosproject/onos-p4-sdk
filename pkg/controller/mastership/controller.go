// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mastership

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-p4-sdk/pkg/controller/utils"
	"github.com/onosproject/onos-p4-sdk/pkg/southbound"
	"github.com/onosproject/onos-p4-sdk/pkg/store/topo"
	"google.golang.org/genproto/googleapis/rpc/code"
	"io"

	"time"
)

var log = logging.GetLogger()

const defaultTimeout = 30 * time.Second

// NewController returns a new mastership controller
func NewController(topo topo.Store, conns southbound.ConnManager) *controller.Controller {
	c := controller.NewController("mastership")
	c.Watch(&TopoWatcher{
		topo: topo,
	})

	c.Reconcile(&Reconciler{
		topo:  topo,
		conns: conns,
	})
	return c
}

// Reconciler is mastership reconciler
type Reconciler struct {
	topo  topo.Store
	conns southbound.ConnManager
}

// Reconcile reconciles the mastership state for a gnmi target
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	serviceID := id.Value.(topoapi.ID)
	log.Infow("Reconciling mastership election for the P4RT target", "service ID", serviceID)
	serviceEntity, err := r.topo.Get(ctx, serviceID)
	if err != nil {
		if errors.IsNotFound(err) {
			return controller.Result{}, nil
		}
		log.Warnw("Failed to reconcile mastership election for the P4RT target", "service ID", serviceID, "error", err)
		return controller.Result{}, err
	}
	serviceAspect := &topoapi.Service{}
	err = serviceEntity.GetAspect(serviceAspect)
	if err != nil {
		log.Warnw("Failed to reconcile mastership election for the P4RT target", "service ID", serviceID, "error", err)
		return controller.Result{}, err
	}

	targetID := topoapi.ID(serviceAspect.TargetID)

	targetEntity, err := r.topo.Get(ctx, targetID)
	if err != nil {
		if errors.IsNotFound(err) {
			return controller.Result{}, nil
		}
		log.Warnw("Failed to reconcile mastership election for the P4RT target", "applicationID", serviceID, "targetID", targetEntity.ID, "error", err)
		return controller.Result{}, err
	}

	p4targetInfo := &topoapi.P4RTServerInfo{}
	err = targetEntity.GetAspect(p4targetInfo)
	if err != nil {
		log.Warnw("Failed to reconcile mastership election for the P4RT target", "applicationID", serviceID, "targetID", targetEntity.ID, "error", err)
		return controller.Result{}, err
	}

	controllerID := utils.GetControllerID()
	controllerEntity, err := r.topo.Get(ctx, controllerID)
	if err != nil {
		if errors.IsNotFound(err) {
			return controller.Result{}, nil
		}
		log.Warnw("Failed to reconcile mastership election for the P4RT target", "targetID", targetEntity.ID, "error", err)
		return controller.Result{}, err
	}

	controllerInfo := &topoapi.ControllerInfo{}
	err = controllerEntity.GetAspect(controllerInfo)
	if err != nil {
		log.Warnw("Failed to reconcile mastership election for the P4RT target", "serviceID", serviceEntity, "error", err)
		return controller.Result{}, err
	}

	// List the objects in the topo store
	objects, err := r.topo.List(ctx, &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: topoapi.ConnectionKind,
				},
			},
		},
	})

	if err != nil {
		log.Warnw("Updating MastershipState for service failed", "targetI", serviceEntity.ID, "error", err)
		return controller.Result{}, err
	}
	targetRelations := make(map[topoapi.ID]topoapi.Object)
	for _, object := range objects {
		if object.GetRelation().TgtEntityID == serviceID {
			targetRelations[object.ID] = object
		}
	}

	mastership := serviceAspect.Mastershipstate
	if _, ok := targetRelations[topoapi.ID(mastership.ConnectionID)]; !ok {
		if len(targetRelations) == 0 {
			if mastership.ConnectionID == "" {
				return controller.Result{}, nil
			}
			log.Infow("Master in term resigned for the P4RT target", "serviceID", serviceEntity.ID, "mastership term", mastership.Term)
			mastership.ConnectionID = ""
			serviceAspect.Mastershipstate = mastership
			err = serviceEntity.SetAspect(serviceAspect)
			if err != nil {
				return controller.Result{}, err
			}
			// Update mastership state in the P4RT target entity
			err = r.topo.Update(ctx, serviceEntity)
			if err != nil {
				if !errors.IsNotFound(err) && !errors.IsConflict(err) {
					log.Warnw("Updating MastershipState for P4 target failed", "serviceID", serviceEntity.ID, "error", err)
					return controller.Result{}, err
				}
				log.Warn(err)
				return controller.Result{}, nil
			}
			return controller.Result{}, nil
		}
		conn, err := r.conns.GetByTarget(ctx, targetID)
		if err != nil {
			if errors.IsNotFound(err) {
				return controller.Result{}, nil
			}
			log.Warnw("Failed to reconcile mastership election for the P4RT target", "targetID", targetEntity.ID, "error", err)
			return controller.Result{}, err
		}

		electionID := mastership.Term + 1
		log.Infow("Sending MasterArbitrationUpdate message", "target ID", targetEntity.ID, "election ID", electionID)
		err = conn.SendArbitrationRequest(p4targetInfo.DeviceID, electionID, controllerInfo.Role.Name)
		if err != nil {
			if errors.IsNotFound(err) || errors.IsInvalid(err) {
				log.Warnw("Failed to reconcile mastership election for the P4RT target", "targetID", targetEntity.ID, "error", err)
				return controller.Result{}, nil
			}
			log.Warnw("Failed to reconcile mastership election for the P4RT target", "targetID", targetEntity.ID, "error", err.Error())
			return controller.Result{}, err
		}
		response, err := conn.RecvArbitrationResponse()
		if err != nil {
			log.Warnw("Failed to reconcile mastership election for the P4RT target", "targetID", targetEntity.ID, "error", err)
			// If the election_id is set and is already used by another controller
			// for the same (device_id, role), the P4Runtime server shall terminate the stream by returning an INVALID_ARGUMENT error.
			if errors.IsInvalid(err) {
				log.Warnw("Invalid argument, failed to reconcile mastership election for the P4RT target", "error", err)
				return controller.Result{}, err
			}
			if err == io.EOF {
				log.Warnw("End of file")
				return controller.Result{}, nil
			}

		}

		/*status is set differently based on whether the notification is sent to the primary or a backup controller:
		If there is a primary:
		   * For the primary, status is OK (with status.code set to google.rpc.OK).
		   * For all backup controllers, status is set to non-OK (with status.code set to google.rpc.ALREADY_EXISTS).
		Otherwise, if there is no primary currently, for all backup controllers, status is set to non-OK (with status.code set to google.rpc.NOT_FOUND).*/
		statusCode := response.Arbitration.Status.Code
		if statusCode == int32(code.Code_OK) {
			for _, targetRelation := range targetRelations {
				if targetRelation.GetRelation().SrcEntityID == utils.GetControllerID() {
					responseElectionID := response.Arbitration.ElectionId.Low
					log.Infow("Current node is selected as master, updating mastership status", "targetID", targetEntity.ID, "election ID", responseElectionID)
					mastership.ConnectionID = string(targetRelation.ID)
					mastership.Term = responseElectionID
					mastership.Role = response.Arbitration.Role.Name
					serviceAspect.Mastershipstate = mastership
					err = serviceEntity.SetAspect(serviceAspect)
					if err != nil {
						log.Warnw("Updating MastershipState for P4 target failed", "targetID", targetEntity.ID, "error", err)
						return controller.Result{}, err
					}

					// Update mastership state in the P4RT target entity
					err = r.topo.Update(ctx, serviceEntity)
					if err != nil {
						if !errors.IsNotFound(err) && !errors.IsConflict(err) {
							log.Warnw("Updating MastershipState for P4 target failed", "targetID", targetEntity.ID, "error", err)
							return controller.Result{}, err
						}
						log.Warn(err)
						return controller.Result{}, nil
					}
					return controller.Result{}, nil
				}
			}

		} else if statusCode == int32(code.Code_ALREADY_EXISTS) {
			log.Infow("Master is already selected for target", "targetID", targetEntity.ID)
			return controller.Result{}, nil
		} else if statusCode == int32(code.Code_NOT_FOUND) {
			log.Infow("No master found for target, retrying master arbitration update request", "targetID", targetEntity.ID)
			return controller.Result{
				Requeue: id,
			}, nil
		}
		return controller.Result{}, nil

	}
	return controller.Result{}, nil
}
