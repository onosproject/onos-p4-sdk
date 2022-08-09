// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import (
	"context"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"io"
)

// ReadClient :
type ReadClient interface {
	ReadEntities(ctx context.Context, request *p4api.ReadRequest, ch chan *p4api.Entity, opts ...grpc.CallOption) error
}

type readClient struct {
	p4runtimeClient p4api.P4RuntimeClient
}

// ReadEntities :
func (r readClient) ReadEntities(ctx context.Context, request *p4api.ReadRequest, outputCh chan *p4api.Entity, opts ...grpc.CallOption) error {
	stream, err := r.p4runtimeClient.Read(ctx, request, opts...)
	if err != nil {
		return err
	}
	go func() {
		inputChan := newReadEntitiesStream(outputCh)
		defer close(inputChan)
		for {
			rep, err := stream.Recv()
			if err == io.EOF || err == context.Canceled {
				break
			}
			if err != nil {
				// TODO should we return the error and break the loop?
				log.Warn(err)
				continue
			}
			for _, entity := range rep.Entities {
				inputChan <- entity
			}

		}
	}()
	return nil

}

var _ ReadClient = &readClient{}
