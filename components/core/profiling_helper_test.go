package main

import "testing"

func TestProfilingHelperPathIsExternal(t *testing.T) {
    if profilingHelper != "/opt/libexec/routerforge-profiler" {
        t.Fatalf("unexpected profiling helper path: %q", profilingHelper)
    }
}
