# Work Log

## WineBot Integration Status

Result:
- `supragoflow.exe` runs correctly under Wine when executed directly as `winebot` user in WineBot container.
- WineBot launcher path (`scripts/run-app.sh` headless attached flow) is the problem area for CLI-style EXEs.

Artifacts and issue filed:
- WineBot issue: `https://github.com/mark-e-deyoung/WineBot/issues/2`
- Repro + logs + patch draft gist: `https://gist.github.com/mark-e-deyoung/c96406c7cfc7ba6c4d99eebe64e51048`
- Local artifact bundle: `/tmp/winebot-supragoflow-issue`

## Resume Plan (after WineBot check)

1. Pull/check WineBot issue updates and merged fix.
2. Re-run WineBot validation for SupraGoFlow EXE:
   - `./scripts/run-app.sh /apps/supragoflow.exe --mode headless --args "--version --json"`
3. If fixed, capture successful output and update this file.
4. Optionally add/strengthen SupraGoFlow CI path that validates WineBot launcher compatibility (not just bare Wine).

## Quick Resume Commands

From SupraGoFlow repo:

```bash
cd /home/mark/Projects/SupraGoFlow/workspace/supragoflow
git pull --ff-only
./scripts/gg test
./scripts/gg build windows amd64
```

From WineBot repo (current repro path):

```bash
cd /home/mark/Projects/WineBot
./scripts/run-app.sh /apps/supragoflow.exe --mode headless --args "--version --json"
```
