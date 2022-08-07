// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"io"
)

// StreamClient p4runtime master stream client
type StreamClient interface {
	SendArbitrationRequest(deviceID uint64, electionID uint64, role string) error
	RecvArbitrationResponse() (*p4api.StreamMessageResponse_Arbitration, error)
	PacketIn(outputCh chan *p4api.PacketIn) error
	PacketOut(packetOut *p4api.PacketOut) error
}

type streamClient struct {
	p4runtimeClient p4api.P4RuntimeClient
	streamChannel   p4api.P4Runtime_StreamChannelClient
}

func (s *streamClient) PacketOut(packetOut *p4api.PacketOut) error {
	request := &p4api.StreamMessageRequest{
		Update: &p4api.StreamMessageRequest_Packet{
			Packet: packetOut,
		},
	}
	log.Info("Sending packet out")
	err := s.streamChannel.Send(request)
	if err != nil {
		log.Warnw("cannot send packet out message", "error", err)
		return err
	}
	return nil

}

func (s *streamClient) PacketIn(outputChan chan *p4api.PacketIn) error {
	go func() {
		inputChan := newPacketInStream(outputChan)
		defer close(inputChan)
		for {
			in, err := s.streamChannel.Recv()
			if err == io.EOF || err == context.Canceled {
				log.Warnw("Closing stream channel", "error", err)
				break
			}
			if err != nil {
				log.Warnw("Error in receiving stream response message", "error", err)
				continue
			}
			switch v := in.Update.(type) {
			case *p4api.StreamMessageResponse_Packet:
				log.Debugw("Sending PACKET_IN response message", "packetIn", v)
				inputChan <- v.Packet

			}
		}
	}()
	return nil
}

func (s *streamClient) RecvArbitrationResponse() (*p4api.StreamMessageResponse_Arbitration, error) {
	in, err := s.streamChannel.Recv()
	if err != nil {
		return nil, err
	}

	switch v := in.Update.(type) {
	case *p4api.StreamMessageResponse_Arbitration:
		log.Infow("Received arbitration response", "response", v)
		if err != nil {
			return nil, err
		}
		return v, nil

	}
	return nil, errors.NewNotSupported("not an arbitration response message")

}

func (s *streamClient) SendArbitrationRequest(deviceID uint64, electionID uint64, role string) error {
	request := &p4api.StreamMessageRequest{
		Update: &p4api.StreamMessageRequest_Arbitration{Arbitration: &p4api.MasterArbitrationUpdate{
			DeviceId: deviceID,
			ElectionId: &p4api.Uint128{
				Low:  electionID,
				High: 0,
			},
			Role: &p4api.Role{
				Name: role,
			},
		}},
	}
	log.Infow("Sending master arbitration request", "request", request)
	err := s.streamChannel.Send(request)
	return err
}

var _ StreamClient = &streamClient{}
