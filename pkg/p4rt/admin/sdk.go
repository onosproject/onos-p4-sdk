// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-p4-sdk/pkg/southbound"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
)

// Client an admin interface for setting/getting pipeline configuration
type Client interface {
	SetForwardingPipelineConfig(ctx context.Context, spec *PipelineConfigSpec, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error)
	GetForwardingPipelineConfig(ctx context.Context, responseType p4api.GetForwardingPipelineConfigRequest_ResponseType, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error)
}

type adminClient struct {
	targetID   topoapi.ID
	conn       southbound.Conn
	deviceID   uint64
	role       string
	electionID *p4api.Uint128
}

func (a *adminClient) SetForwardingPipelineConfig(ctx context.Context, spec *PipelineConfigSpec, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error) {
	request := &p4api.SetForwardingPipelineConfigRequest{
		Config: &p4api.ForwardingPipelineConfig{
			P4DeviceConfig: spec.P4DeviceConfig,
			P4Info:         spec.P4Info,
		},
		DeviceId:   a.deviceID,
		Role:       a.role,
		ElectionId: a.electionID,
		Action:     spec.Action,
	}

	log.Debugw("Setting forwarding pipeline device config", "request", request, "pipeline config spec", spec)
	return a.conn.SetForwardingPipelineConfig(ctx, request, opts...)
}

func (a *adminClient) GetForwardingPipelineConfig(ctx context.Context, responseType p4api.GetForwardingPipelineConfigRequest_ResponseType, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error) {
	request := &p4api.GetForwardingPipelineConfigRequest{
		DeviceId:     a.deviceID,
		ResponseType: responseType,
	}
	log.Debugw("Getting forwarding pipeline device config", "responseType", responseType)
	return a.conn.GetForwardingPipelineConfig(ctx, request, opts...)
}

var _ Client = &adminClient{}
