# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
ok  	couponbatch/cmd/couponbatch	0.001s
ok  	couponbatch/internal/analytics	0.001s
ok  	couponbatch/internal/config	0.001s
ok  	couponbatch/internal/export	0.001s
ok  	couponbatch/internal/httpapi	0.010s
ok  	couponbatch/internal/model	0.001s
?   	couponbatch/internal/observe	[no test files]
ok  	couponbatch/internal/permission	0.001s
ok  	couponbatch/internal/service	0.018s
ok  	couponbatch/internal/store	0.015s
--- FAIL: TestWorkflow28 (0.00s)
    workflow_test.go:105: expected reclaimed state, got confirmed
FAIL
FAIL	couponbatch/internal/workflow	0.021s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/couponbatch): exit `0`
- Frontend build (web): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/couponbatch): exit `0`
- Frontend build (web): exit `0`
