# PhysK TDrive large-file fork

Fork of [jeetrex17/TDrive](https://github.com/jeetrex17/TDrive).

The `master` branch mirrors upstream. The `large-file-support` branch
contains our minimal, reviewable changes.

## What changes

- Raise the stored logical file cap from **40 GiB to 256 GiB**.
- Raise the per-file logical part cap from **32 to 1,024**.
- Keep Telegram's individual document part limit at **1,900 MiB**.
- Use the same upstream multipart manifests, encryption, projections,
  upload retry handling and download format.
- All upload, download and restore paths share these constants. There is
  **no new backup format** and no staging/splitting required in our mapper.

This is not an unlimited upload size. Bounds are retained to avoid
pathological part counts and enormous memory/metadata requirements. The
stored-size cap includes encryption overhead, so files very close to
256 GiB may need a little headroom.

## Binary compatibility and restoration

Files larger than 40 GiB uploaded by this fork **will not download through
unmodified TDrive v2.2.0** because upstream's resumable download rejects
stored files exceeding 40 GiB or 32 parts.

To restore an oversized archive, use a CLI or desktop build from this
branch. Builds of Linux amd64, Windows amd64 and macOS arm64 CLI are
provided in GitHub Actions artifacts. Ordinary files remain compatible
with upstream.

The CLI binary and daemon need to be from the same forked build for the
new limits to take effect. Replacing only the executable while a
v2.2.0 daemon is running **does not update the running daemon**. Stop
active uploads first, stop the old daemon, start the patched version
and unlock the encryption vault. The mapper must remain paused while
transitioning to prevent accidental mixed-version writes.

Large encrypted uploads may still take hours and Telegram can apply
repeated FLOOD_WAIT cooldowns. Do not rely on this as the only copy of
irreplaceable recordings. Interrupted encrypted multipart uploads may
need to restart depending on TDrive's current resume implementation.

## Keeping up with upstream

The fork has two branches:

- `master`: a clean upstream tracking branch, updated by fast-forward.
- `large-file-support`: the working branch with our delta.

`.github/workflows/fork-sync-build.yml` runs weekly or manually.
It fetches `jeetrex17/TDrive` master, fast-forwards the mirror
and merges upstream into the patched branch. Then it runs Go tests
and cross-compiles the three CLIs. **Only if tests and builds
pass** does it push both branches and publish build artifacts.

If changes conflict or tests fail, it stops without force-pushing;
the previous working branch stays available. Inspect the failed
Actions run and fix it manually. The workflow does not deploy to
your NAS, desktop or t3code automatically.

## Build manually

From the repository root with the Go toolchain in `go.mod`:

```bash
go test ./backend/services/file ./backend/projection ./backend/crypto ./backend/daemon
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags '-s -w -X main.buildVersion=physk-large-256GiB' \
  -o dist/tdrive-linux-amd64 ./cmd/tdrive
```

Platform-specific desktop UI builds require the Wails v3 build dependencies
and frontend assets; the CLI cross-build does **not** build the GUI.
