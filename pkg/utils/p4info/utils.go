// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4info

import (
	"github.com/onosproject/onos-lib-go/pkg/errors"
	p4configapi "github.com/p4lang/p4runtime/go/p4/config/v1"
)

// Helper an interface to extract required information from P4Info struct
type Helper interface {
	Table(tableName string) (*p4configapi.Table, error)
	TableID(tableName string) (uint32, error)
	Action(actionName string) (*p4configapi.Action, error)
	ActionID(actionName string) (uint32, error)
	MatchFieldID(tableName string, fieldName string) (uint32, error)
	ActionProfile(actionProfileName string) (*p4configapi.ActionProfile, error)
	ActionProfileID(actionProfileName string) (uint32, error)
	Digest(digestName string) (*p4configapi.Digest, error)
	DigestID(digestName string) (uint32, error)
	Counter(counterName string) (*p4configapi.Counter, error)
	CounterID(counterName string) (uint32, error)
	DirectCounter(counterName string) (*p4configapi.DirectCounter, error)
	DirectCounterID(counterName string) (uint32, error)
	DirectMeter(meterName string) (*p4configapi.DirectMeter, error)
	DirectMeterID(meterName string) (uint32, error)
	RegisterID(registerName string) (uint32, error)
	Register(registerName string) (*p4configapi.Register, error)
	ValueSetID(valueSetName string) (uint32, error)
	ValueSet(valueSetName string) (*p4configapi.ValueSet, error)
	ControllerPacketMetadata(ControllerPacketMetadataName string) (*p4configapi.ControllerPacketMetadata, error)
	ControllerPacketMetadataID(ControllerPacketMetadataName string) (uint32, error)
}

type p4Info struct {
	p4Info *p4configapi.P4Info
}

func (p *p4Info) ControllerPacketMetadata(controllerPacketMetadataName string) (*p4configapi.ControllerPacketMetadata, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, controllerPacketMetadata := range p.p4Info.ControllerPacketMetadata {
		if controllerPacketMetadata.Preamble.Name == controllerPacketMetadataName {
			return controllerPacketMetadata, nil
		}
	}
	return nil, errors.NewNotFound("controller packet metadata  %s not found", controllerPacketMetadataName)
}

func (p *p4Info) ControllerPacketMetadataID(controllerPacketMetadataName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}
	controllerPacketMetadata, err := p.ControllerPacketMetadata(controllerPacketMetadataName)
	if err != nil {
		return 0, err
	}
	return controllerPacketMetadata.Preamble.Id, nil
}

func (p *p4Info) ValueSetID(valueSetName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}
	valueSet, err := p.ValueSet(valueSetName)
	if err != nil {
		return 0, err
	}
	return valueSet.Preamble.Id, nil
}

func (p *p4Info) ValueSet(valueSetName string) (*p4configapi.ValueSet, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, valueSet := range p.p4Info.ValueSets {
		if valueSet.Preamble.Name == valueSetName {
			return valueSet, nil
		}
	}
	return nil, errors.NewNotFound("value set  %s not found", valueSetName)
}

func (p *p4Info) RegisterID(registerName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}
	register, err := p.Register(registerName)
	if err != nil {
		return 0, err
	}
	return register.Preamble.Id, nil
}

func (p *p4Info) Register(registerName string) (*p4configapi.Register, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, register := range p.p4Info.Registers {
		if register.Preamble.Name == registerName {
			return register, nil
		}
	}
	return nil, errors.NewNotFound("register %s not found", registerName)
}

func (p *p4Info) DirectMeter(directMeterName string) (*p4configapi.DirectMeter, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, directMeter := range p.p4Info.DirectMeters {
		if directMeter.Preamble.Name == directMeterName {
			return directMeter, nil
		}
	}
	return nil, errors.NewNotFound("direct meter %s not found", directMeterName)
}

func (p *p4Info) DirectMeterID(directMeterName string) (uint32, error) {
	directMeter, err := p.DirectMeter(directMeterName)
	if err != nil {
		return 0, err
	}
	return directMeter.Preamble.Id, nil
}

