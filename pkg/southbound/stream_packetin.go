// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import p4api "github.com/p4lang/p4runtime/go/p4/v1"

type bufferedChannelPacketIn struct {
	packetInBuffer []*p4api.PacketIn
}

func newPacketInStream(out chan *p4api.PacketIn) chan<- *p4api.PacketIn {
	b := bufferedChannelPacketIn{}
	in := make(chan *p4api.PacketIn)
	go func() {
		for len(b.packetInBuffer) > 0 || in != nil {
			select {
			case v, ok := <-in:
				if !ok {
					in = nil
				} else {
					b.packetInBuffer = append(b.packetInBuffer, v)
				}
			case b.out(out) <- b.currentVal():
				b.packetInBuffer = b.packetInBuffer[1:]
			}
		}
		close(out)
	}()
	return in

}

func (b *bufferedChannelPacketIn) currentVal() *p4api.PacketIn {
	if len(b.packetInBuffer) == 0 {
		return nil
	}
	return b.packetInBuffer[0]
}

func (b *bufferedChannelPacketIn) out(out chan *p4api.PacketIn) chan *p4api.PacketIn {
	if len(b.packetInBuffer) == 0 {
		return nil
	}
	return out
}
