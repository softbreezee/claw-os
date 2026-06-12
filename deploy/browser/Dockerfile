# claw-browser — a long-lived anti-detect browser container.
#
# Built with Podman (podman build / podman run); this Dockerfile is
# OCI-standard so it works with either runtime, but the project uses
# Podman — see deploy/browser/claw-browser.sh.
#
# Why this exists: claw-os itself runs on the physical host (not in a
# container). But the browser layer — camoufox-cli + a full anti-detect
# Firefox — is heavy, stateful, and runs untrusted page content. We keep
# it OUT of the host by running it in this one always-on container.
#
# claw-os drives it from the host via the exec tool:
#     podman exec claw-browser camoufox-cli open <url>
#     podman exec claw-browser camoufox-cli snapshot -i
#
# The container stays up (see deploy/browser/claw-browser.sh) because
# Camoufox runs a background daemon with a 2-3 min cold start; tearing it
# down per call would be unusable.
#
# Adapted from fastclaw's sandbox image, trimmed to "just the browser".

FROM node:22-bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# System deps: Python toolchain + fetch/cert essentials + CJK/emoji
# fonts so Firefox renders real Chinese glyphs instead of tofu boxes on
# 小红书 / 抖音 / A-share pages. `sudo` is required because camoufox-cli's
# `install --with-deps` shells out to `sudo apt-get` for Firefox runtime
# libs even when already root.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        python3 \
        python3-pip \
        python3-venv \
        ca-certificates \
        curl \
        git \
        sudo \
        fonts-noto-cjk \
        fonts-noto-color-emoji \
 && ln -sf /usr/bin/python3 /usr/local/bin/python

# camoufox-cli + common scraping helpers. --break-system-packages is the
# documented PEP 668 escape hatch and safe inside a container.
RUN pip install --break-system-packages --no-cache-dir \
        camoufox-cli \
        requests \
        pillow \
        beautifulsoup4 \
        lxml

# Pull Camoufox (anti-detect Firefox) + its headless runtime libs.
# --with-deps apt-installs Firefox deps, so it must run before we wipe
# apt lists. We let a failure abort the build — a silently broken browser
# would ship an image that looks fine until the first `open` blows up.
RUN camoufox-cli install --with-deps

RUN rm -rf /var/lib/apt/lists/*

# Optional proxy shim: auto-inject --proxy from $HTTPS_PROXY into every
# camoufox-cli call when the caller didn't pass one. Camoufox's browser
# process does NOT honour HTTPS_PROXY env — only the --proxy flag — so
# behind a proxy-only network (e.g. GFW) the first Page.goto fails with
# NS_ERROR_NET_INTERRUPT without this, even though curl/pip work via env.
#
# Move the real binary aside and replace it with a shim on PATH so the
# agent's `camoufox-cli ...` calls flow through transparently.
RUN mv /usr/local/bin/camoufox-cli /usr/local/bin/camoufox-cli-bin \
 && printf '%s\n' \
        '#!/bin/sh' \
        'proxy="${HTTPS_PROXY:-${https_proxy:-${HTTP_PROXY:-${http_proxy:-}}}}"' \
        'if [ -n "$proxy" ]; then' \
        '    for a in "$@"; do' \
        '        case "$a" in' \
        '            --proxy|--proxy=*) exec /usr/local/bin/camoufox-cli-bin "$@" ;;' \
        '        esac' \
        '    done' \
        '    exec /usr/local/bin/camoufox-cli-bin --proxy "$proxy" "$@"' \
        'fi' \
        'exec /usr/local/bin/camoufox-cli-bin "$@"' \
    > /usr/local/bin/camoufox-cli \
 && chmod +x /usr/local/bin/camoufox-cli

WORKDIR /workspace

# Keep the container alive so `docker exec` can reach it. camoufox-cli's
# own daemon starts lazily on the first `open`; this just holds PID 1.
CMD ["tail", "-f", "/dev/null"]
