package main

import (
    "flag"
    "log"
    "net"
    "net/http"
    "net/http/pprof"
    "strings"
    "time"
)

func loopbackAddress(address string) bool {
    host, _, err := net.SplitHostPort(address)
    if err != nil {
        return false
    }

    ip := net.ParseIP(strings.Trim(host, "[]"))
    return ip != nil && ip.IsLoopback()
}

func main() {
    listen := flag.String(
        "listen",
        "127.0.0.1:6061",
        "profiling listen address",
    )
    flag.Parse()

    if !loopbackAddress(*listen) {
        log.Fatal("profiling listen address must be loopback-only")
    }

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
        Addr:              *listen,
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
    }

    log.Printf("RouterForge profiler listening on %s", *listen)
    log.Fatal(server.ListenAndServe())
}
