// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	p4configapi "github.com/p4lang/p4runtime/go/p4/config/v1"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
)

// PipelineConfigSpec pipeline config info
type PipelineConfigSpec struct {
	P4Info         *p4configapi.P4Info
	P4DeviceConfig []byte
	Action         p4api.SetForwardingPipelineConfigRequest_Action
}
