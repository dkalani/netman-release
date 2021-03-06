// This file was generated by counterfeiter
package fakes

import (
	"netman-agent/rules"
	"sync"

	"github.com/pivotal-golang/lager"
)

type Rule struct {
	EnforceStub        func(string, rules.IPTables, lager.Logger) error
	enforceMutex       sync.RWMutex
	enforceArgsForCall []struct {
		arg1 string
		arg2 rules.IPTables
		arg3 lager.Logger
	}
	enforceReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Rule) Enforce(arg1 string, arg2 rules.IPTables, arg3 lager.Logger) error {
	fake.enforceMutex.Lock()
	fake.enforceArgsForCall = append(fake.enforceArgsForCall, struct {
		arg1 string
		arg2 rules.IPTables
		arg3 lager.Logger
	}{arg1, arg2, arg3})
	fake.recordInvocation("Enforce", []interface{}{arg1, arg2, arg3})
	fake.enforceMutex.Unlock()
	if fake.EnforceStub != nil {
		return fake.EnforceStub(arg1, arg2, arg3)
	} else {
		return fake.enforceReturns.result1
	}
}

func (fake *Rule) EnforceCallCount() int {
	fake.enforceMutex.RLock()
	defer fake.enforceMutex.RUnlock()
	return len(fake.enforceArgsForCall)
}

func (fake *Rule) EnforceArgsForCall(i int) (string, rules.IPTables, lager.Logger) {
	fake.enforceMutex.RLock()
	defer fake.enforceMutex.RUnlock()
	return fake.enforceArgsForCall[i].arg1, fake.enforceArgsForCall[i].arg2, fake.enforceArgsForCall[i].arg3
}

func (fake *Rule) EnforceReturns(result1 error) {
	fake.EnforceStub = nil
	fake.enforceReturns = struct {
		result1 error
	}{result1}
}

func (fake *Rule) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.enforceMutex.RLock()
	defer fake.enforceMutex.RUnlock()
	return fake.invocations
}

func (fake *Rule) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ rules.Rule = new(Rule)
