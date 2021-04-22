package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/nextwavedevs/drop/business/auth"
	"github.com/nextwavedevs/drop/business/data/studio"
	"github.com/nextwavedevs/drop/business/validate"
	"github.com/nextwavedevs/drop/foundation/database"
	"github.com/nextwavedevs/drop/foundation/web"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)


type studioGroup struct {
	studio studio.Studio
}

func (sg studioGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.studioGroup.query")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	pageNumber, err := strconv.Atoi(params["page"])
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid page format: %s", params["page"]), http.StatusBadRequest)
	}
	rowsPerPage, err := strconv.Atoi(params["rows"])
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid rows format: %s", params["rows"]), http.StatusBadRequest)
	}

	users, err := sg.studio.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "unable to query for users")
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

func (sg studioGroup) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.studioGroup.queryByID")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	usr, err := sg.studio.QueryByID(ctx, v.TraceID, params["id"])
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (sg studioGroup) queryByLocation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.studioGroup.queryByLocation")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	pageNumber, err := strconv.Atoi(params["page"])
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid page format: %s", params["page"]), http.StatusBadRequest)
	}
	rowsPerPage, err := strconv.Atoi(params["rows"])
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid rows format: %s", params["rows"]), http.StatusBadRequest)
	}

	city := params["city"]

	users, err := sg.studio.QueryByLocation(ctx, v.TraceID, pageNumber, rowsPerPage, city)
	if err != nil {
		return errors.Wrap(err, "unable to query for users")
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

func (sg studioGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.studioGroup.create")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var ns studio.NewStudio
	if err := web.Decode(r, &ns); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	std, err := sg.studio.Create(ctx, v.TraceID, ns, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &std)
	}

	return web.Respond(ctx, w, std, http.StatusCreated)
}

func (sg studioGroup) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.studioGroup.update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var upd studio.UpdateStudio
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	params := web.Params(r)
	err := sg.studio.Update(ctx, v.TraceID, params["id"], upd, v.Now)
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s  User: %+v", params["id"], &upd)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (sg studioGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.studioGroup.delete")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	params := web.Params(r)
	err := sg.studio.Delete(ctx, v.TraceID, claims, params["id"])
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}