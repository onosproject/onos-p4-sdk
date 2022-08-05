// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-p4-sdk/pkg/southbound"
	"github.com/onosproject/onos-p4-sdk/pkg/types"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
)

// P4TargetClient an interface to interact with the P4 programmable device via P4Runtime API
type P4TargetClient interface {
	ID() topoapi.ID
	Capabilities(ctx context.Context, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error)
	Read(ctx context.Context, entities []*p4api.Entity, opts ...grpc.CallOption) ([]*p4api.Entity, error)
	Write(ctx context.Context, updates []*p4api.Update, atomicity p4api.WriteRequest_Atomicity, opts ...grpc.CallOption) (*p4api.WriteResponse, error)
	StreamChannel() p4api.P4Runtime_StreamChannelClient
	SetForwardingPipelineConfig(ctx context.Context, spec *types.PipelineConfigSpec, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error)
	GetForwardingPipelineConfig(ctx context.Context, responseType p4api.GetForwardingPipelineConfigRequest_ResponseType, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error)
}

type p4TargetClient struct {
	targetID   topoapi.ID
	conn       southbound.Conn
	deviceID   uint64
	role       string
	electionID *p4api.Uint128
}

// SetForwardingPipelineConfig sets forwarding pipeline config
func (p *p4TargetClient) SetForwardingPipelineConfig(ctx context.Context, spec *types.PipelineConfigSpec, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error) {
	request := &p4api.SetForwardingPipelineConfigRequest{
		Config: &p4api.ForwardingPipelineConfig{
			P4DeviceConfig: spec.P4DeviceConfig,
			P4Info:         spec.P4Info,
		},
		DeviceId:   p.deviceID,
		Role:       p.role,
		ElectionId: p.electionID,
		Action:     spec.Action,
	}

	return p.conn.SetForwardingPipelineConfig(ctx, request, opts...)
}

// GetForwardingPipelineConfig gets forwarding pipeline config
func (p *p4TargetClient) GetForwardingPipelineConfig(ctx context.Context, responseType p4api.GetForwardingPipelineConfigRequest_ResponseType, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error) {
	request := &p4api.GetForwardingPipelineConfigRequest{
		DeviceId:     p.deviceID,
		ResponseType: responseType,
	}

	return p.conn.GetForwardingPipelineConfig(ctx, request, opts...)
}

// StreamChannel returns a stream channel client
func (p *p4TargetClient) StreamChannel() p4api.P4Runtime_StreamChannelClient {
	return p.conn.StreamChannel()
}

// Write  Update one or more P4 entities on the target.
func (p *p4TargetClient) Write(ctx context.Context, updates []*p4api.Update, atomicity p4api.WriteRequest_Atomicity, opts ...grpc.CallOption) (*p4api.WriteResponse, error) {
	request := &p4api.WriteRequest{
		DeviceId:   p.deviceID,
		Role:       p.role,
		ElectionId: p.electionID,
		Updates:    updates,
		Atomicity:  atomicity,
	}
	return p.conn.Write(ctx, request, opts...)
}

// Read one or more P4 entities from the target.
func (p *p4TargetClient) Read(ctx context.Context, entities []*p4api.Entity, opts ...grpc.CallOption) ([]*p4api.Entity, error) {
	request := &p4api.ReadRequest{
		Role:     p.role,
		DeviceId: p.deviceID,
		Entities: entities,
	}
	return p.conn.ReadEntities(ctx, request, opts...)
}

// Capabilities return P4Runtime server capability
func (p *p4TargetClient) Capabilities(ctx context.Context, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error) {
	request := &p4api.CapabilitiesRequest{}
	return p.conn.Capabilities(ctx, request)
}

// ID gets target ID
func (p *p4TargetClient) ID() topoapi.ID {
	return p.conn.TargetID()
}

var _ P4TargetClient = &p4TargetClient{}
