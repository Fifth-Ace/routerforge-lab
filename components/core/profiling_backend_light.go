//go:build !core_pprof

package main

import "log"

func startProfilingBackend(listen string) {
	profilingState.mu.Lock()
	profilingState.Running = false
	profilingState.Error = "profiling support requires the core_pprof build"
	profilingState.mu.Unlock()

	log.Printf("profiling unavailable in light Core build")
}
