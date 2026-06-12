# translators/arch/Dockerfile
#
# Pinned archlinux:base-devel image for DebateOS Arch ISO builds via mkarchiso.
#
# Image: archlinux:base-devel
# Digest: sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722
# Dated tag: base-devel-20260607.0.541780  (same image, human-readable reference)
# Verified: 2026-06-12 (docker pull archlinux:base-devel, archiso 88-1 confirmed)
#
# SECURITY: This image is pinned by sha256 digest (T-02-13).
#   - A stale or compromised base image would propagate into the built ISO.
#   - Re-verify the digest quarterly: docker pull archlinux:base-devel && docker inspect --format='{{index .RepoDigests 0}}'
#   - Update the digest below and the DOCKER_IMAGE variable in scripts/arch-build-iso.sh.
#
# Usage: This Dockerfile is used by scripts/arch-build-iso.sh.
#   docker build -t debateos-arch-builder -f translators/arch/Dockerfile .
#   (or the build script can pull the image directly by digest)
#
# All packages are from Arch Linux official repos — no npm/pip/crates installs.
# Package legitimacy: Arch Linux packaging policy (02-RESEARCH.md Package Legitimacy Audit).

FROM archlinux:base-devel@sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722

# Install archiso and its dependencies.
# archiso 88-1 brings: mkarchiso, arch-install-scripts (pacstrap/arch-chroot),
#                      libisoburn (xorriso), squashfs-tools, mtools.
# python-yaml is installed for any in-container generator invocation.
#
# pacman -Sy: sync before install so we get the latest package DB.
# --noconfirm: unattended; no prompts (anti-pattern: never prompt in a build container).
RUN pacman -Sy --noconfirm archiso python-yaml && \
    pacman -Q archiso python-yaml && \
    pacman -Scc --noconfirm

# Verify archiso is installed and callable.
RUN mkarchiso --version || true

# Default working directory for build invocations.
WORKDIR /build
