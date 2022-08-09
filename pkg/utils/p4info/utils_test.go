// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4info

import (
	_ "embed"
	"fmt"
	p4configapi "github.com/p4lang/p4runtime/go/p4/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/prototext"
	"os"
	"testing"
)

//go:embed testdata/p4info.txt
var p4InfoBytes []byte

var p4InfoTest p4configapi.P4Info

const (
	aclPreIngressTableName               = "ingress.acl_pre_ingress.acl_pre_ingress_table"
	aclPreIngressTableID                 = uint32(33554689)
	aclPreIngressDirectCounterName       = "ingress.acl_pre_ingress.acl_pre_ingress_counter"
	aclPreIngressDirectCounterID         = uint32(318767361)
	aclDropAction                        = "acl_drop"
	aclDropActionID                      = uint32(16777481)
	aclIngressMeterName                  = "ingress.acl_ingress.acl_ingress_meter"
	aclIngressMeterID                    = uint32(352321792)
	actionProfileWCMPGroupSelectorName   = "ingress.routing.wcmp_group_selector"
	actionProfileWCMPGroupSelectorID     = uint32(299650760)
	noActionName                         = "NoAction"
	noActionID                           = uint32(21257015)
	packetInControllerPacketMetadataName = "packet_in"
	packetInControllerPacketMetadataID   = uint32(81826293)
)

func TestMain(m *testing.M) {
	err := prototext.Unmarshal(p4InfoBytes, &p4InfoTest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestP4Info_ActionProfileID(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	actionProfileID, err := p4InfoHelper.ActionProfileID(actionProfileWCMPGroupSelectorName)
	assert.NoError(t, err)
	assert.Equal(t, actionProfileWCMPGroupSelectorID, actionProfileID)
}

func TestP4Info_ControllerPacketMetadata(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	packetInControllerPacketMetadata, err := p4InfoHelper.ControllerPacketMetadata(packetInControllerPacketMetadataName)
	assert.NoError(t, err)
	assert.Equal(t, packetInControllerPacketMetadata.Preamble.Id, packetInControllerPacketMetadataID)

}

func TestP4Info_TableID(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	tableID, err := p4InfoHelper.TableID(aclPreIngressTableName)
	assert.NoError(t, err)
	assert.Equal(t, aclPreIngressTableID, tableID)
}

func TestP4Info_ActionID(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	actionID, err := p4InfoHelper.ActionID(noActionName)
	assert.NoError(t, err)
	assert.Equal(t, noActionID, actionID)

	dropActionID, err := p4InfoHelper.ActionID(aclDropAction)
	assert.NoError(t, err)
	assert.Equal(t, aclDropActionID, dropActionID)

}

func TestP4Info_Action(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	noAction, err := p4InfoHelper.Action(noActionName)
	assert.NoError(t, err)
	assert.Equal(t, noAction.Preamble.Id, noActionID)

}

func TestP4Info_MatchFieldID(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	matchFieldID, err := p4InfoHelper.MatchFieldID(aclPreIngressTableName, "src_mac")
	assert.NoError(t, err)
	assert.Equal(t, uint32(4), matchFieldID)
}

func TestP4Info_DirectCounterID(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	directCounterID, err := p4InfoHelper.DirectCounterID(aclPreIngressDirectCounterName)
	assert.NoError(t, err)
	assert.Equal(t, aclPreIngressDirectCounterID, directCounterID)

}

func TestP4Info_Table(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	table, err := p4InfoHelper.Table(aclPreIngressTableName)
	assert.NoError(t, err)
	assert.Equal(t, aclPreIngressTableID, table.Preamble.Id)
}

func TestP4Info_DirectCounter(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	directCounter, err := p4InfoHelper.DirectCounter(aclPreIngressDirectCounterName)
	assert.NoError(t, err)
	assert.Equal(t, aclPreIngressDirectCounterID, directCounter.Preamble.Id)
}

func TestP4Info_DirectMeterID(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	directMeterID, err := p4InfoHelper.DirectMeterID(aclIngressMeterName)
	assert.NoError(t, err)
	assert.Equal(t, aclIngressMeterID, directMeterID)

}

func TestP4Info_DirectMeter(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	directMeter, err := p4InfoHelper.DirectMeter(aclIngressMeterName)
	assert.NoError(t, err)
	assert.Equal(t, aclIngressMeterID, directMeter.Preamble.Id)
}

func TestP4Info_ActionProfile(t *testing.T) {
	p4InfoHelper := NewHelper(&p4InfoTest)
	actionProfile, err := p4InfoHelper.ActionProfile(actionProfileWCMPGroupSelectorName)
	assert.NoError(t, err)
	assert.Equal(t, actionProfile.Preamble.Id, actionProfileWCMPGroupSelectorID)

}
