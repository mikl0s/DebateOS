# translators/debian/Dockerfile
#
# Pinned debian:stable image for DebateOS Debian ISO builds via live-build (lb build).
#
# Image: debian:stable
# Digest: sha256:fa0ca9c113cbc97c1e2eb40d7012b43bdfb70abe4218229c366de911c5b32cd2
# Verified: 2026-06-13 (docker inspect debian:stable; live-build 1:20250505+deb13u1 confirmed)
#
# SECURITY: This image is pinned by sha256 digest (T-04-13).
#   - A stale or compromised base image would propagate into the built ISO.
#   - Re-verify the digest quarterly:
#     docker pull debian:stable && docker inspect --format='{{index .RepoDigests 0}}'
#   - Update the digest below and the DOCKER_IMAGE variable in scripts/debian-build-iso.sh.
#
# Usage: This Dockerfile is used by scripts/debian-build-iso.sh.
#   docker build -t debateos-debian-builder -f translators/debian/Dockerfile .
#   (or the build script can pull the image directly by digest)
#
# All packages are from Debian official apt repos — no npm/pip/crates installs.
# Package legitimacy: 04-RESEARCH.md §Package Legitimacy Audit (all OK).

FROM debian:stable@sha256:fa0ca9c113cbc97c1e2eb40d7012b43bdfb70abe4218229c366de911c5b32cd2

# Install live-build and its dependencies.
# live-build 1:20250505+deb13u1: lb config + lb build — generates Debian live ISO with embedded d-i
# debootstrap 1.0.141: bootstrap minimal Debian rootfs (used by lb bootstrap stage)
# squashfs-tools: mksquashfs compression for the live filesystem
# xorriso: ISO creation (ISO9660 + El Torito; live-build calls it internally)
# python3: generator runtime (for in-container generate() invocations if needed)
#
# apt-get update: sync package DB before install.
# --no-install-recommends: minimal image; live-build pulls its own deps.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        live-build \
        debootstrap \
        squashfs-tools \
        xorriso \
        python3 \
        python3-yaml \
        jq && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Verify live-build is installed and callable.
RUN lb --version

# Default working directory for build invocations.
WORKDIR /build
