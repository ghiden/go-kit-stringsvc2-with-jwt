package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/metrics"
)

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	countResult    metrics.Histogram
	next           StringService
}

func (mw instrumentingMiddleware) Uppercase(ctx context.Context, s string) (output string, err error) {
	defer func(begin time.Time) {
		custCl, _ := ctx.Value(jwt.JWTClaimsContextKey).(*customClaims)
		lvs := []string{"method", "uppercase", "client", custCl.ClientID, "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.next.Uppercase(ctx, s)
	return
}

func (mw instrumentingMiddleware) Count(ctx context.Context, s string) (n int) {
	defer func(begin time.Time) {
		custCl, _ := ctx.(context.Context).Value(jwt.JWTClaimsContextKey).(*customClaims)
		lvs := []string{"method", "uppercase", "client", custCl.ClientID, "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		mw.countResult.Observe(float64(n))
	}(time.Now())

	n = mw.next.Count(ctx, s)
	return
}

type instrumentingAuthMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           AuthService
}

func (mw instrumentingAuthMiddleware) Auth(clientID string, clientSecret string) (token string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "auth", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	token, err = mw.next.Auth(clientID, clientSecret)
	return
}
