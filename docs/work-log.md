# Work Log

## WineBot Integration Status

Result:
- `supragoflow.exe` runs correctly under Wine when executed directly as `winebot` user in WineBot container.
- WineBot launcher path (`scripts/run-app.sh` headless attached flow) was problematic for CLI-style EXEs but is now fixed.
- WineBot issue #2 (launcher behavior) is closed/fixed in WineBot v0.9.5+.

Artifacts and issue filed:
- WineBot issue: `https://github.com/mark-e-deyoung/WineBot/issues/2` (Closed)
- Confirmed fix availability in `ghcr.io/mark-e-deyoung/winebot:latest-rel-runner`.

## Resume Plan (Resolved)

1. WineBot issue verified as closed.
2. CI updated to use WineBot container for Windows smoke testing, ensuring launcher compatibility.

## Quick Resume Commands

From SupraGoFlow repo:

```bash
cd /home/mark/Projects/SupraGoFlow/workspace/supragoflow
git pull --ff-only
./scripts/gg test
./scripts/gg build windows amd64
```
