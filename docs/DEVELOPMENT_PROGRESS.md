# RouterForge Development Progress

This document records work that has actually been completed and measured.

It complements the development/master plan:

- the plan describes what RouterForge intends to investigate or build;
- this document records what was actually done;
- rejected approaches remain recorded so the same experiments are not repeated
  without new evidence.

Last updated: 2026-09-05.

---

# Phase 0 — Baseline

Status: **DONE**

The initial production baseline was captured before optimization work.

Known complete candidate package:

- IPK size: 18,412,671 B
- unpacked payload: 40,887,519 B

Original production-shape stripped embedded Core:

- 8,519,680 B

Exact compatibility/size baseline used for optimization measurements:

- Go 1.21.13
- linux/arm64
- CGO_ENABLED=0

Phase 0 established the reference point against which subsequent changes are
measured.

---

# Phase 1 — Core v1 parity

Status: **DONE**

Final Phase 1 commit:

`3ed173449e1d4adba8ca1cd80d0917b073a7d243`

Completed work:

- portable HTTP parity coverage;
- official parity gate;
- Go 1.27 Windows parity evidence;
- Module ABI v1 Core behavior locked;
- Unix-domain-socket live proxy behavior retained for Unix/Linux;
- API behavior preserved;
- `/api/health` POST behavior explicitly locked by parity;
- production semantics established before optimization.

An inherited opkg file-descriptor behavior was classified separately rather
than being made part of the required Core compatibility contract.

---

# Phase 2 — Go Core optimization

Status: **ACTIVE**

Phase 2 started from the exact Phase 1 contract.

## 2.1 Size attribution

Status: **DONE**

Original stripped no-embed Phase 1 Core:

- 6,291,456 B

Investigated contributors included:

- Go runtime;
- net/http;
- crypto/tls;
- crypto/x509;
- networking;
- encoding/json;
- reflect;
- math/big;
- x/text normalization;
- IDNA;
- RouterForge's own embedded data and code.

Build flag attribution established:

- `-trimpath` gives a small measurable saving;
- `-s -w` gives the major linker-size saving;
- `-buildid=` produced no standalone executable-size saving in the measured
  build.

## 2.2 External HTTPS replacement spike

Status: **REJECTED**

Replacing external HTTPS calls with curl produced:

- 0 B stripped executable saving.

Conclusion:

Do not trade architecture/semantics for this approach.

## 2.3 ReverseProxy replacement spike

Status: **REJECTED**

Replacing `httputil.ReverseProxy` with a narrower Unix-socket client produced:

- 65,536 B saving;
- approximately 1.16%.

Conclusion:

Benefit was too small for the additional semantic and maintenance risk.

## 2.4 Embedded registry compression spike

Status: **REJECTED**

Marketplace registry JSON:

- raw JSON: 34,679 B
- gzip representation: approximately 3.9 KiB

Despite strong source-data compression, the final stripped ELF size did not
change.

Conclusion:

No executable-size benefit.

---

# Phase 2.5 — Profiling build split

Status: **DONE**

In-process `net/http/pprof` was moved behind a build variant.

The default light Core no longer links pprof.

A profiling-capable Core variant remains available so profiling continues to
run inside the process it profiles.

Exact stripped no-embed comparison:

Phase 1:

- 6,291,456 B

Phase 2 light:

- 5,636,096 B

Saving:

- 655,360 B
- 10.42%

Production-shape embedded comparison:

Old:

- 8,519,680 B

Phase 2 light:

- 7,864,320 B

Saving:

- 655,360 B

---

# Phase 2.6 — Core runtime memory A/B

Status: **DONE**

Seven cold process runs were performed on the Keenetic test router.

A test-only profiling marker path was used so the existing production
profiling marker could not contaminate the experiment.

Phase 1 average:

- VmRSS: 5992.0 KiB
- VmSize: 1231556.6 KiB
- VmData: 40868.6 KiB
- threads: 5
- FDs: 7

Phase 2 light average:

- VmRSS: 5205.7 KiB
- VmSize: 1230437.1 KiB
- VmData: 40341.1 KiB
- threads: 5
- FDs: 7

Phase 2 minus Phase 1:

- VmRSS: -786.3 KiB
- VmRSS: -13.1%
- VmSize: -1119.5 KiB
- VmData: -527.5 KiB
- threads: unchanged
- FDs: unchanged

Production Core on port 2233 remained healthy during the test.

Conclusion:

The profiling split improves both executable size and resident memory.

---

# Phase 2.7 — gzip package compression levels

Status: **DONE / NO PRODUCTION CHANGE**

Controlled comparison of the same tar stream:

- gzip -1: 4,901,327 B
- gzip -6: 4,633,307 B
- gzip -9: 4,620,341 B

gzip -9 versus gzip -6:

- saving: 12,966 B
- saving: 0.2798%

Conclusion:

Increasing gzip level is not a useful RouterForge optimization.

---

# Phase 2.8 — Core UPX attribution

Status: **DONE**

UPX version tested:

- UPX 5.2.1

The official UPX archive checksum was verified before use.

UPX was always applied to disposable copies during attribution.

Exact Phase 2 light Core:

- plain: 5,636,096 B
- UPX best: 2,048,412 B
- UPX ultra: 1,674,264 B

UPX ultra raw saving:

- 3,961,832 B
- approximately 70.3%

gzip -6 representation of the same exact Phase 2 binaries:

Plain:

- 2,312,596 B

UPX best:

- 2,002,981 B
- saving: 309,615 B
- 13.388%

UPX ultra:

- 1,674,234 B
- saving: 638,362 B
- 27.604%

