# ┌───────────────────────────────────────────────────────────────┐
# │ Justfile for (Go project)                              │
# │                                                               │
# │ Commands: just → show this help message                       │
# └───────────────────────────────────────────────────────────────┘

set shell := ["bash", "-eu", "-o", "pipefail", "-c"]
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]
set dotenv-load := true

PROJECT := "kvstok"
BIN := "./bin/" + PROJECT
BIN_API := "./bin/" + PROJECT + "-api"

# Extract version from VERSION file
version := `cat VERSION`

# Default recipe
default: help

# ─── Help ──────────────────────────────────────────────────────
help:
    @echo "Available commands for {{PROJECT}}:"
    @echo ""
    @echo "=== Development ==="
    @echo " just / just help     → Show this help message"
    @echo " just build / b       → Build both CLI and API binaries"
    @echo " just build-cli       → Build CLI binary only"
    @echo " just build-api       → Build API binary only"
    @echo " just run / r         → Run CLI"
    @echo " just run-api         → Run API server"
    @echo ""
    @echo "=== Quality ==="
    @echo " just fmt             → Format code"
    @echo " just fmt-check       → Check formatting"
    @echo " just vet             → Vet code"
    @echo " just lint            → Run golangci-lint"
    @echo " just test            → Run tests"
    @echo " just check           → Run full quality check (no auto-fix)"
    @echo ""
    @echo "=== Dependencies ==="
    @echo " just deps            → Download dependencies"
    @echo " just tidy            → Tidy go modules"
    @echo " just tidy-check      → Check if go modules are tidy (dry-run)"
    @echo " just update          → Update dependencies"
    @echo ""
    @echo "=== Maintenance ==="
    @echo " just clean           → Remove binaries"
    @echo " just pre-commit      → Full validation before commit"
    @echo ""
    @echo "=== Release ==="
    @echo " just goreleaser-check    → Validate .goreleaser.yml"
    @echo " just goreleaser-snapshot → Local release validation build"
    @echo " just release-verify      → Full release validation"
    @echo " just release-dry-run → Preview release"
    @echo " just release         → Create tag + push"
    @echo " just release-clean   → Delete old release/tag and recreate"
    @echo " just release-local   → Install locally for testing"

# ─── Build & Development ───────────────────────────────────────
build: build-cli build-api

build-cli:
    @echo "🔨 Building {{PROJECT}} CLI..."
    @mkdir -p bin
    go build -o {{BIN}} .

build-api:
    @echo "🔨 Building {{PROJECT}} API..."
    @mkdir -p bin
    go build -o {{BIN_API}} ./api/cmd

run:
    @echo "🚀 Running {{PROJECT}} CLI..."
    go run .

run-api:
    @echo "🚀 Running {{PROJECT}} API server..."
    go run ./api/cmd

b: build
r: run

# ─── Quality ───────────────────────────────────────────────────
fmt:
    @echo "🎨 Formatting code..."
    gofmt -w -s .

fmt-check:
    #!/usr/bin/env bash
    echo "🔍 Checking formatting..."
    files=$(gofmt -l -s . 2>&1)
    if [ -n "$files" ]; then
        echo "❌ Formatting issues found:"
        echo "$files"
        echo ""
        echo "💡 Run 'just fmt' to fix them."
        exit 1
    else
        echo "✅ All files are properly formatted."
    fi

vet:
    @echo "🔎 Running go vet..."
    go vet ./...

lint:
    @echo "🧹 Running golangci-lint..."
    golangci-lint run

test:
    @echo "🧪 Running tests..."
    go test -v -race -covermode=atomic -count=1 ./...

# check: pure validation, no auto-fixes (safe for CI)
check:
    just fmt-check
    just vet
    just lint
    just test
    just tidy-check
    @echo "✅ All checks passed!"

# ─── Dependencies ──────────────────────────────────────────────
deps:
    @echo "📦 Downloading dependencies..."
    go mod download

