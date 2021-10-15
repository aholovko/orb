/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didorbgenerator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/orb/pkg/activitypub/vocab"
	"github.com/trustbloc/orb/pkg/anchor/subject"
)

var logger = log.New("anchorevent")

const (
	// ID specifies the ID of the generator.
	ID = "https://w3id.org/orb#v0"

	// Namespace specifies the namespace of the generator.
	Namespace = "did:orb"

	// Version specifies the version of the generator.
	Version = uint64(0)

	multihashPrefix          = "did:orb:uAAA"
	multihashPrefixDelimiter = ":"
)

// Generator generates a content object for did:orb anchor events.
type Generator struct {
	*options
}

// Opt defines an option for the generator.
type Opt func(opts *options)

type options struct {
	id        string
	namespace string
	version   uint64
}

// WithNamespace sets the namespace of the generator.
func WithNamespace(ns string) Opt {
	return func(opts *options) {
		opts.namespace = ns
	}
}

// WithVersion sets the version of the generator.
func WithVersion(version uint64) Opt {
	return func(opts *options) {
		opts.version = version
	}
}

// WithID sets the ID of the generator.
func WithID(id string) Opt {
	return func(opts *options) {
		opts.id = id
	}
}

// New returns a new generator.
func New(opts ...Opt) *Generator {
	optns := &options{
		id:        ID,
		namespace: Namespace,
		version:   Version,
	}

	for _, opt := range opts {
		opt(optns)
	}

	return &Generator{
		options: optns,
	}
}

// ID returns the ID of the generator.
func (g *Generator) ID() string {
	return g.id
}

// Namespace returns the Namespace for the DID method.
func (g *Generator) Namespace() string {
	return g.namespace
}

// Version returns the Version of this generator.
func (g *Generator) Version() uint64 {
	return g.version
}

// CreateContentObject creates a content object from the given payload.
func (g *Generator) CreateContentObject(payload *subject.Payload) (vocab.Document, error) {
	if payload.CoreIndex == "" {
		return nil, fmt.Errorf("payload is missing core index")
	}

	if len(payload.PreviousAnchors) == 0 {
		return nil, fmt.Errorf("payload is missing previous anchors")
	}

	var resources []*resource

	for key, value := range payload.PreviousAnchors {
		logger.Debugf("RESOURCE - Key [%s] Value [%s]", key, value)

		resourceID := fmt.Sprintf("%s:%s", multihashPrefix, key)

		var res *resource

		if value == "" {
			res = &resource{ID: resourceID}
		} else {
			pos := strings.LastIndex(value, ":")
			if pos == -1 {
				return nil, fmt.Errorf("invalid previous anchor hashlink[%s] - must contain separator ':'", value)
			}

			res = &resource{ID: resourceID, PreviousAnchor: value[:pos]}
		}

		resources = append(resources, res)
	}

	contentObj := &contentObject{
		Subject: payload.CoreIndex,
		Properties: &propertiesType{
			Generator: g.id,
			Resources: resources,
		},
	}

	contentObjDoc, err := vocab.MarshalToDoc(contentObj)
	if err != nil {
		return nil, fmt.Errorf("marshal content object to document: %w", err)
	}

	return contentObjDoc, nil
}

// CreatePayload creates a payload from the given anchor event.
func (g *Generator) CreatePayload(anchorEvent *vocab.AnchorEventType) (*subject.Payload, error) {
	anchorObj, err := anchorEvent.AnchorObject(anchorEvent.Anchors())
	if err != nil {
		return nil, fmt.Errorf("anchor object for [%s]: %w", anchorEvent.Anchors(), err)
	}

	contentObj := &contentObject{}

	err = anchorObj.ContentObject().Unmarshal(contentObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshal content object: %w", err)
	}

	if contentObj.Subject == "" {
		return nil, fmt.Errorf("content object is missing subject")
	}

	resources := contentObj.Resources()

	operationCount := uint64(len(resources))

	prevAnchors, err := g.getPreviousAnchors(resources, anchorEvent.Parent())
	if err != nil {
		return nil, fmt.Errorf("failed to parse previous anchors from anchorEvent: %w", err)
	}

	return &subject.Payload{
		Namespace:       g.namespace,
		Version:         g.version,
		CoreIndex:       contentObj.Subject,
		OperationCount:  operationCount,
		PreviousAnchors: prevAnchors,
		AnchorOrigin:    anchorEvent.AttributedTo().String(),
		Published:       anchorEvent.Published(),
	}, nil
}

func (g *Generator) getPreviousAnchors(resources []*resource, previous []*url.URL) (map[string]string, error) {
	previousAnchors := make(map[string]string)

	for _, res := range resources {
		suffix, err := removeMultihashPrefix(res.ID)
		if err != nil {
			return nil, err
		}

		var prevAnchor string

		if res.PreviousAnchor != "" {
			suffix, prevAnchor, err = getPreviousAnchorForResource(suffix, res.PreviousAnchor, previous)
			if err != nil {
				return nil, fmt.Errorf("get previous anchor for resource: %w", err)
			}
		}

		previousAnchors[suffix] = prevAnchor
	}

	return previousAnchors, nil
}

type propertiesType struct {
	Generator string      `json:"https://w3id.org/activityanchors#generator,omitempty"`
	Resources []*resource `json:"https://w3id.org/activityanchors#resources,omitempty"`
}

type resource struct {
	ID             string `json:"ID"`
	PreviousAnchor string `json:"previousAnchor,omitempty"`
}

type contentObject struct {
	Subject    string          `json:"subject,omitempty"`
	Properties *propertiesType `json:"properties,omitempty"`
}

func (t *contentObject) Resources() []*resource {
	if t == nil || t.Properties == nil {
		return nil
	}

	return t.Properties.Resources
}

func getPreviousAnchorForResource(suffix, res string, previous []*url.URL) (string, string, error) {
	for _, prev := range previous {
		if !strings.HasPrefix(prev.String(), res) {
			continue
		}

		logger.Debugf("Found previous anchor [%s] for suffix [%s]", prev, suffix)

		return suffix, prev.String(), nil
	}

	return "", "", fmt.Errorf("resource[%s] not found in previous anchor list", res)
}

func removeMultihashPrefix(id string) (string, error) {
	prefix := multihashPrefix + multihashPrefixDelimiter

	if !strings.HasPrefix(id, prefix) {
		return "", fmt.Errorf("ID has to start with %s", prefix)
	}

	return id[len(prefix):], nil
}