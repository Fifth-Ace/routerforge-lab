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
  6265a0     101496 D runtime.mheap_
  6164a0      65792 D runtime.trace
  5f1560      34679 d main..gobytes.1
  60e720      32128 D runtime.semtable
  5ec960      19426 D vendor/golang.org/x/text/unicode/norm.decomps
  5e85e0      17280 D vendor/golang.org/x/net/idna.idnaValues
  5e4520      16576 D vendor/golang.org/x/text/unicode/bidi.bidiValues
  40bba0      13996 r runtime.findfunctab
  5e14a0      12416 D vendor/golang.org/x/text/unicode/norm.nfkcValues
   b5a10      11856 T unicode.map.init.1
  5de920      11136 D strconv.detailedPowersOfTen
   95920      10832 T time.parse
   928f0       9888 T time.Time.appendFormat
  2b31b0       9696 T main.integrationCatalog
  643c00       9040 D runtime.memstats
  5dc700       8720 D vendor/golang.org/x/net/idna.idnaSparseValues
  1a2790       8480 T crypto/tls.(*Conn).readRecordOrCCS
  60c7a0       8048 D runtime.cpuprof
  16ad00       7792 T encoding/asn1.parseField
  641de0       7696 D crypto/internal/edwards25519.basepointNafTablePrecomp
  1b2f50       7472 T crypto/tls.(*clientHelloMsg).marshal
  40f380       7472 r runtime.typelink
   e5130       7200 T fmt.(*pp).printValue
   f4eb0       7136 T encoding/json.(*decodeState).literalStore
  2a3440       6944 T runtime/pprof.(*profileBuilder).emitLocation
  250be0       6464 T net/http.(*Transport).dialConn
  5d9560       6144 D vendor/golang.org/x/text/unicode/norm.nfcValues
   26880       5920 T runtime.initMetrics
  277c90       5792 T regexp/syntax.dumpInst
   bf9f0       5776 T reflect.Value.call
  1b8690       5728 T crypto/tls.(*clientHelloMsg).unmarshal
   f3750       5568 T encoding/json.(*decodeState).object
  1b9cf0       5488 T crypto/tls.(*serverHelloMsg).marshal
  293120       5168 T internal/profile.(*Profile).String
  26d260       5120 T regexp/syntax.(*compiler).compile
   ba110       5072 T reflect.deepValueEqual
  5d81e0       4992 D vendor/golang.org/x/net/idna.idnaIndex
  269260       4912 T net/http/httputil.(*ReverseProxy).ServeHTTP
  1d4050       4896 T mime.FormatMediaType
  280120       4864 T regexp.makeOnePass.func1
  5d6ee0       4856 D math/rand.rngCooked
  1cced0       4784 T log.formatHeader
  243210       4752 T net/http.(*socksDialer).connect
  11a870       4624 T net.(*Resolver).goLookupIPCNAMEOrder
  271200       4624 T regexp/syntax.(*parser).factor
   fd190       4480 T encoding/json.typeFields
  2a06b0       4464 T runtime/pprof.(*profileBuilder).pbMapping
  2b20b0       4352 T main.builtinModuleCatalog
  16fbd0       4288 T encoding/asn1.makeBody
   5ac60       4272 T runtime.selectgo
   9edb0       4208 T time.LoadLocationFromTZData
  23b6d0       4176 T net/http.(*chunkWriter).writeHeader
  285b90       4160 T internal/profile.(*Profile).postDecode
  29ccb0       4144 T runtime/pprof.writeHeapInternal
   88670       4128 T syscall.forkAndExecInChild1
  607ac0       4112 D runtime.itabTableInit
   e7570       4096 T fmt.(*pp).doPrintf
  13dec0       4080 T crypto/internal/nistec/fiat.p521Mul
  191600       3984 T crypto/x509.parseCertificate
  1612b0       3968 T crypto/elliptic.(*CurveParams).addJacobian
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

## Candidate experiment: optional pprof linkage

Measurement method:

- exact Go 1.21.13
- linux/arm64 measurement target
- CGO disabled
- no frontend embed
- stripped size measured with -trimpath and -s -w -buildid=
- separate unstripped experimental binary used for symbol verification
- net/http/pprof references removed temporarily for measurement only
- profiling.go restored byte-for-byte immediately after builds

Results:

- normal stripped Core: 6291456 bytes
- experimental stripped Core without pprof closure: 5636096 bytes
- pprof linkage cost: 655360 bytes
- reduction: 10.42 percent
- remaining runtime/pprof, internal/profile, or net/http/pprof symbol hits: 0

The experimental source modification was not committed.

The Go 1.21.13 Core test suite passed again after the original profiling source was restored.
The official Core v1 parity gate also passed after restoration.

Interpretation:

The optional pprof implementation contributes 655360 bytes, or 640 KiB, to the stripped always-on Core binary.
This is 10.42 percent of the current 6291456-byte stripped no-embed Core baseline.

This saving is large enough to make pprof separation the first measured Phase 2 optimization candidate.

The optimization must preserve the frozen profiling contract while removing the heavy pprof implementation from the always-on Core binary.

## Architecture spike: split profiler helper

Goal:

Measure a light always-on Core retaining profiling marker/config/status/middleware behaviour while moving net/http/pprof into a separate helper process.

Measurement:

- Go 1.21.13
- linux/arm64
- CGO disabled
- stripped
- frontend not embedded
- production profiling.go modified only temporarily and restored byte-for-byte
- temporary helper not committed

Results:

