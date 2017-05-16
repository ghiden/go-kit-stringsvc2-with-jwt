package main

import (
	"context"
	"time"

	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
)

type loggingMiddleware struct {
	logger log.Logger
	next   StringService
}

func (mw loggingMiddleware) Uppercase(ctx context.Context, s string) (output string, err error) {
	defer func(begin time.Time) {
		custCl, _ := ctx.Value(jwt.JWTClaimsContextKey).(*customClaims)
		_ = mw.logger.Log(
			"method", "uppercase",
			"client", custCl.ClientID,
			"input", s,
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.next.Uppercase(ctx, s)
	return
}

func (mw loggingMiddleware) Count(ctx context.Context, s string) (n int) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "count",
			"input", s,
			"n", n,
			"took", time.Since(begin),
		)
	}(time.Now())

	n = mw.next.Count(ctx, s)
	return
}

type loggingAuthMiddleware struct {
	logger log.Logger
	next   AuthService
}

func (mw loggingAuthMiddleware) Auth(clientID string, clientSecret string) (token string, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "auth",
			"clientID", clientID,
			"token", token,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	token, err = mw.next.Auth(clientID, clientSecret)
	return
}
