package main

import (
	"net/http"
	"os"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
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

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	fieldKeys := []string{"method", "error"}
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

	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)

	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

	requestAuthCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "auth_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestAuthLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "auth_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	key := []byte("supersecret")
	var auth AuthService
	auth = authService{key}
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
	http.Handle("/metrics", promhttp.Handler())
	logger.Log("msg", "HTTP", "addr", ":8080")
	logger.Log("err", http.ListenAndServe(":8080", nil))
}
