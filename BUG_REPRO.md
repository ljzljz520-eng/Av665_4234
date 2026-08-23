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
ok  	fire-equipment-control/cmd/firecontrol	0.143s
ok  	fire-equipment-control/internal/attachment	0.004s
ok  	fire-equipment-control/internal/audit	0.005s
ok  	fire-equipment-control/internal/dashboard	0.005s
ok  	fire-equipment-control/internal/dispatch	0.014s
ok  	fire-equipment-control/internal/domain	0.006s
ok  	fire-equipment-control/internal/importexport	0.005s
ok  	fire-equipment-control/internal/inspection	0.009s
ok  	fire-equipment-control/internal/query	0.006s
ok  	fire-equipment-control/internal/registry	0.001s
?   	fire-equipment-control/internal/report	[no test files]
--- FAIL: TestBusiness022Regression (0.12s)
    concurrency_test.go:28: both barrier confirmations must be recorded: [{JobID:pending-V-122-01 Actor:操作员甲 Error:pending confirmation consumed twice} {JobID:pending-V-122-02 Actor:操作员乙 Error:<nil>}]
FAIL
FAIL	fire-equipment-control/internal/service	0.149s
ok  	fire-equipment-control/internal/storage	0.131s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/firecontrol): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/firecontrol): exit `0`