func (p *p4Info) DirectCounter(directCounterName string) (*p4configapi.DirectCounter, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, directCounter := range p.p4Info.DirectCounters {
		if directCounter.Preamble.Name == directCounterName {
			return directCounter, nil
		}
	}
	return nil, errors.NewNotFound("direct counter %s not found", directCounterName)
}

func (p *p4Info) DirectCounterID(directCounterName string) (uint32, error) {
	directCounter, err := p.DirectCounter(directCounterName)
	if err != nil {
		return 0, err
	}
	return directCounter.Preamble.Id, nil
}

func (p *p4Info) CounterID(counterName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}

	counter, err := p.Counter(counterName)
	if err != nil {
		return 0, err
	}
	return counter.Preamble.Id, nil
}

func (p *p4Info) Counter(counterName string) (*p4configapi.Counter, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, counter := range p.p4Info.Counters {
		if counter.Preamble.Name == counterName {
			return counter, nil
		}
	}
	return nil, errors.NewNotFound("counter %s not found", counterName)
}

func (p *p4Info) Digest(digestName string) (*p4configapi.Digest, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, digest := range p.p4Info.Digests {
		if digest.Preamble.Name == digestName {
			return digest, nil
		}
	}

	return nil, errors.NewNotFound("digest %s not found", digestName)
}

func (p *p4Info) DigestID(digestName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}
	for _, digest := range p.p4Info.Digests {
		if digest.Preamble.Name == digestName {
			return digest.Preamble.Id, nil
		}
	}
	return 0, errors.NewNotFound("digest %s not found", digestName)
}

func (p *p4Info) ActionProfile(actionProfileName string) (*p4configapi.ActionProfile, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, actionProfile := range p.p4Info.ActionProfiles {
		if actionProfile.Preamble.Name == actionProfileName {
			return actionProfile, nil
		}
	}

	return nil, errors.NewNotFound("action profile %s not found", actionProfileName)

}

func (p *p4Info) ActionProfileID(actionProfileName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}
	actionProfile, err := p.ActionProfile(actionProfileName)
	if err != nil {
		return 0, err
	}

	return actionProfile.Preamble.Id, nil
}

func (p *p4Info) MatchFieldID(tableName string, fieldName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}
	table, err := p.Table(tableName)
	if err != nil {
		return 0, err
	}
	for _, mf := range table.MatchFields {
		if mf.Name == fieldName {
			return mf.Id, nil
		}
	}
	return 0, errors.NewNotFound("field name %s is not found in table %s", fieldName, tableName)
}

func (p *p4Info) Action(actionName string) (*p4configapi.Action, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, action := range p.p4Info.Actions {
		if action.Preamble.Name == actionName {
			return action, nil
		}
	}
	return nil, errors.NewNotFound("action %s not found", actionName)
}

func (p *p4Info) ActionID(actionName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}

	action, err := p.Action(actionName)
	if err != nil {
		return 0, err
	}
	return action.Preamble.Id, nil
}

func (p *p4Info) Table(tableName string) (*p4configapi.Table, error) {
	if p.p4Info == nil {
		return nil, errors.NewInvalid("p4 info is not initialized")
	}
	for _, table := range p.p4Info.Tables {
		if table.Preamble.Name == tableName {
			return table, nil
		}
	}
	return nil, errors.NewNotFound("table %s not found", tableName)
}

func (p *p4Info) TableID(tableName string) (uint32, error) {
	if p.p4Info == nil {
		return 0, errors.NewInvalid("p4 info is not initialized")
	}

	table, err := p.Table(tableName)
	if err != nil {
		return 0, err
	}
	return table.Preamble.Id, nil

}

// NewHelper creates an instance P4Info helper interface
func NewHelper(p4info *p4configapi.P4Info) Helper {
	return &p4Info{
		p4Info: p4info,
	}
}

var _ Helper = &p4Info{}
