// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import (
	"context"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
)

// PipelineConfigClient :
type PipelineConfigClient interface {
	SetForwardingPipelineConfig(ctx context.Context, request *p4api.SetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error)
	GetForwardingPipelineConfig(ctx context.Context, request *p4api.GetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error)
}

type pipelineConfigClient struct {
	p4runtimeClient p4api.P4RuntimeClient
}

// SetForwardingPipelineConfig :
func (p *pipelineConfigClient) SetForwardingPipelineConfig(ctx context.Context, request *p4api.SetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error) {
	return p.p4runtimeClient.SetForwardingPipelineConfig(ctx, request, opts...)
}

// GetForwardingPipelineConfig :
func (p *pipelineConfigClient) GetForwardingPipelineConfig(ctx context.Context, request *p4api.GetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error) {
	return p.p4runtimeClient.GetForwardingPipelineConfig(ctx, request, opts...)
}

var _ PipelineConfigClient = &pipelineConfigClient{}