Both UPX variants passed UPX integrity verification and ran successfully on
the Keenetic test router.

Seven-run runtime A/B:

Plain:

- VmRSS: 5233.1 KiB

UPX best:

- VmRSS: 5872.0 KiB
- delta: +638.9 KiB

UPX ultra:

- VmRSS: 5866.9 KiB
- delta: +633.8 KiB

Conclusion:

UPX ultra provides substantially better storage efficiency than UPX best for
essentially the same measured Core RSS cost.

---

# Phase 2.9 — Whole-production UPX attribution

Status: **DONE**

Goal:

Determine whether compression is valuable enough to reduce the priority of
additional source-level size cutting.

The experiment used the complete original Go production executable set
without removing functionality or dependencies.

The System module used its preserved pre-C-test Go production binary, not the
temporary C prototype.

Original production executable set:

| Executable | Plain bytes | UPX ultra bytes |
| --- | ---: | ---: |
| routerforge | 8,716,288 | 3,855,436 |
| routerforge-admin | 4,653,056 | 1,396,584 |
| routerforge-dns | 5,963,776 | 1,767,800 |
| routerforge-network | 5,046,272 | 1,500,988 |
| routerforge-storage | 5,046,272 | 1,500,972 |
| routerforge-system | 4,915,200 | 1,490,020 |
| routerforge-thermal | 5,046,272 | 1,500,972 |

Totals:

Plain:

- 39,387,136 B
- 37.562 MiB

UPX ultra:

- 13,012,772 B
- 12.410 MiB

Raw storage saving:

- 26,374,364 B
- 25.153 MiB
- 66.962%

gzip -6 representation:

Plain:

- 17,244,434 B
- 16.446 MiB

UPX ultra:

- 13,012,639 B
- 12.410 MiB

Compressed payload saving:

- 4,231,795 B
- 4.036 MiB
- 24.540%

`routerforge-storage` and `routerforge-thermal` were also observed to be
byte-identical in the examined production set. Deduplication is a separate
future optimization and was not used to inflate the UPX result.

---

# Phase 2.10 — Whole-production UPX runtime A/B

Status: **DONE**

Both the complete plain stack and complete UPX-ultra stack were started from
test paths alongside the existing production installation.

Smoke result:

- plain Core: PASS
- plain Admin: PASS
- plain DNS: PASS
- plain Network: PASS
- plain Storage: PASS
- plain System: PASS
- plain Thermal: PASS
- UPX Core: PASS
- UPX Admin: PASS
- UPX DNS: PASS
- UPX Network: PASS
- UPX Storage: PASS
- UPX System: PASS
- UPX Thermal: PASS
- Core HTTP health: PASS

Seven complete stack runs were measured.

Plain stack average:

- VmRSS: 38,161.7 KiB
- VmSize: 8,613,748.6 KiB
- VmData: 291,612.6 KiB
- threads at sample point: 30.43
- FDs: 51.00

UPX-ultra stack average:

- VmRSS: 44,296.0 KiB
- VmSize: 8,614,346.9 KiB
- VmData: 292,126.9 KiB
- threads at sample point: 31.71
- FDs: 49.00

UPX minus plain:

- VmRSS: +6,134.3 KiB
- VmRSS: +16.074%
- VmSize: +598.3 KiB
- VmData: +514.3 KiB

Production Core on port 2233 remained healthy before and after the complete
experiment.

The sampled thread and FD differences are not treated as functional
optimizations because the processes were sampled shortly after startup and
the DNS test path did not exactly reproduce every production connection.

Primary measured trade:

- approximately 25.153 MiB less installed executable storage;
- approximately 4.036 MiB less gzip-compressed executable payload;
- approximately 5.99 MiB additional aggregate resident memory.

---

# Phase 2.11 — Production decision

Status: **APPROVED FOR AARCH64 PRODUCTION IMPLEMENTATION**

Decision:

Use UPX ultra compression for RouterForge AArch64 production executable
packages.

Reasons:

1. Full production executable storage fell from 37.562 MiB to 12.410 MiB.
2. Raw executable storage decreased by 66.962%.
3. gzip-compressed representation still decreased by 24.540%.
4. All seven tested UPX binaries passed integrity checks.
5. All seven tested UPX processes started successfully on the Keenetic
   AArch64 test router.
6. Core health remained correct.
7. The measured aggregate RSS increase was approximately 5.99 MiB.
8. For the targeted router class this RAM/storage trade is considered
   favorable.

Production policy:

- AArch64 RouterForge executable packages are UPX-compressed by default.
- UPX compression must be followed by `upx -t`.
- A compression/integrity failure fails the package build.
- A plain-build escape hatch remains available for diagnosis.
- MIPS/MIPSel are NOT automatically moved to UPX by this result.
- MIPS/MIPSel require architecture-specific runtime validation before their
  production policy changes.

This decision does not invalidate the successful Phase 2 light-Core work.
Source-level changes that improve both storage and RAM remain valuable.

What changes is prioritization:

Further high-risk dependency cutting for very small size gains is deferred
while production packaging captures the much larger measured UPX saving.

---

# Next

Phase 2 productionization:

- integrate pinned UPX into AArch64 package builds;
- enforce UPX integrity checking;
- add CI validation that AArch64 executable payloads are UPX packed;
- retain plain MIPS/MIPSel package behavior;
- build the complete beta package set;
- validate CI;
- promote the exact validated production commit.

After productionization:

- measure installed package sizes from actual generated IPKs;
- optionally investigate storage/thermal binary deduplication;
- separately validate UPX on MIPS/MIPSel hardware;
- continue only source optimizations that offer meaningful size, RAM,
  reliability, or portability gains.
