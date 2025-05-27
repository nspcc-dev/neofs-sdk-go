package debugprint

import (
	"context"
	"time"

	"go.uber.org/zap"
)

var log *zap.Logger

func SetLogger(l *zap.Logger) { log = l }

var uniqueField = zap.String("grep-by", "debugprint")

type incomingRequest struct {
	typ   string
	id    string
	start time.Time

	details []any // (string, any) pairs

	defaultLogFields []zap.Field
}

var incomingRequestContextKey = "incoming-request"

func incomingRequestFromContext(ctx context.Context) (incomingRequest, bool) {
	res, ok := ctx.Value(incomingRequestContextKey).(incomingRequest)
	return res, ok
}

func NewIncomingRequestContext(parent context.Context, typ, id string, items ...any) context.Context {
	now := time.Now()
	fs := []zap.Field{
		uniqueField,
		zap.String("request", typ),
		zap.String("ID", id),
		zap.String("start", now.UTC().Format(time.RFC3339)),
	}

	for i := 0; i < len(items); i += 2 {
		fs = append(fs, zap.Any(items[i].(string), items[i+1]))
	}

	return context.WithValue(parent, incomingRequestContextKey, incomingRequest{
		typ:              typ,
		id:               id,
		start:            now,
		details:          items,
		defaultLogFields: fs,
	})
}

func LogRequestRecv(ctx context.Context) {
	req, ok := incomingRequestFromContext(ctx)
	if !ok {
		return
	}
	log.Info("start request", req.defaultLogFields...)
}

func LogRequestFin(ctx context.Context) {
	req, ok := incomingRequestFromContext(ctx)
	if !ok {
		return
	}
	log.Info("finish request", append(req.defaultLogFields,
		zap.Stringer("elapsed", time.Since(req.start)),
	)...)
}

type RequestStage struct {
	ctx   context.Context
	name  string
	start time.Time
}

func LogRequestStageStart(ctx context.Context, stage string) RequestStage {
	req, ok := incomingRequestFromContext(ctx)
	if !ok {
		return RequestStage{ctx: ctx}
	}
	log.Info("starting stage '"+stage+"'...", append(req.defaultLogFields,
		zap.Stringer("since start", time.Since(req.start)),
	)...)
	return RequestStage{
		ctx:   ctx,
		name:  stage,
		start: time.Now(),
	}
}

func LogRequestStageFinish(stage RequestStage) {
	req, ok := incomingRequestFromContext(stage.ctx)
	if !ok {
		return
	}
	log.Info("finished stage '"+stage.name+"'", append(req.defaultLogFields,
		zap.Stringer("elapsed", time.Since(stage.start)),
	)...)
}
