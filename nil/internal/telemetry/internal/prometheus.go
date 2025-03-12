package internal

import (
	"net/http"
	"strconv"

	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartPrometheusServer(port int) {
	if port == 0 {
		return
	}

	server := http.NewServeMux()
	server.Handle("/metrics", promhttp.Handler())
	go func() {
		check.PanicIfErr(http.ListenAndServe(":"+strconv.Itoa(port), server)) //nolint:gosec
	}()
}