- current stripped Core: 6291456 bytes
- experimental light Core: 5636096 bytes
- always-on Core saving: 655360 bytes
- always-on Core reduction: 10.42 percent
- standalone profiler helper: 5111808 bytes
- combined light Core plus helper: 10747904 bytes
- combined disk delta versus current Core: 4456448 bytes
- heavy pprof symbol hits remaining in light Core: 0

Interpretation:

The always-on Core saving is material and the split architecture remains a viable optimization candidate.

Shipping the helper inside the mandatory Core package would increase total installed binary bytes, so the helper should be considered as an optional profiling package rather than unconditional Core payload.

No production implementation was committed by this spike.
Original Core source was restored before Go 1.21.13 tests and the official Core v1 parity gate were run.

## Candidate experiment: Go external HTTPS closure

Measurement temporarily replaced external registry, release-index, checksum/small HTTPS fetches, and verified release-asset downloads with /opt/bin/curl.
Module ABI HTTP server, Unix-socket HTTP client, and reverse proxy remained unchanged.

Results:

- current light Core: 5636096 bytes
- experimental Core without Go external HTTPS clients: 5636096 bytes
- saving: 0 bytes
- reduction: 0 percent
- remaining TLS/x509/IDNA symbol hits in the experimental binary: 585

The modified source files were restored byte-for-byte before tests.
No production source implementation is included in this measurement commit.

## Candidate experiment: httputil ReverseProxy closure

Measurement temporarily replaced net/http/httputil.ReverseProxy with a manual http.Client forwarder over the existing Unix-socket http.Transport.
Module routing, Unix transport, and the rest of net/http remained present.

Results:

- current light Core: 5636096 bytes
- experimental manual-proxy Core: 5570560 bytes
- saving: 65536 bytes
- reduction: 1.16 percent

The modified source file was restored byte-for-byte before tests.
No production source implementation is included in this measurement commit.

## Production build flag attribution

Exact Go 1.21.13, linux/arm64, CGO disabled, embedded frontend, light profiling backend.

Sizes:

- default production-shape build: 10510193 bytes
- trimpath only: 10442511 bytes
- strip only (-s -w): 7929856 bytes
- empty build ID only: 10510193 bytes
- trimpath plus strip: 7864320 bytes
- current production flags: 7864320 bytes

Isolated deltas:

- trimpath standalone saving: 67682 bytes
- strip standalone saving: 2580337 bytes
- empty build ID standalone saving: 0 bytes
- empty build ID saving on top of trimpath plus strip: 0 bytes
- current combined saving: 2645873 bytes

## Current light Core package-symbol attribution

The following values sum go tool nm symbol sizes from an unstripped linux/arm64 light Core.
They are attribution/ranking data, not stripped-file byte accounting.

- net/http: 422405 bytes (412.5 KiB)
- crypto/tls: 195369 bytes (190.8 KiB)
- main: 167487 bytes (163.6 KiB)
- net: 136146 bytes (133 KiB)
- encoding/json: 83524 bytes (81.6 KiB)
- time: 77116 bytes (75.3 KiB)
- vendor/golang.org/x/text/unicode/norm: 75282 bytes (73.5 KiB)
- reflect: 74177 bytes (72.4 KiB)
- math/big: 57136 bytes (55.8 KiB)
- crypto/x509: 56356 bytes (55 KiB)
- crypto/internal/nistec: 52996 bytes (51.8 KiB)
- strconv: 50729 bytes (49.5 KiB)
- vendor/golang.org/x/net/idna: 44736 bytes (43.7 KiB)
- encoding/asn1: 39832 bytes (38.9 KiB)
- syscall: 38884 bytes (38 KiB)
- fmt: 38704 bytes (37.8 KiB)
- crypto/internal/nistec/fiat: 32112 bytes (31.4 KiB)
- os: 31562 bytes (30.8 KiB)
- strings: 25744 bytes (25.1 KiB)
- crypto/aes: 24322 bytes (23.8 KiB)
## Hardware runtime A/B: Phase1 versus light Phase2

A/B measurement was performed on the test Keenetic using exact Phase1 and Phase2 source revisions, Go 1.21.13, linux/arm64, CGO disabled, no embedded frontend, and identical `-trimpath -ldflags="-s -w"` build flags.

Both test binaries used a test-only profiling marker path so the installed production profiling marker and the production Core process were not modified.

Binary sizes:

- Phase1 Core: 6291456 bytes
- Phase2 light Core: 5636096 bytes
- saving: 655360 bytes
- reduction: 10.42 percent

Seven cold process runs per variant were measured after one second of settling.

Average Phase1:

- VmRSS: 5992.0 KiB
- VmSize: 1231556.6 KiB
- VmData: 40868.6 KiB
- threads: 5.00
- file descriptors: 7.00

Average Phase2:

- VmRSS: 5205.7 KiB
- VmSize: 1230437.1 KiB
- VmData: 40341.1 KiB
- threads: 5.00
- file descriptors: 7.00

Phase2 minus Phase1:

- VmRSS: -786.3 KiB (-13.1 percent)
- VmSize: -1119.5 KiB
- VmData: -527.5 KiB
- threads: unchanged
- file descriptors: unchanged

`smaps_rollup` was unavailable on the test router, so PSS was not measured.

Startup timing from this run is not treated as evidence because the BusyBox-compatible probe had one-second resolution and both variants became ready before the first interval.

The installed production Core remained running throughout the experiment and `/api/health` on port 2233 returned successfully after the A/B run.

Conclusion: removing the heavy pprof implementation from the default light Core is both a flash-size and resident-memory optimization. The optional in-process `core_pprof` build variant remains available when runtime Go profiling is required.
