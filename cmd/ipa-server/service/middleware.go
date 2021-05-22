package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// type logmw struct {
// 	logger log.Logger
// 	Service
// }

// func LoggingMiddleware(logger log.Logger, debug bool) ServiceMiddleware {
// 	if debug {
// 		logger = level.NewFilter(logger, level.AllowDebug())
// 	} else {
// 		logger = level.NewFilter(logger, level.AllowInfo())
// 	}
// 	return func(next Service) Service {
// 		return logmw{logger, next}
// 	}
// }

// func (mw logmw) List(publicURL string) ([]Item, error) {
// 	defer func(begin time.Time) {
// 		level.Info(logging).Log(
// 			"err", err,
// 			"took", time.Since(begin),
// 		)
// 		level.Debug(logging).Log(
// 			"request", fmt.Sprintf("%+v", request),
// 			"response", fmt.Sprintf("%+v", response),
// 		)
// 	}(time.Now())
// 	return mw.Service.List(publicURL)
// }

// func (mw logmw) log() {
// 	level.Info(mw.logger).Log(
// 		"err", err,
// 		"took", time.Since(begin),
// 	)
// 	level.Debug(logging).Log(
// 		"request", fmt.Sprintf("%+v", request),
// 		"response", fmt.Sprintf("%+v", response),
// 	)
// }

func LoggingMiddleware(logger log.Logger, name string, debug bool) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		var logging log.Logger
		if debug {
			logging = level.NewFilter(logger, level.AllowDebug())
		} else {
			logging = level.NewFilter(logger, level.AllowInfo())
		}
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				level.Debug(logging).Log(
					"modle", name,
					"request", fmt.Sprintf("%+v", request),
					"response", fmt.Sprintf("%+v", response),
				)
				if err != nil {
					level.Info(logging).Log(
						"err", err,
						"modle", name,
						"took", time.Since(begin),
					)
				} else {
					level.Info(logging).Log(
						"modle", name,
						"took", time.Since(begin),
					)
				}
			}(time.Now())
			response, err = next(ctx, request)
			return
		}
	}
}
