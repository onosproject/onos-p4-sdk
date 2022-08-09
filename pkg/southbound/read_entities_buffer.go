// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import p4api "github.com/p4lang/p4runtime/go/p4/v1"

type bufferedChannelEntities struct {
	entitiesBuffer []*p4api.Entity
}

func newReadEntitiesStream(out chan *p4api.Entity) chan<- *p4api.Entity {
	b := bufferedChannelEntities{}
	in := make(chan *p4api.Entity)
	go func() {
		for len(b.entitiesBuffer) > 0 || in != nil {
			select {
			case v, ok := <-in:
				if !ok {
					in = nil
				} else {
					b.entitiesBuffer = append(b.entitiesBuffer, v)
				}
			case b.out(out) <- b.currentVal():
				b.entitiesBuffer = b.entitiesBuffer[1:]
			}
		}
		close(out)
	}()
	return in

}

func (b *bufferedChannelEntities) currentVal() *p4api.Entity {
	if len(b.entitiesBuffer) == 0 {
		return nil
	}
	return b.entitiesBuffer[0]
}

func (b *bufferedChannelEntities) out(out chan *p4api.Entity) chan *p4api.Entity {
	if len(b.entitiesBuffer) == 0 {
		return nil
	}
	return out
}
