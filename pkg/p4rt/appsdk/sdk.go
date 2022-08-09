// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package appsdk

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-p4-sdk/pkg/p4rt/admin"
	"github.com/onosproject/onos-p4-sdk/pkg/southbound"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
)

// TargetClient an interface to interact with the P4 programmable device via P4Runtime API
type TargetClient interface {
	ID() topoapi.ID
	Capabilities(ctx context.Context, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error)
	Read(ctx context.Context, entities []*p4api.Entity, ch chan *p4api.Entity, opts ...grpc.CallOption) error
	Write(ctx context.Context, updates []*p4api.Update, atomicity p4api.WriteRequest_Atomicity, opts ...grpc.CallOption) (*p4api.WriteResponse, error)
	GetForwardingPipelineConfig(ctx context.Context, responseType p4api.GetForwardingPipelineConfigRequest_ResponseType, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error)
	PacketIn(outputCh chan *p4api.PacketIn) error
	PacketOut(packetOut *p4api.PacketOut) error
}

type targetClient struct {
	targetID   topoapi.ID
	conn       southbound.Conn
	deviceID   uint64
	role       string
	electionID *p4api.Uint128
}

func (p *targetClient) PacketIn(outputCh chan *p4api.PacketIn) error {
	log.Debugw("Receiving packet in")
	return p.conn.PacketIn(outputCh)
}

func (p *targetClient) PacketOut(packetOut *p4api.PacketOut) error {
	log.Debugw("Sending packet out", "packet out", packetOut)
	return p.conn.PacketOut(packetOut)
}

// SetForwardingPipelineConfig sets forwarding pipeline config
func (p *targetClient) SetForwardingPipelineConfig(ctx context.Context, spec *admin.PipelineConfigSpec, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error) {
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

	log.Debugw("Setting forwarding pipeline device config", "request", request, "pipeline config spec", spec)
	return p.conn.SetForwardingPipelineConfig(ctx, request, opts...)
}

// GetForwardingPipelineConfig gets forwarding pipeline config
func (p *targetClient) GetForwardingPipelineConfig(ctx context.Context, responseType p4api.GetForwardingPipelineConfigRequest_ResponseType, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error) {
	request := &p4api.GetForwardingPipelineConfigRequest{
		DeviceId:     p.deviceID,
		ResponseType: responseType,
	}

	log.Debugw("Getting forwarding pipeline device config", "responseType", responseType)
	return p.conn.GetForwardingPipelineConfig(ctx, request, opts...)
}

// Write  Update one or more P4 entities on the target.
func (p *targetClient) Write(ctx context.Context, updates []*p4api.Update, atomicity p4api.WriteRequest_Atomicity, opts ...grpc.CallOption) (*p4api.WriteResponse, error) {
	request := &p4api.WriteRequest{
		DeviceId:   p.deviceID,
		Role:       p.role,
		ElectionId: p.electionID,
		Updates:    updates,
		Atomicity:  atomicity,
	}
	log.Debugw("Writing updates", "request", request)
	return p.conn.Write(ctx, request, opts...)
}

// Read one or more P4 entities from the target.
func (p *targetClient) Read(ctx context.Context, entities []*p4api.Entity, ch chan *p4api.Entity, opts ...grpc.CallOption) error {
	request := &p4api.ReadRequest{
		Role:     p.role,
		DeviceId: p.deviceID,
		Entities: entities,
	}
	log.Debugw("Reading entities", "entities", entities)
	return p.conn.ReadEntities(ctx, request, ch, opts...)
}

// Capabilities return P4Runtime server capability
func (p *targetClient) Capabilities(ctx context.Context, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error) {
	request := &p4api.CapabilitiesRequest{}
	return p.conn.Capabilities(ctx, request)
}

// ID gets target ID
func (p *targetClient) ID() topoapi.ID {
	return p.conn.TargetID()
}

var _ TargetClient = &targetClient{}
