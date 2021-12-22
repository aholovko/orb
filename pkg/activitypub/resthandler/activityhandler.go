/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resthandler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/trustbloc/orb/pkg/activitypub/store/spi"
	"github.com/trustbloc/orb/pkg/activitypub/store/storeutil"
	"github.com/trustbloc/orb/pkg/activitypub/vocab"
	orberrors "github.com/trustbloc/orb/pkg/errors"
)

// NewActivity returns a new 'activities/{id}' REST handler that retrieves a single activity by ID.
func NewActivity(cfg *Config, activityStore spi.Store, verifier signatureVerifier,
	sortOrder spi.SortOrder, tm authTokenManager) *Activity {
	h := &Activity{}

	h.handler = newHandler(ActivitiesPath, cfg, activityStore, h.handle, verifier, sortOrder, tm)

	return h
}

// NewOutbox returns a new 'outbox' REST handler that retrieves a service's outbox.
func NewOutbox(cfg *Config, activityStore spi.Store, verifier signatureVerifier,
	sortOrder spi.SortOrder, tm authTokenManager) *ReadOutbox {
	h := &ReadOutbox{
		Activities: &Activities{
			getID:        getID("outbox"),
			getObjectIRI: getObjectIRI(cfg.ObjectIRI),
		},
	}

	h.Activities.handler = newHandler(OutboxPath, cfg, activityStore, h.handleOutbox, verifier, sortOrder, tm)

	return h
}

// NewInbox returns a new 'inbox' REST handler that retrieves a service's inbox.
func NewInbox(cfg *Config, activityStore spi.Store, verifier signatureVerifier,
	sortOrder spi.SortOrder, tm authTokenManager) *Activities {
	return NewActivities(InboxPath, spi.Inbox, cfg, activityStore,
		getObjectIRI(cfg.ObjectIRI), getID("inbox"), verifier, sortOrder, tm)
}

// NewShares returns a new 'shares' REST handler that retrieves an object's 'Announce' activities.
func NewShares(cfg *Config, activityStore spi.Store, verifier signatureVerifier,
	sortOrder spi.SortOrder, tm authTokenManager) *Activities {
	return NewActivities(SharesPath, spi.Share, cfg, activityStore,
		getObjectIRIFromIDParam, getIDFromParam(cfg.ObjectIRI, SharesPath), verifier, sortOrder, tm)
}

// NewLikes returns a new 'likes' REST handler that retrieves an object's 'Like' activities.
func NewLikes(cfg *Config, activityStore spi.Store, verifier signatureVerifier,
	sortOrder spi.SortOrder, tm authTokenManager) *Activities {
	return NewActivities(LikesPath, spi.Like, cfg, activityStore,
		getObjectIRIFromIDParam, getIDFromParam(cfg.ObjectIRI, LikesPath), verifier, sortOrder, tm)
}

type getIDFunc func(objectIRI *url.URL, req *http.Request) (*url.URL, error)

type getObjectIRIFunc func(req *http.Request) (*url.URL, error)

// Activities implements a REST handler that retrieves activities.
type Activities struct {
	*handler

	refType      spi.ReferenceType
	getID        getIDFunc
	getObjectIRI getObjectIRIFunc
}

// NewActivities returns a new activities REST handler.
func NewActivities(path string, refType spi.ReferenceType, cfg *Config, activityStore spi.Store,
	getObjectIRI getObjectIRIFunc, getID getIDFunc, verifier signatureVerifier,
	sortOrder spi.SortOrder, tm authTokenManager) *Activities {
	h := &Activities{
		refType:      refType,
		getID:        getID,
		getObjectIRI: getObjectIRI,
	}

	h.handler = newHandler(path, cfg, activityStore, h.handle, verifier, sortOrder, tm)

	return h
}

