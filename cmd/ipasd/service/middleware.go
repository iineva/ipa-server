package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func LoggingMiddleware(logger log.Logger, name string, debug bool) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		var logging log.Logger
		if debug {
			logging = level.NewFilter(logger, level.AllowDebug())
		} else {
			logging = level.NewFilter(logger, level.AllowInfo())
		}
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			level.Info(logging).Log(
				"modle", name,
				"status", "start",
				"request", fmt.Sprintf("%+v", request),
			)
			defer func(begin time.Time) {
				level.Debug(logging).Log(
					"modle", name,
					"status", "done",
					"request", fmt.Sprintf("%+v", request),
					"response", fmt.Sprintf("%+v", response),
				)
				if err != nil {
					level.Info(logging).Log(
						"err", err,
						"modle", name,
						"status", "done",
						"took", time.Since(begin),
					)
				} else {
					level.Info(logging).Log(
						"modle", name,
						"status", "done",
						"took", time.Since(begin),
					)
				}
			}(time.Now())
			response, err = next(ctx, request)
			return
		}
	}
}