tidy:
    @echo "🧹 Tidying go modules..."
    go mod tidy

# dry-run: fails if go.mod/go.sum are not already tidy
tidy-check:
    @echo "🔍 Checking go modules..."
    go mod tidy -diff

update:
    @echo "⬆️ Updating dependencies..."
    @echo "⚠️  This updates ALL deps including indirect ones."
    go get -u ./...
    go mod tidy
    @echo "✅ Dependencies updated!"

# ─── Maintenance ───────────────────────────────────────────────
clean:
    @echo "🧹 Cleaning binaries..."
    rm -rf bin/

# pre-commit: auto-fixes first, then validates
pre-commit:
    @echo "🚦 Running pre-commit checks..."
    just fmt
    just vet
    just lint
    just test
    just tidy
    @echo "🎉 Pre-commit checks completed!"

# ─── Release Artifacts Cleanup (local only) ────────────────────
clean-release-artifacts:
    @echo "🧹 Cleaning local release artifacts..."
    rm -rf {{BIN}} {{BIN_API}} 2>/dev/null || true
    rm -rf release/*.tar.gz release/*.zip 2>/dev/null || true
    @echo "→ Local artifacts removed"

# ─── GoReleaser Validation ─────────────────────────────────────

goreleaser-check:
    @echo "🔍 Validating GoReleaser configuration..."
    @echo "⚠️ GoReleaser v2.16 reports deprecation warnings as failures."
    @echo "⚠️ Snapshot build is the authoritative validation."
    goreleaser check || true

goreleaser-snapshot:
    @echo "📦 Running local GoReleaser snapshot build..."
    goreleaser release --snapshot --clean

release-verify:
    @echo "🚦 Running full release validation..."
    just pre-commit
    just goreleaser-snapshot
    @echo "✅ Release validation completed!"

# ─── Release ───────────────────────────────────────────────────
release-dry-run:
    @echo "Current version in VERSION → {{version}}"
    @echo "Tag that will be created → v{{version}}"
    @echo ""
    @echo "This will trigger the GitHub Actions workflow to build official binaries."

release-clean:
    @echo "🧹 Preparing fresh release for v{{version}}..."
    just clean-release-artifacts

    @echo "→ Deleting remote GitHub Release and tag (v{{version}})..."
    gh release delete "v{{version}}" --yes --cleanup-tag 2>/dev/null \
        && echo " → Remote release + tag deleted" \
        || echo " → No previous remote release found"

    @echo "→ Deleting remote tag (v{{version}}) if still present..."
    git push origin --delete "v{{version}}" 2>/dev/null \
        && echo " → Remote tag deleted" \
        || echo " → Remote tag already gone"

    @echo "→ Deleting local tag (v{{version}})..."
    git tag -d "v{{version}}" 2>/dev/null \
        && echo " → Local tag deleted" \
        || echo " → No local tag found"

    @echo "→ Syncing tags from remote..."
    git fetch --tags --force

    @echo ""
    @echo "✅ Clean complete. Run 'just release' to publish again."

release:
    @echo "=== Preparing release v{{version}} ==="

    just release-verify

    @echo "📦 Committing dependency files (if changed)..."
    git add go.mod go.sum VERSION
    git commit -m "chore: release v{{version}}" \
        || echo "→ No changes to commit"

    @echo "🏷️ Creating annotated tag v{{version}}..."
    git tag -a "v{{version}}" -m "Release v{{version}}"

    @echo "⬆️ Pushing commit and tag to GitHub..."
    git push --follow-tags
    # git push origin main --follow-tags

    @echo ""
    @echo "🎉 Tag v{{version}} pushed successfully!"
    @echo "→ GitHub Actions is now building the official binaries and creating the release."

# Local install (for testing only)
release-local:
    @echo "📦 Installing {{PROJECT}} locally..."
    go install .
    go install ./api/cmd
    @echo "✅ {{PROJECT}} installed locally from source"
