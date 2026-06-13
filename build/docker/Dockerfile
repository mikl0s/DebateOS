# build/docker/Dockerfile
#
# Multi-stage Docker image for DebateOS speech builds.
#
# Stage 1 (builder): Compiles the static debateos binary using golang:1.24.
# Stage 2 (runtime): archlinux:base-devel (same digest as translators/arch/Dockerfile)
#                    installs archiso + python-yaml; copies binary + translators + schemas.
#
# This image powers BOTH build channels (BLD-01, BLD-02):
#   Channel 1 (local): docker run -v <speech>:/speech -v <out>:/out ghcr.io/mikl0s/debateos:latest
#   Channel 2 (CI):    build/actions/build-speech.yml (workflow_call; container: this image)
#
# SECURITY (T-03-CTX, T-03-IMG):
#   - Build context is the repo root (not HOME); .dockerignore excludes secrets/config dir.
#   - Runtime base is pinned by sha256 digest (identical to translators/arch/Dockerfile pin).
#   - Re-verify the digest quarterly:
#       docker pull archlinux:base-devel && docker inspect --format='{{index .RepoDigests 0}}'
#     Then update BOTH this file AND translators/arch/Dockerfile to the new digest.
#   - Single-pin source of truth: the digest MUST match translators/arch/Dockerfile.
#
# Dated tag: base-devel-20260607.0.541780  (same image, human-readable reference)
# Verified: 2026-06-12

# ─── Stage 1: Build the static debateos binary ───────────────────────────────
FROM golang:1.24 AS builder

WORKDIR /src

# Copy dependency manifests first for Docker layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source tree and build a fully static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/debateos ./cmd/debateos

# ─── Stage 2: archlinux runtime (digest-pinned — MUST match translators/arch/Dockerfile) ──
FROM archlinux:base-devel@sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722

# Install archiso (mkarchiso, libisoburn, squashfs-tools, mtools) and python-yaml
# for in-container generator invocations.  pacman -Sy syncs the DB first.
# --noconfirm: unattended; no prompts in a build container.
RUN pacman -Sy --noconfirm archiso python-yaml && \
    pacman -Q archiso python-yaml && \
    pacman -Scc --noconfirm

# Copy the statically-linked debateos binary from the builder stage.
COPY --from=builder /out/debateos /usr/local/bin/debateos

# Copy translators and schemas into the image so the entrypoint can invoke
# translators/arch/translate without any host-side mounts for tooling.
COPY translators/ /debateos/translators/
COPY schemas/ /debateos/schemas/

# Entrypoint: speech mounted at /speech, ISO output written to /out.
# See build/docker/entrypoint.sh for details.
COPY build/docker/entrypoint.sh /debateos/entrypoint.sh
RUN chmod +x /debateos/entrypoint.sh

WORKDIR /build

ENTRYPOINT ["/debateos/entrypoint.sh"]
