/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package log

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStandardFields(t *testing.T) {
	const module = "test_module"

	u1 := parseURL(t, "https://example1.com")
	u2 := parseURL(t, "https://example2.com")
	u3 := parseURL(t, "https://example3.com")
	hl := parseURL(t, "hl:1234")

	t.Run("console error", func(t *testing.T) {
		stdErr := newMockWriter()

		logger := NewStructured(module,
			WithStdErr(stdErr),
			WithFields(WithServiceName("myservice")),
		)

		logger.Error("Sample error", WithError(errors.New("some error")))

		require.Contains(t, stdErr.Buffer.String(), `Sample error	{"service": "myservice", "error": "some error"}`)
	})

	t.Run("json error", func(t *testing.T) {
		stdErr := newMockWriter()

		logger := NewStructured(module,
			WithStdErr(stdErr), WithEncoding(JSON),
			WithFields(WithServiceName("myservice")),
		)

		logger.Error("Sample error", WithError(errors.New("some error")))

		l := unmarshalLogData(t, stdErr.Bytes())

		require.Equal(t, "myservice", l.Service)
		require.Equal(t, "test_module", l.Logger)
		require.Equal(t, "Sample error", l.Msg)
		require.Contains(t, l.Caller, "log/fields_test.go")
		require.Equal(t, "some error", l.Error)
		require.Equal(t, "error", l.Level)
	})

	t.Run("json fields 1", func(t *testing.T) {
		stdOut := newMockWriter()

		logger := NewStructured(module, WithStdOut(stdOut), WithEncoding(JSON))

		now := time.Now()

		query := &mockObject{Field1: "value1", Field2: 1234}

		logger.Info("Some message",
			WithMessageID("msg1"), WithData([]byte(`{"field":"value"}`)),
			WithActorIRI(u1), WithActivityID(u2), WithActivityType("Create"),
			WithServiceIRI(parseURL(t, u2.String())), WithServiceName("service1"),
			WithServiceEndpoint("/services/service1"),
			WithSize(1234), WithCacheExpiration(12*time.Second),
			WithTargetIRI(u1), WithQueue("queue1"),
			WithHTTPStatus(http.StatusNotFound), WithParameter("param1"),
			WithReferenceType("followers"), WithURI(u2), WithURIs(u1, u2),
			WithSenderURL(u1), WithAnchorURI(u3), WithAnchorEventURI(u3),
			WithAcceptListType("follow"),
			WithURLAdditions(u1, u3),
			WithURLDeletions(u1),
			WithRequestURL(u1), WithRequestBody([]byte(`request body`)), WithResponse([]byte(`response body`)),
			WithRequestHeaders(map[string][]string{"key1": {"v1", "v2"}, "key2": {"v3"}}),
			WithObjectIRI(u1), WithReferenceIRI(u2),
			WithKeyIRI(u1), WithKeyOwnerIRI(u2), WithKeyType("ed25519"),
			WithCurrentIRI(u1), WithNextIRI(u2),
			WithTotal(12), WithType("type1"), WithQuery(query),
			WithAnchorHash("sfsfsdfsd"), WithMinimum(2), WithSuffix("1234"), WithHashlink(hl.String()),
			WithVerifiableCredential([]byte(`{"id":"https://example.com/vc1"}`)),
			WithVerifiableCredentialID("https://example.com/vc1"),
			WithParent("parent1"), WithParents([]string{"parent1", "parent2"}),
			WithProof([]byte(`{"id":"https://example.com/proof1"}`)),
			WithCreatedTime(now), WithWitnessURI(u1), WithWitnessURIs(u1, u2), WithWitnessPolicy("some policy"),
			WithAnchorOrigin(u1.String()), WithOperationType("Create"), WithCoreIndex("1234"),
		)

		t.Logf(stdOut.String())
		l := unmarshalLogData(t, stdOut.Bytes())

		require.Equal(t, `Some message`, l.Msg)
		require.Equal(t, `msg1`, l.MessageID)
		require.Equal(t, `{"field":"value"}`, l.Data)
		require.Equal(t, u1.String(), l.ActorID)
		require.Equal(t, u2.String(), l.ActivityID)
		require.Equal(t, `Create`, l.ActivityType)
		require.Equal(t, `service1`, l.Service)
		require.Equal(t, `/services/service1`, l.ServiceEndpoint)
		require.Equal(t, u2.String(), l.ServiceIri)
		require.Equal(t, 1234, l.Size)
		require.Equal(t, `12s`, l.CacheExpiration)
		require.Equal(t, u1.String(), l.Target)
		require.Equal(t, `queue1`, l.Queue)
		require.Equal(t, 404, l.HTTPStatus)
		require.Equal(t, `param1`, l.Parameter)
		require.Equal(t, `followers`, l.ReferenceType)
		require.Equal(t, u2.String(), l.URI)
		require.Equal(t, []string{u1.String(), u2.String()}, l.URIs)
		require.Equal(t, u3.String(), l.AnchorURI)
		require.Equal(t, u3.String(), l.AnchorEventURI)
		require.Equal(t, `follow`, l.AcceptListType)
		require.Equal(t, []string{u1.String(), u3.String()}, l.Additions)
		require.Equal(t, []string{u1.String()}, l.Deletions)
		require.Equal(t, u1.String(), l.RequestURL)
		require.Equal(t, `request body`, l.RequestBody)
		require.Equal(t, `response body`, l.Response)
		require.Equal(t, map[string][]string{"key1": {"v1", "v2"}, "key2": {"v3"}}, l.RequestHeaders)
		require.Equal(t, u1.String(), l.ObjectIRI)
		require.Equal(t, u2.String(), l.Reference)
		require.Equal(t, u1.String(), l.KeyID)
		require.Equal(t, u2.String(), l.KeyOwnerID)
		require.Equal(t, "ed25519", l.KeyType)
		require.Equal(t, u1.String(), l.Current)
		require.Equal(t, u2.String(), l.Next)
		require.Equal(t, 12, l.Total)
		require.Equal(t, 2, l.Minimum)
		require.Equal(t, "type1", l.Type)
		require.Equal(t, query, l.Query)
		require.Equal(t, "sfsfsdfsd", l.AnchorHash)
		require.Equal(t, "1234", l.Suffix)
		require.Equal(t, hl.String(), l.Hashlink)
		require.Equal(t, `{"id":"https://example.com/vc1"}`, l.VerifiableCredential)
		require.Equal(t, "https://example.com/vc1", l.VerifiableCredentialID)
		require.Equal(t, "parent1", l.Parent)
		require.Equal(t, []string{"parent1", "parent2"}, l.Parents)
		require.Equal(t, `{"id":"https://example.com/proof1"}`, l.Proof)
		require.Equal(t, now.Format("2006-01-02T15:04:05.000Z0700"), l.CreatedTime)
		require.Equal(t, u1.String(), l.WitnessURI)
		require.Equal(t, []string{u1.String(), u2.String()}, l.WitnessURIs)
		require.Equal(t, "some policy", l.WitnessPolicy)
		require.Equal(t, u1.String(), l.AnchorOrigin)
		require.Equal(t, "Create", l.OperationType)
		require.Equal(t, "1234", l.CoreIndex)
	})

	t.Run("json fields 2", func(t *testing.T) {
		stdOut := newMockWriter()

		logger := NewStructured(module, WithStdOut(stdOut), WithEncoding(JSON))

		cfg := &mockObject{Field1: "value1", Field2: 1234}
		aoep := &mockObject{Field1: "value11", Field2: 999}
		rr := &mockObject{Field1: "value22", Field2: 777}
		rm := &mockObject{Field1: "value33", Field2: 888}

		logger.Info("Some message",
			WithActorID(u1.String()), WithTarget(u2.String()),
			WithConfig(&mockObject{Field1: "value1", Field2: 1234}),
			WithRequestURLString(u1.String()),
			WithKeyID("key1"), WithURIString(u1.String()),
			WithAnchorEventURIString(u3.String()), WithAnchorURIString(u3.String()),
			WithHashlinkURI(hl), WithParentURI(u1),
			WithProofDocument(map[string]interface{}{"id": "https://example.com/proof1"}),
			WithWitnessURIString(u1.String()), WithWitnessURIStrings(u1.String(), u2.String()),
			WithHash("hash1"), WithAnchorOriginEndpoint(aoep), WithKey("key1"),
			WithCID("cid1"), WithResolvedCID("cid2"), WithAnchorCID("cid3"),
			WithCIDVersion(1), WithMultihash("fsdfervs"), WithCASData([]byte("cas data")),
			WithDomain(u1.String()), WithLink(u2.String()), WithLinks(u1.String(), u2.String()),
			WithTaskMgrInstanceID("12345"), WithRetries(7), WithMaxRetries(12),
			WithSubscriberPoolSize(30), WithTaskMonitorInterval(5*time.Second),
			WithTaskExpiration(10*time.Second), WithDeliveryDelay(3*time.Second),
			WithOperationID("op1"), WithTaskOwnerID("123"), WithTimeSinceLastUpdate(2*time.Minute),
			WithGenesisTime(1233), WithDID("did:orb:123:456"), WithHRef(u3.String()),
			WithID("id1"), WithResource("res1"), WithResolutionResult(rr),
			WithResolutionModel(rm), WithResolutionEndpoints(u1.String(), u2.String(), u3.String()),
		)

		l := unmarshalLogData(t, stdOut.Bytes())

		require.Equal(t, `Some message`, l.Msg)
		require.Equal(t, u1.String(), l.ActorID)
		require.Equal(t, u2.String(), l.Target)
		require.Equal(t, cfg, l.Config)
		require.Equal(t, u1.String(), l.RequestURL)
		require.Equal(t, "key1", l.KeyID)
		require.Equal(t, u1.String(), l.URI)
		require.Equal(t, u1.String(), l.URI)
		require.Equal(t, u3.String(), l.AnchorEventURI)
		require.Equal(t, u3.String(), l.AnchorURI)
		require.Equal(t, hl.String(), l.Hashlink)
		require.Equal(t, u1.String(), l.Parent)
		require.Equal(t, `{"id":"https://example.com/proof1"}`, l.Proof)
		require.Equal(t, u1.String(), l.WitnessURI)
		require.Equal(t, []string{u1.String(), u2.String()}, l.WitnessURIs)
		require.Equal(t, "hash1", l.Hash)
		require.Equal(t, aoep, l.AnchorOriginEndpoint)
		require.Equal(t, "key1", l.Key)
		require.Equal(t, "cid1", l.CID)
		require.Equal(t, "cid2", l.ResolvedCID)
		require.Equal(t, "cid3", l.AnchorCID)
		require.Equal(t, 1, l.CIDVersion)
		require.Equal(t, "fsdfervs", l.Multihash)

		casData, err := base64.StdEncoding.DecodeString(l.CASData)
		require.NoError(t, err)
		require.Equal(t, "cas data", string(casData))

		require.Equal(t, u1.String(), l.Domain)
		require.Equal(t, u2.String(), l.Link)
		require.Equal(t, []string{u1.String(), u2.String()}, l.Links)
		require.Equal(t, "12345", l.TaskMgrInstanceID)
		require.Equal(t, 7, l.Retries)
		require.Equal(t, 12, l.MaxRetries)
		require.Equal(t, 30, l.SubscriberPoolSize)
		require.Equal(t, "5s", l.TaskMonitorInterval)
		require.Equal(t, "10s", l.TaskExpiration)
		require.Equal(t, "3s", l.DeliveryDelay)
		require.Equal(t, "op1", l.OperationID)
		require.Equal(t, "123", l.TaskOwnerID)
		require.Equal(t, "2m0s", l.TimeSinceLastUpdate)
		require.Equal(t, 1233, l.GenesisTime)
		require.Equal(t, "did:orb:123:456", l.DID)
		require.Equal(t, u3.String(), l.HRef)
		require.Equal(t, "id1", l.ID)
		require.Equal(t, "res1", l.Resource)
		require.Equal(t, rr, l.ResolutionResult)
		require.Equal(t, rm, l.ResolutionModel)
		require.Equal(t, []string{u1.String(), u2.String(), u3.String()}, l.ResolutionEndpoints)
	})
}

