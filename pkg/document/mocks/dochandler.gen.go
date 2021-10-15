// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/trustbloc/sidetree-core-go/pkg/api/operation"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
)

type Resolver struct {
	ResolveDocumentStub        func(string, ...*operation.AnchoredOperation) (*document.ResolutionResult, error)
	resolveDocumentMutex       sync.RWMutex
	resolveDocumentArgsForCall []struct {
		arg1 string
		arg2 []*operation.AnchoredOperation
	}
	resolveDocumentReturns struct {
		result1 *document.ResolutionResult
		result2 error
	}
	resolveDocumentReturnsOnCall map[int]struct {
		result1 *document.ResolutionResult
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Resolver) ResolveDocument(arg1 string, arg2 ...*operation.AnchoredOperation) (*document.ResolutionResult, error) {
	fake.resolveDocumentMutex.Lock()
	ret, specificReturn := fake.resolveDocumentReturnsOnCall[len(fake.resolveDocumentArgsForCall)]
	fake.resolveDocumentArgsForCall = append(fake.resolveDocumentArgsForCall, struct {
		arg1 string
		arg2 []*operation.AnchoredOperation
	}{arg1, arg2})
	fake.recordInvocation("ResolveDocument", []interface{}{arg1, arg2})
	fake.resolveDocumentMutex.Unlock()
	if fake.ResolveDocumentStub != nil {
		return fake.ResolveDocumentStub(arg1, arg2...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.resolveDocumentReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Resolver) ResolveDocumentCallCount() int {
	fake.resolveDocumentMutex.RLock()
	defer fake.resolveDocumentMutex.RUnlock()
	return len(fake.resolveDocumentArgsForCall)
}

func (fake *Resolver) ResolveDocumentCalls(stub func(string, ...*operation.AnchoredOperation) (*document.ResolutionResult, error)) {
	fake.resolveDocumentMutex.Lock()
	defer fake.resolveDocumentMutex.Unlock()
	fake.ResolveDocumentStub = stub
}

func (fake *Resolver) ResolveDocumentArgsForCall(i int) (string, []*operation.AnchoredOperation) {
	fake.resolveDocumentMutex.RLock()
	defer fake.resolveDocumentMutex.RUnlock()
	argsForCall := fake.resolveDocumentArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Resolver) ResolveDocumentReturns(result1 *document.ResolutionResult, result2 error) {
	fake.resolveDocumentMutex.Lock()
	defer fake.resolveDocumentMutex.Unlock()
	fake.ResolveDocumentStub = nil
	fake.resolveDocumentReturns = struct {
		result1 *document.ResolutionResult
		result2 error
	}{result1, result2}
}

func (fake *Resolver) ResolveDocumentReturnsOnCall(i int, result1 *document.ResolutionResult, result2 error) {
	fake.resolveDocumentMutex.Lock()
	defer fake.resolveDocumentMutex.Unlock()
	fake.ResolveDocumentStub = nil
	if fake.resolveDocumentReturnsOnCall == nil {
		fake.resolveDocumentReturnsOnCall = make(map[int]struct {
			result1 *document.ResolutionResult
			result2 error
		})
	}
	fake.resolveDocumentReturnsOnCall[i] = struct {
		result1 *document.ResolutionResult
		result2 error
	}{result1, result2}
}

func (fake *Resolver) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.resolveDocumentMutex.RLock()
	defer fake.resolveDocumentMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Resolver) recordInvocation(key string, args []interface{}) {
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