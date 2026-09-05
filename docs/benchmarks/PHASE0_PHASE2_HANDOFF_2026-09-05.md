# RouterForge optimization handoff — Phase 0 through Phase 2

Date: 2026-09-05

Branch:

`phase2/core-go-opt`

This document records completed work and the current experimental direction.
It does not declare Phase 2 complete and does not enable UPX in production.

---

## Phase 0 — baseline

Phase 0 established the measurement baseline and the production package shape.

Known stable candidate package:

- IPK total: 18,412,671 B
- unpacked payload: 40,887,519 B

Core production-shape stripped executable before Phase 2 optimization:

- 8,519,680 B

Exact Go compatibility baseline for Phase 2 measurements:

- Go 1.21.13
- linux/arm64
- CGO_ENABLED=0

Phase 0 is complete.

---

## Phase 1 — Core v1 parity

Phase 1 locked the Core v1 behavior before optimization.

Final Phase 1 commit:

`3ed173449e1d4adba8ca1cd80d0917b073a7d243`

Phase 1 established:

- portable HTTP parity coverage;
- official parity gate;
- Go 1.27 Windows parity evidence;
- Unix-domain-socket live proxy behavior retained for Unix/Linux;
- `/api/health` POST behavior locked by parity;
- existing API and module behavior preserved.

Known inherited opkg file-descriptor behavior was classified outside the
required Core compatibility contract.

Phase 1 is complete.

---

## Phase 2 — Go Core optimization

Phase 2 began from the exact Phase 1 commit.

### Size attribution

The original stripped no-embed Phase 1 Core:

- 6,291,456 B

Package attribution showed the largest linked contributors included:

- Go runtime;
- net/http;
- crypto/tls;
- main;
- net;
- encoding/json;
- time;
- x/text normalization;
- reflect;
- math/big;
- crypto/x509;
- IDNA and related networking dependencies.

Build flag attribution established:

- `-trimpath` provides a small measurable saving;
- `-s -w` provides the major linker-size saving;
- `-buildid=` produced 0 B standalone size saving in the measured build.

External HTTPS replacement with curl:

- 0 B executable saving.

ReverseProxy replacement spike:

- 65,536 B saving;
- approximately 1.16%;
- rejected as insufficient benefit for the semantic risk.

Embedded marketplace registry gzip spike:

- source JSON compressed strongly;
- stripped ELF size remained unchanged;
- no binary-size benefit;
- rejected.

### Profiling split

The main successful source-level optimization was isolating in-process
`net/http/pprof` behind a build variant.

Default light build:

- no `net/http/pprof` linkage;
- profiling-enabled build remains available separately;
- profiling continues to run in-process when the profiling build is used.

Exact stripped no-embed comparison:

Phase 1:

- 6,291,456 B

Phase 2 light:

- 5,636,096 B

Saving:

- 655,360 B
- 10.42%

Production-shape embedded Core:

Old:

- 8,519,680 B

Phase 2 light:

- 7,864,320 B

Saving:

- 655,360 B

### Hardware memory A/B

Seven cold process runs were performed on the Keenetic test router using
test-only profiling-marker paths so the existing production profiling marker
could not affect the comparison.

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

Production Core on port 2233 remained healthy during testing.

---

## Package compression investigation

Current RouterForge packaging and the examined AWG Manager packaging both use
gzip-compressed tar layers. The examined AWG build scripts do not establish
UPX as the reason for its package size.

### gzip compression levels

A controlled gzip comparison on the same tar stream produced:

- gzip -1: 4,901,327 B
- gzip -6: 4,633,307 B
- gzip -9: 4,620,341 B

gzip -9 versus gzip -6:

- saving: 12,966 B
- saving: 0.2798%

Conclusion:

Changing the normal package compression from gzip -6 class behavior to
gzip -9 is not currently justified by size benefit.

---

## UPX investigation

UPX 5.2.1 was tested.

The official Win64 package SHA256 was verified before use.

UPX was applied only to disposable copies of binaries.

### Initial installed-Core experiment

Installed Core:

- raw: 8,716,288 B

UPX best:

- raw: 4,470,084 B
- raw saving: 48.716%

UPX ultra:

- raw: 3,855,436 B
- raw saving: 55.767%

After gzip -6:

Plain:

- 4,633,369 B

UPX best:

- 4,375,159 B
- saving: 258,210 B
- 5.573%

UPX ultra:

- 3,855,721 B
- saving: 777,648 B
- 16.784%

This installed Core was not the exact Phase 2 benchmark binary, so a second
controlled experiment was performed.

### Exact Phase 2 UPX experiment

Exact Phase 2 light binary:

- 5,636,096 B

UPX best:

- 2,048,412 B

UPX ultra:

- 1,674,264 B

Raw saving with UPX ultra:

- 3,961,832 B
- approximately 70.3%

Both UPX variants passed `upx -t`.

Both variants also started successfully on the Keenetic test router and
served the Core health API.

### Exact Phase 2 runtime A/B

Seven-run averages:

Plain:

- VmRSS: 5233.1 KiB
- VmSize: 1230510.3 KiB
- VmData: 40414.3 KiB
- threads: 5.00
- FDs: 7.00

UPX best:

- VmRSS: 5872.0 KiB
- VmSize: 1230558.3 KiB
- VmData: 40454.3 KiB
- threads: 4.43 average at the sample point
- FDs: 7.00

UPX ultra:

- VmRSS: 5866.9 KiB
- VmSize: 1230452.6 KiB
- VmData: 40344.6 KiB
- threads: 4.29 average at the sample point
- FDs: 7.00

UPX ultra versus plain:

- VmRSS: +633.8 KiB
- VmSize: -57.7 KiB
- VmData: -69.7 KiB
- FDs: unchanged

The lower sampled thread averages in UPX variants are not treated as an
optimization result because sampling occurred shortly after startup.

Production Core on port 2233 remained healthy.

### Exact Phase 2 gzip comparison

gzip -6 representation of the same exact binaries:

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

---

## Current interpretation

UPX ultra is a serious candidate rather than a cosmetic executable-size trick.

For the exact Phase 2 Core it trades approximately:

- 3.96 MiB less installed executable storage;
- 0.61 MiB less gzip-compressed package payload;

for approximately:

- 0.62 MiB additional resident memory.

The trade is potentially favorable on router-class devices where available
RAM is substantially larger than constrained persistent storage.

However, Core-only results are insufficient for a production decision.

UPX is NOT enabled in the production builder at this point.

---

## Current stopping point / next experiment

Before doing further source-level dependency cutting, Phase 2 changes
direction temporarily.

The next experiment is:

**Whole-production RouterForge UPX attribution.**

Goal:

Take the complete set of currently deployed production RouterForge
executables exactly as they exist, without removing dependencies, features,
profiling code, frontend assets, or other "fat".

For every production executable:

1. preserve the original binary;
2. create an UPX ultra-compressed copy;
3. verify the UPX copy;
4. measure original and UPX raw size;
5. measure gzip -6 representation of both;
6. calculate per-component and aggregate storage savings;
7. test executable compatibility on the Keenetic test router;
8. measure aggregate resident-memory cost for the complete running stack.

No production binary is to be replaced during the attribution experiment.

The result will determine whether aggressive source-level size reduction is
still worth prioritizing.

Possible outcomes:

- UPX provides large whole-project storage savings for a small acceptable
  aggregate RAM cost: simplify/defer further Go dependency cutting.
- UPX aggregate RAM/startup/compatibility cost is too high: continue source
  optimization.
- Mixed result: use UPX selectively per module.

Phase 2 remains open until this experiment is complete.
