###############################################################################
# Super GOLANG Dev Container
#
# Go version is automatically read from the project's go.mod.
#
# Example:
#
#   go 1.26.3
#
# The container will install Go 1.26.3 automatically.
###############################################################################


###############################################################################
# Stage 1 - Bootstrap Go and build development tools
###############################################################################

FROM debian:bookworm AS tools

ENV DEBIAN_FRONTEND=noninteractive

ARG GO_ARCH=amd64

WORKDIR /tmp


###############################################################################
# System packages required to install Go and build tools
###############################################################################

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        wget \
        git \
        bash \
        build-essential \
    && rm -rf /var/lib/apt/lists/*


###############################################################################
# Read Go version from go.mod
###############################################################################

COPY go.mod /tmp/go.mod

RUN set -eux; \
    GO_VERSION="$(sed -nE \
        's/^go[[:space:]]+([0-9]+\.[0-9]+(\.[0-9]+)?).*$/\1/p' \
        /tmp/go.mod | head -n1)"; \
    \
    if [ -z "${GO_VERSION}" ]; then \
        echo "ERROR: Could not determine Go version from go.mod"; \
        exit 1; \
    fi; \
    \
    echo "Detected Go version: ${GO_VERSION}"; \
    \
    wget -q \
        "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" \
        -O "/tmp/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"; \
    \
    rm -rf /usr/local/go; \
    \
    tar \
        -C /usr/local \
        -xzf "/tmp/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"; \
    \
    rm "/tmp/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"; \
    \
    /usr/local/go/bin/go version


###############################################################################
# Go environment
###############################################################################

ENV PATH="/usr/local/go/bin:/usr/local/bin:${PATH}"

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOPROXY=https://proxy.golang.org,direct \
    GOSUMDB=sum.golang.org \
    GOBIN=/opt/go-tools/bin


###############################################################################
# Go development tools
###############################################################################

RUN mkdir -p "${GOBIN}"


RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    \
    TOOLS="\
        golang.org/x/tools/gopls@latest \
        golang.org/x/tools/cmd/goimports@latest \
        mvdan.cc/gofumpt@latest \
        github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
        honnef.co/go/tools/cmd/staticcheck@latest \
        golang.org/x/vuln/cmd/govulncheck@latest \
        github.com/go-delve/delve/cmd/dlv@latest \
        github.com/air-verse/air@latest \
        gotest.tools/gotestsum@latest \
        github.com/kyoh86/richgo@latest \
    "; \
    \
    for tool in ${TOOLS}; do \
        echo; \
        echo "============================================================"; \
        echo "Installing ${tool}"; \
        echo "============================================================"; \
        go install "${tool}"; \
    done; \
    \
    echo; \
    echo "Go tools installation completed."; \
    echo; \
    echo "Go version:"; \
    go version; \
    echo; \
    echo "Installed tools:"; \
    ls -lah "${GOBIN}"


###############################################################################
# Stage 2 - Development container
###############################################################################

FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04


###############################################################################
# Environment
###############################################################################

ENV DEBIAN_FRONTEND=noninteractive


###############################################################################
# System packages
###############################################################################

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        gcc \
        g++ \
        make \
        curl \
        wget \
        unzip \
        ca-certificates \
        gnupg \
        sqlite3 \
        libsqlite3-dev \
        ripgrep \
        fd-find \
        bat \
        git \
        git-delta \
        hyperfine \
        jq \
        tree \
        htop \
        lsof \
    && rm -rf /var/lib/apt/lists/*


###############################################################################
# GitHub CLI
###############################################################################

RUN mkdir -p /etc/apt/keyrings && \
    curl -fsSL \
        https://cli.github.com/packages/githubcli-archive-keyring.gpg \
        -o /etc/apt/keyrings/githubcli-archive-keyring.gpg && \
    chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
        > /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends gh && \
    rm -rf /var/lib/apt/lists/*


###############################################################################
# just command runner
###############################################################################

RUN curl --proto '=https' --tlsv1.2 -sSf \
    https://just.systems/install.sh \
    | bash -s -- --to /usr/local/bin


###############################################################################
# Go SDK
#
# IMPORTANT:
# The Go version is determined from go.mod in the "tools" stage.
# We copy the resulting Go installation into the final container.
###############################################################################

COPY --from=tools /usr/local/go /usr/local/go


###############################################################################
# Go development tools
###############################################################################

COPY --from=tools /opt/go-tools/bin/ /usr/local/bin/


###############################################################################
# Go environment
###############################################################################

ENV PATH="/usr/local/go/bin:/usr/local/bin:${PATH}"

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOPROXY=https://proxy.golang.org,direct \
    GOSUMDB=sum.golang.org


###############################################################################
# Shell helpers
###############################################################################

RUN cat >> /etc/bash.bashrc <<'EOF'

alias ll='ls -lah'
alias fmt='gofumpt -w .'
alias lint='golangci-lint run'
alias test='gotestsum'
alias air='air'

EOF


###############################################################################
# Workspace
###############################################################################

USER vscode

WORKDIR /workspaces


###############################################################################
# Validation
###############################################################################

RUN set -eux; \
    go version; \
    git --version; \
    gh --version; \
    just --version; \
    gopls version; \
    gofumpt -version; \
    golangci-lint version; \
    staticcheck -debug.version; \
    govulncheck -version; \
    dlv version


###############################################################################
# Container health
###############################################################################

HEALTHCHECK \
    --interval=5m \
    --timeout=10s \
    --start-period=30s \
    --retries=3 \
    CMD \
        go version >/dev/null 2>&1 && \
        gopls version >/dev/null 2>&1 && \
        golangci-lint version >/dev/null 2>&1 \
        || exit 1