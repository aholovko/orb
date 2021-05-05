// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"net/http"
	"sync"
)

type HTTPSigner struct {
	SignRequestStub        func(pubKeyID string, req *http.Request) error
	signRequestMutex       sync.RWMutex
	signRequestArgsForCall []struct {
		pubKeyID string
		req      *http.Request
	}
	signRequestReturns struct {
		result1 error
	}
	signRequestReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *HTTPSigner) SignRequest(pubKeyID string, req *http.Request) error {
	fake.signRequestMutex.Lock()
	ret, specificReturn := fake.signRequestReturnsOnCall[len(fake.signRequestArgsForCall)]
	fake.signRequestArgsForCall = append(fake.signRequestArgsForCall, struct {
		pubKeyID string
		req      *http.Request
	}{pubKeyID, req})
	fake.recordInvocation("SignRequest", []interface{}{pubKeyID, req})
	fake.signRequestMutex.Unlock()
	if fake.SignRequestStub != nil {
		return fake.SignRequestStub(pubKeyID, req)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.signRequestReturns.result1
}

func (fake *HTTPSigner) SignRequestCallCount() int {
	fake.signRequestMutex.RLock()
	defer fake.signRequestMutex.RUnlock()
	return len(fake.signRequestArgsForCall)
}

func (fake *HTTPSigner) SignRequestArgsForCall(i int) (string, *http.Request) {
	fake.signRequestMutex.RLock()
	defer fake.signRequestMutex.RUnlock()
	return fake.signRequestArgsForCall[i].pubKeyID, fake.signRequestArgsForCall[i].req
}

func (fake *HTTPSigner) SignRequestReturns(result1 error) {
	fake.SignRequestStub = nil
	fake.signRequestReturns = struct {
		result1 error
	}{result1}
}

func (fake *HTTPSigner) SignRequestReturnsOnCall(i int, result1 error) {
	fake.SignRequestStub = nil
	if fake.signRequestReturnsOnCall == nil {
		fake.signRequestReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.signRequestReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *HTTPSigner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.signRequestMutex.RLock()
	defer fake.signRequestMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *HTTPSigner) recordInvocation(key string, args []interface{}) {
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