func (h *Activities) handle(w http.ResponseWriter, req *http.Request) {
	ok, _, err := h.Authorize(req)
	if err != nil {
		logger.Errorf("[%s] Error authorizing request: %s", h.endpoint, err)

		h.writeResponse(w, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	if !ok {
		h.writeResponse(w, http.StatusUnauthorized, []byte(unauthorizedResponse))

		return
	}

	h.handleActivityRefsOfType(w, req, h.refType)
}

func (h *Activities) handleActivityRefsOfType(w http.ResponseWriter, req *http.Request, refType spi.ReferenceType) {
	objectIRI, id, err := h.getObjectIRIAndID(req)
	if err != nil {
		logger.Debugf("[%s] Error getting object IRI and ID: %s", h.endpoint, err)

		if orberrors.IsBadRequest(err) {
			h.writeResponse(w, http.StatusBadRequest, []byte(badRequestResponse))
		} else {
			h.writeResponse(w, http.StatusInternalServerError, []byte(internalServerErrorResponse))
		}

		return
	}

	if h.isPaging(req) {
		h.handleActivitiesPage(w, req, objectIRI, id, refType)
	} else {
		h.handleActivities(w, req, objectIRI, id, refType)
	}
}

func (h *Activities) handleActivities(rw http.ResponseWriter, _ *http.Request, objectIRI, id *url.URL,
	refType spi.ReferenceType) {
	activities, err := h.getActivities(objectIRI, id, refType)
	if err != nil {
		logger.Errorf("[%s] Error retrieving %s for object IRI [%s]: %s",
			h.endpoint, h.refType, objectIRI, err)

		h.writeResponse(rw, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	activitiesCollBytes, err := h.marshal(activities)
	if err != nil {
		logger.Errorf("[%s] Unable to marshal %s collection for object IRI [%s]: %s",
			h.endpoint, h.refType, objectIRI, err)

		h.writeResponse(rw, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	h.writeResponse(rw, http.StatusOK, activitiesCollBytes)
}

func (h *Activities) handleActivitiesPage(rw http.ResponseWriter, req *http.Request, objectIRI, id *url.URL,
	refType spi.ReferenceType) {
	var page *vocab.OrderedCollectionPageType

	var err error

	pageNum, ok := h.getPageNum(req)
	if ok {
		page, err = h.getPage(objectIRI, id, refType,
			spi.WithPageSize(h.PageSize),
			spi.WithPageNum(pageNum),
			spi.WithSortOrder(h.sortOrder),
		)
	} else {
		page, err = h.getPage(objectIRI, id, refType,
			spi.WithPageSize(h.PageSize),
			spi.WithSortOrder(h.sortOrder),
		)
	}

	if err != nil {
		logger.Errorf("[%s] Error retrieving page for object IRI [%s]: %s",
			h.endpoint, objectIRI, err)

		h.writeResponse(rw, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	pageBytes, err := h.marshal(page)
	if err != nil {
		logger.Errorf("[%s] Unable to marshal page for object IRI [%s]: %s",
			h.endpoint, objectIRI, err)

		h.writeResponse(rw, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	h.writeResponse(rw, http.StatusOK, pageBytes)
}

func (h *Activities) getActivities(objectIRI, id *url.URL,
	refType spi.ReferenceType) (*vocab.OrderedCollectionType, error) {
	it, err := h.activityStore.QueryReferences(refType,
		spi.NewCriteria(
			spi.WithObjectIRI(objectIRI),
		),
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = it.Close()
		if err != nil {
			logger.Errorf("failed to close iterator: %s", err.Error())
		}
	}()

	firstURL, err := h.getPageURL(id, -1)
	if err != nil {
		return nil, err
	}

	totalItems, err := it.TotalItems()
	if err != nil {
		return nil, fmt.Errorf("failed to get total items from reference query: %w", err)
	}

	lastURL, err := h.getPageURL(id, getLastPageNum(totalItems, h.PageSize, h.sortOrder))
	if err != nil {
		return nil, err
	}

	return vocab.NewOrderedCollection(nil,
		vocab.WithContext(vocab.ContextActivityStreams),
		vocab.WithID(id),
		vocab.WithFirst(firstURL),
		vocab.WithLast(lastURL),
		vocab.WithTotalItems(totalItems),
	), nil
}

func (h *Activities) getPage(objectIRI, id *url.URL, refType spi.ReferenceType,
	opts ...spi.QueryOpt) (*vocab.OrderedCollectionPageType, error) {
	it, err := h.activityStore.QueryActivities(
		spi.NewCriteria(
			spi.WithReferenceType(refType),
			spi.WithObjectIRI(objectIRI),
		), opts...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = it.Close()
		if err != nil {
			logger.Errorf("failed to close iterator: %s", err.Error())
		}
	}()

	options := storeutil.GetQueryOptions(opts...)

	activities, err := storeutil.ReadActivities(it, options.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*vocab.ObjectProperty, len(activities))

	for i, activity := range activities {
		items[i] = vocab.NewObjectProperty(vocab.WithActivity(activity))
	}

	totalItems, err := it.TotalItems()
	if err != nil {
		return nil, fmt.Errorf("failed to get total items from activity query: %w", err)
	}

	id, prev, next, err := h.getIDPrevNextURL(id, totalItems, options)
	if err != nil {
		return nil, err
	}

	return vocab.NewOrderedCollectionPage(items,
		vocab.WithContext(vocab.ContextActivityStreams),
		vocab.WithID(id),
		vocab.WithPrev(prev),
		vocab.WithNext(next),
		vocab.WithTotalItems(totalItems),
	), nil
}

func (h *Activities) getObjectIRIAndID(req *http.Request) (*url.URL, *url.URL, error) {
	objectIRI, err := h.getObjectIRI(req)
	if err != nil {
		return nil, nil, err
	}

	id, err := h.getID(objectIRI, req)
	if err != nil {
		return nil, nil, err
	}

	return objectIRI, id, nil
}

// Activity implements a REST handler that retrieves a single activity by ID.
type Activity struct {
	*handler
}

func (h *Activity) handle(w http.ResponseWriter, req *http.Request) {
	authorized, _, err := h.Authorize(req)
	if err != nil {
		logger.Errorf("[%s] Error authorizing request: %s", h.endpoint, err)

		h.writeResponse(w, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	activityIRI, err := h.getActivityIRI(req)
	if err != nil {
		logger.Debugf("[%s] Get activity IRI: %s", h.endpoint, err)

		h.writeResponse(w, http.StatusBadRequest, []byte(badRequestResponse))

		return
	}

	activity, err := h.activityStore.GetActivity(activityIRI)
	if err != nil {
		if errors.Is(err, spi.ErrNotFound) {
			logger.Debugf("[%s] Activity ID not found [%s]", h.endpoint, activityIRI)

			h.writeResponse(w, http.StatusNotFound, []byte(notFoundResponse))

			return
		}

		logger.Errorf("[%s] Unable to retrieve activity [%s]: %s", h.endpoint, activityIRI, err)

		h.writeResponse(w, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	if !authorized {
		if !activity.To().Contains(vocab.PublicIRI) {
			logger.Debugf("[%s] Unauthorized for activity ID [%s]", h.endpoint, activityIRI)

			h.writeResponse(w, http.StatusUnauthorized, []byte(unauthorizedResponse))

			return
		}
	}

	activityBytes, err := h.marshal(activity)
	if err != nil {
		logger.Errorf("[%s] Unable to marshal activity [%s]: %s", h.endpoint, activityIRI, err)

		h.writeResponse(w, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	logger.Debugf("[%s] Returning activity: %s", h.endpoint, activityBytes)

	h.writeResponse(w, http.StatusOK, activityBytes)
}

func (h *Activity) getActivityIRI(req *http.Request) (*url.URL, error) {
	id := getIDParam(req)

	if id == "" {
		return nil, errors.New("activity ID not specified")
	}

	activityID := fmt.Sprintf("%s/activities/%s", h.ObjectIRI, id)

	logger.Debugf("[%s] Retrieving activity from store [%s]", h.endpoint, activityID)

	activityIRI, err := url.Parse(activityID)
	if err != nil {
		return nil, fmt.Errorf("invalid activity ID [%s]: %w", id, err)
	}

	return activityIRI, nil
}

// ReadOutbox defines an endpoint that retrieves activities from the outbox.
// The caller has access to all activities if they are authorized, otherwise only public activities are returned.
type ReadOutbox struct {
	*Activities
}

func (h *ReadOutbox) handleOutbox(w http.ResponseWriter, req *http.Request) {
	ok, _, err := h.Authorize(req)
	if err != nil {
		logger.Errorf("[%s] Error authorizing request: %s", h.endpoint, err)

		h.writeResponse(w, http.StatusInternalServerError, []byte(internalServerErrorResponse))

		return
	}

	if ok {
		logger.Debugf("[%s] Client authorized. Returning items in outbox.", h.endpoint)

		h.handleActivityRefsOfType(w, req, spi.Outbox)
	} else {
		logger.Debugf("[%s] Client not authorized. Returning only items in outbox marked as public.", h.endpoint)

		h.handleActivityRefsOfType(w, req, spi.PublicOutbox)
	}
}

func getObjectIRIFromIDParam(req *http.Request) (*url.URL, error) {
	id := getIDParam(req)
	if id == "" {
		return nil, orberrors.NewBadRequest(errors.New("id not specified in URL"))
	}

	return url.Parse(id)
}
