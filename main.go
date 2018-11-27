package main

import (
	"crypto/subtle"
	"net/http"
	"os"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	jwt "github.com/dgrijalva/jwt-go"
	gokitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
)

const (
	basicAuthUser = "prometheus"
	basicAuthPass = "password"
)

func methodControl(method string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func basicAuth(username string, password string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorised\n"))
			return
		}

		h.ServeHTTP(w, r)
	})
}

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	fieldKeys := []string{"method", "client", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{}) // no fields here

	var svc StringService
	svc = stringService{}
	svc = loggingMiddleware{logger, svc}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	key := []byte("supersecret")
	keys := func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}

	jwtOptions := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(authErrorEncoder),
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerBefore(gokitjwt.HTTPToContext()),
	}

	uppercaseHandler := httptransport.NewServer(
		gokitjwt.NewParser(keys, jwt.SigningMethodHS256, gokitjwt.MapClaimsFactory)(makeUppercaseEndpoint(svc)),
		decodeUppercaseRequest,
		encodeResponse,
		jwtOptions...,
	)

	countHandler := httptransport.NewServer(
		gokitjwt.NewParser(keys, jwt.SigningMethodHS256, gokitjwt.MapClaimsFactory)(makeCountEndpoint(svc)),
		decodeCountRequest,
		encodeResponse,
		jwtOptions...,
	)

	authFieldKeys := []string{"method", "error"}
	requestAuthCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "auth_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, authFieldKeys)
	requestAuthLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "auth_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, authFieldKeys)

	// API clients database
	var clients = map[string]string{
		"mobile": "m_secret",
		"web":    "w_secret",
	}
	var auth AuthService
	auth = authService{key, clients}
	auth = loggingAuthMiddleware{logger, auth}
	auth = instrumentingAuthMiddleware{requestAuthCount, requestAuthLatency, auth}

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(authErrorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	authHandler := httptransport.NewServer(
		makeAuthEndpoint(auth),
		decodeAuthRequest,
		encodeResponse,
		options...,
	)
	http.Handle("/auth", methodControl("POST", authHandler))

	http.Handle("/uppercase", methodControl("POST", uppercaseHandler))
	http.Handle("/count", methodControl("POST", countHandler))
	http.Handle("/metrics", basicAuth(basicAuthUser, basicAuthPass, promhttp.Handler()))
	logger.Log("msg", "HTTP", "addr", ":8080")
	logger.Log("err", http.ListenAndServe(":8080", nil))
}
