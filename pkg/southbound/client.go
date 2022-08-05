// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"io"
)

// Client P4runtime client interface
type Client interface {
	io.Closer
	WriteClient
	ReadClient
	StreamClient
	PipelineConfigClient
	Capabilities(ctx context.Context, request *p4api.CapabilitiesRequest, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error)
}

type client struct {
	grpcClient           *grpc.ClientConn
	p4runtimeClient      p4api.P4RuntimeClient
	writeClient          *writeClient
	readClient           *readClient
	pipelineConfigClient *pipelineConfigClient
	streamClient         *streamClient
}

func (c *client) StreamChannel() p4api.P4Runtime_StreamChannelClient {
	return c.streamClient.StreamChannel()
}

func (c *client) RecvArbitrationResponse() (*p4api.StreamMessageResponse_Arbitration, error) {
	response, err := c.streamClient.RecvArbitrationResponse()
	return response, errors.FromGRPC(err)
}

func (c *client) SendArbitrationRequest(deviceID uint64, electionID uint64, role string) error {
	err := c.streamClient.SendArbitrationRequest(deviceID, electionID, role)
	return errors.FromGRPC(err)
}

func (c *client) ReadEntities(ctx context.Context, request *p4api.ReadRequest, opts ...grpc.CallOption) ([]*p4api.Entity, error) {
	log.Debugw("Received read entities request", "request", request)
	entities, err := c.readClient.ReadEntities(ctx, request, opts...)
	if err != nil {
		return nil, errors.FromGRPC(err)
	}
	return entities, nil

}

// Write Updates one or more P4 entities on the target.
func (c *client) Write(ctx context.Context, request *p4api.WriteRequest, opts ...grpc.CallOption) (*p4api.WriteResponse, error) {
	log.Debugw("Received Write request", "request", request)
	writeResponse, err := c.writeClient.Write(ctx, request, opts...)
	return writeResponse, errors.FromGRPC(err)
}

// SetForwardingPipelineConfig  sets the P4 forwarding-pipeline config.
func (c *client) SetForwardingPipelineConfig(ctx context.Context, request *p4api.SetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error) {
	log.Debugw("Received SetForwardingPipelineConfig request", "request", request)
	setForwardingPipelineConfigResponse, err := c.pipelineConfigClient.SetForwardingPipelineConfig(ctx, request, opts...)
	return setForwardingPipelineConfigResponse, errors.FromGRPC(err)
}

// GetForwardingPipelineConfig  gets the current P4 forwarding-pipeline config.
func (c *client) GetForwardingPipelineConfig(ctx context.Context, request *p4api.GetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error) {
	log.Debugw("Received GetForwardingPipelineConfig request", "request", request)
	getForwardingPipelineConfigResponse, err := c.pipelineConfigClient.GetForwardingPipelineConfig(ctx, request, opts...)
	return getForwardingPipelineConfigResponse, errors.FromGRPC(err)
}

// Capabilities discovers the capabilities of the P4Runtime server implementation.
func (c *client) Capabilities(ctx context.Context, request *p4api.CapabilitiesRequest, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error) {
	log.Debugw("Received Capabilities request", "request", request)
	capabilitiesResponse, err := c.p4runtimeClient.Capabilities(ctx, request, opts...)
	return capabilitiesResponse, errors.FromGRPC(err)
}

func (c *client) Close() error {
	return c.grpcClient.Close()
}

var _ Client = &client{}
