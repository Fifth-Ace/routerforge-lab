# RouterForge Phase 2 — Go Core size baseline

Status: measurement-only baseline

Source baseline:

- Phase 1 completion: 3ed173449e1d4adba8ca1cd80d0917b073a7d243
- Phase 2 branch: phase2/core-go-opt
- Toolchain: go version go1.21.13 windows/amd64
- Target: linux/arm64
- CGO: disabled

## Binary matrix

| Variant | Bytes | MiB | SHA256 |
|---|---:|---:|---|
| noembed-default | 9136798 | 8.714 | 34705f488a1e35e7ae573585bdffd69c2b15b64b41cf2ca514acec1ddc80550f |
| noembed-stripped | 6291456 | 6 | 7f2de89cbb2eefccee1146129033768da857d068bf7a45e15c33cdc4d5b83935 |
| embed-default | 11441905 | 10.912 | f89e5f6d4b8cdb999112be1ba83c2ec7be7e9de1b3669585b0ff10bcb50de0ea |
| embed-stripped | 8519680 | 8.125 | 95247ad9361acf278171fd89bc8b9ce3965bd3f478d54b58d1551ebf6af35dea |

## Immediate deltas

- no-embed stripping saving: 2845342 bytes
- embed stripping saving: 2922225 bytes
- frontend embed delta, default builds: 2305107 bytes

Stripped variants use:

-trimpath -ldflags="-s -w -buildid="

No application source behaviour was changed for these measurements.

## Largest symbols in default no-embed ARM64 binary

BEGIN TOP SYMBOLS
  63f228          1 D net.mptcpAvailable
  63f229          1 D net.hasSOLMPTCP
  63f22a          1 D net/http.http2DebugGoroutines
  63f22b          1 D net/http.http2VerboseLogs
  63f22c          1 D net/http.http2logFrameWrites
  63f22d          1 D net/http.http2logFrameReads
  63f22e          1 D net/http.http2inTests
  63f22f          1 D net/http.omitBundledHTTP2
  63f230          1 D os.testingForceReadDirLstat
  63f231          1 D os.checkWrapErr
  63f232          1 D net/http/httputil.inOurTests
  63f233          1 D sync/atomic.firstStoreInProgress
  63f234          1 D reflect.callGC
  63f235          1 D runtime.useAeshash
  63f236          1 D runtime.iscgo
  63f237          1 D runtime.cgoHasExtraM
  63f238          1 D runtime.arm64HasATOMICS
  63f239          1 D runtime.arm64UseAlignedLoads
  63f23a          1 D runtime.useCheckmark
  63f23b          1 D runtime.metricsInit
  63f23c          1 D runtime.disableMemoryProfiling
  63f23d          1 D runtime.doubleCheckReadMemStats
  63f248          1 D syscall.SocketDisableIPv6
  63f23f          1 D runtime.secureMode
  63f240          1 D runtime.didothers
  63f241          1 D runtime.mainStarted
  63f242          1 D runtime.freezing
  63f243          1 D runtime.casgstatusAlwaysTrack
  63f244          1 D runtime.inForkedChild
  63f245          1 D runtime.islibrary
  63f220          0 d runtime.noptrbss
  63f220          0 D regexp.arrayNoInts
  63f220          0 D internal/godebug.stderr
  63f218          0 d runtime.ebss
  40f24c          0 r runtime.etypes
  40f24c          0 r runtime.erodata
  411c70          0 r runtime.symtab
  411c70          0 r runtime.esymtab
  608ae0          0 d runtime.bss
  608ad0          0 d runtime.edata
  411c80          0 r runtime.pclntab
  5bfd88          0 r runtime.epclntab
  5c01a0          0 d runtime.noptrdata
  5f9cd7          0 d runtime.enoptrdata
  5f9ce0          0 d runtime.data
  3fa9a8          0 r runtime.egcbss
  3fa01d          0 r runtime.egcdata
  3f8af8          0 r runtime.gcbits.*
  39a5c0          0 r go:func.*
  362258          0 r go:string.*
  2d0000          0 R type:*
  2d0000          0 r runtime.types
  2d0000          0 r runtime.rodata
  2ccb20          0 t runtime.etext
       0          0 _ go.go
   11000          0 t runtime.text
  645f50          0 d runtime.enoptrbss
  645f50          0 d runtime.covctrs
  645f50          0 d runtime.ecovctrs
  645f50          0 d runtime.end
END TOP SYMBOLS

## Raw local evidence

Generated locally under .rf-phase2-size/

- sizes.json
- noembed-default-top-symbols.txt
- core-deps.txt
- per-variant Go buildinfo
- four measurement binaries

Raw measurement binaries are intentionally not committed.

## Phase 2 interpretation

This commit is measurement-only.

The first optimization must be selected from measured savings and symbol/package attribution.
Core v1 parity remains mandatory after every optimization change.