type mockObject struct {
	Field1 string
	Field2 int
}

type logData struct {
	Level  string `json:"level"`
	Time   string `json:"time"`
	Logger string `json:"logger"`
	Caller string `json:"caller"`
	Msg    string `json:"msg"`
	Error  string `json:"error"`

	MessageID              string              `json:"message-id"`
	Data                   string              `json:"data"`
	ActorID                string              `json:"actor-id"`
	ActivityID             string              `json:"activity-id"`
	ActivityType           string              `json:"activity-type"`
	ServiceIri             string              `json:"service-iri"`
	Service                string              `json:"service"`
	ServiceEndpoint        string              `json:"service-endpoint"`
	Size                   int                 `json:"size"`
	CacheExpiration        string              `json:"cache-expiration"`
	Target                 string              `json:"target"`
	Queue                  string              `json:"queue"`
	HTTPStatus             int                 `json:"http-status"`
	Parameter              string              `json:"parameter"`
	ReferenceType          string              `json:"reference-type"`
	URI                    string              `json:"uri"`
	URIs                   []string            `json:"uris"`
	Sender                 string              `json:"sender"`
	AnchorURI              string              `json:"anchor-uri"`
	AnchorEventURI         string              `json:"anchor-event-uri"`
	Config                 *mockObject         `json:"config"`
	AcceptListType         string              `json:"accept-list-type"`
	Additions              []string            `json:"additions"`
	Deletions              []string            `json:"deletions"`
	RequestURL             string              `json:"request-url"`
	RequestHeaders         map[string][]string `json:"request-headers"`
	RequestBody            string              `json:"request-body"`
	Response               string              `json:"response"`
	ObjectIRI              string              `json:"object-iri"`
	Reference              string              `json:"reference"`
	KeyID                  string              `json:"key-id"`
	KeyOwnerID             string              `json:"key-owner"`
	KeyType                string              `json:"key-type"`
	Current                string              `json:"current"`
	Next                   string              `json:"next"`
	Total                  int                 `json:"total"`
	Minimum                int                 `json:"minimum"`
	Type                   string              `json:"type"`
	Query                  *mockObject         `json:"query"`
	AnchorHash             string              `json:"anchor-hash"`
	Suffix                 string              `json:"suffix"`
	VerifiableCredential   string              `json:"vc"`
	VerifiableCredentialID string              `json:"vc-id"`
	Hashlink               string              `json:"hashlink"`
	Parent                 string              `json:"parent"`
	Parents                []string            `json:"parents"`
	Proof                  string              `json:"proof"`
	CreatedTime            string              `json:"created-time"`
	WitnessURI             string              `json:"witness-uri"`
	WitnessURIs            []string            `json:"witness-uris"`
	WitnessPolicy          string              `json:"witness-policy"`
	AnchorOrigin           string              `json:"anchor-origin"`
	OperationType          string              `json:"operation-type"`
	CoreIndex              string              `json:"core-index"`
	Hash                   string              `json:"hash"`
	AnchorOriginEndpoint   *mockObject         `json:"anchor-origin-endpoint"`
	Key                    string              `json:"key"`
	CID                    string              `json:"cid"`
	ResolvedCID            string              `json:"resolved-cid"`
	AnchorCID              string              `json:"anchor-cid"`
	CIDVersion             int                 `json:"cid-version"`
	Multihash              string              `json:"multihash"`
	CASData                string              `json:"cas-data"`
	Domain                 string              `json:"domain"`
	Link                   string              `json:"link"`
	Links                  []string            `json:"links"`
	TaskMgrInstanceID      string              `json:"task-mgr-instance"`
	Retries                int                 `json:"retries"`
	MaxRetries             int                 `json:"max-retries"`
	SubscriberPoolSize     int                 `json:"subscriber-pool-size"`
	TaskMonitorInterval    string              `json:"task-monitor-interval"`
	TaskExpiration         string              `json:"task-expiration"`
	DeliveryDelay          string              `json:"delivery-delay"`
	OperationID            string              `json:"operation-id"`
	TaskOwnerID            string              `json:"task-owner-id"`
	TimeSinceLastUpdate    string              `json:"time-since-last-update"`
	GenesisTime            int                 `json:"genesis-time"`
	DID                    string              `json:"did"`
	HRef                   string              `json:"href"`
	ID                     string              `json:"id"`
	Resource               string              `json:"resource"`
	ResolutionResult       *mockObject         `json:"resolution-result"`
	ResolutionModel        *mockObject         `json:"resolution-model"`
	ResolutionEndpoints    []string            `json:"resolution-endpoints"`
}

func unmarshalLogData(t *testing.T, b []byte) *logData {
	t.Helper()

	l := &logData{}

	require.NoError(t, json.Unmarshal(b, l))

	return l
}

func parseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	u, err := url.Parse(raw)
	require.NoError(t, err)

	return u
}
