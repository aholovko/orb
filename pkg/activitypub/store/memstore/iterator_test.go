/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package memstore

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/orb/pkg/activitypub/store/spi"
	"github.com/trustbloc/orb/pkg/activitypub/vocab"
)

func TestActivityIterator(t *testing.T) {
	const (
		activityID1 = "activity1"
		activityID2 = "activity2"
	)

	activity1 := vocab.NewAnnounceActivity(activityID1, vocab.NewObjectProperty())
	activity2 := vocab.NewAnnounceActivity(activityID2, vocab.NewObjectProperty())

	results := []*vocab.ActivityType{activity1, activity2}

	it := newActivityIterator(results, 5)
	require.NotNil(t, it)
	require.Equal(t, 5, it.TotalItems())

	a, err := it.Next()
	require.NoError(t, err)
	require.NotNil(t, a)
	require.True(t, a.ID() == activityID1)

	a, err = it.Next()
	require.NoError(t, err)
	require.NotNil(t, a)
	require.True(t, a.ID() == activityID2)

	a, err = it.Next()
	require.Error(t, err)
	require.True(t, errors.Is(err, spi.ErrNotFound))
	require.Nil(t, a)

	require.NotPanics(t, it.Close)
}

func TestReferenceIterator(t *testing.T) {
	ref1 := mustParseURL("https://ref_1")
	ref2 := mustParseURL("https://ref_2")

	results := []*url.URL{ref1, ref2}

	it := newReferenceIterator(results, 5)
	require.NotNil(t, it)
	require.Equal(t, 5, it.TotalItems())

	ref, err := it.Next()
	require.NoError(t, err)
	require.NotNil(t, ref)
	require.True(t, ref.String() == ref1.String())

	ref, err = it.Next()
	require.NoError(t, err)
	require.NotNil(t, ref)
	require.True(t, ref.String() == ref2.String())

	a, err := it.Next()
	require.Error(t, err)
	require.True(t, errors.Is(err, spi.ErrNotFound))
	require.Nil(t, a)

	require.NotPanics(t, it.Close)
}
