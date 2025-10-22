package server

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/smart-kart/framework/env"
	"github.com/smart-kart/framework/logger"
)

// RunProfiler starts the profiler server
func RunProfiler() {
	port := env.GetOrDefault("PROFILER_PORT", "6060")
	addr := ":" + port

	log := logger.New()
	log.Info("profiler server listening on %s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Error("profiler server failed: %v", err)
	}
}