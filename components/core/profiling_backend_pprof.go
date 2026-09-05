//go:build core_pprof

package main

import (
	"log"
	"net/http"
	"net/http/pprof"
	"time"
)

func startProfilingBackend(listen string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	for _, name := range []string{
		"allocs",
		"block",
		"goroutine",
		"heap",
		"mutex",
		"threadcreate",
	} {
		mux.Handle("/debug/pprof/"+name, pprof.Handler(name))
	}

	server := &http.Server{
		Addr:              listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		profilingState.mu.Lock()
		profilingState.Running = true
		profilingState.StartedAt = time.Now()
		profilingState.mu.Unlock()

		log.Printf("profiling listener enabled on %s", listen)
		err := server.ListenAndServe()

		profilingState.mu.Lock()
		profilingState.Running = false
		if err != nil && err != http.ErrServerClosed {
			profilingState.Error = err.Error()
			log.Printf("profiling listener failed: %v", err)
		}
		profilingState.mu.Unlock()
	}()
}
