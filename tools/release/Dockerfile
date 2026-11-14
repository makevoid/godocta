FROM debian:bookworm-slim AS runtime

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl unzip python3 xauth xvfb \
    libasound2 libfontconfig1 libgl1 libgl1-mesa-dri libpulse0 \
    libxcursor1 libxi6 libxinerama1 libxkbcommon0 libxrandr2 && \
    rm -rf /var/lib/apt/lists/*

FROM runtime AS builder
ARG TARGETARCH
COPY godot.sha512 /tmp/godot.sha512
RUN set -eu; \
    case "$TARGETARCH" in arm64) arch=arm64 ;; amd64) arch=x86_64 ;; *) exit 1 ;; esac; \
    base=https://github.com/godotengine/godot/releases/download/4.7.2-stable; \
    editor="Godot_v4.7.2-stable_linux.$arch.zip"; \
    templates=Godot_v4.7.2-stable_export_templates.tpz; \
    cd /tmp; \
    curl --fail --location --retry 3 "$base/$editor" -o "$editor"; \
    curl --fail --location --retry 3 "$base/$templates" -o "$templates"; \
    grep -E " ($editor|$templates)\$" godot.sha512 | sha512sum --check --strict -; \
    unzip -q "$editor"; \
    install -m 755 "${editor%.zip}" /usr/local/bin/godot; \
    mkdir -p /root/.local/share/godot/export_templates/4.7.2.stable; \
    unzip -j -q "$templates" templates/macos.zip templates/linux_release.x86_64 \
      templates/linux_debug.x86_64 templates/version.txt \
      -d /root/.local/share/godot/export_templates/4.7.2.stable; \
    rm -f "$editor" "$templates" "${editor%.zip}" godot.sha512

RUN mkdir -p /opt/godot-notices && \
    curl --fail --location --retry 3 \
      https://raw.githubusercontent.com/godotengine/godot/4.7.2-stable/LICENSE.txt \
      -o /opt/godot-notices/GODOT_LICENSE.txt && \
    curl --fail --location --retry 3 \
      https://raw.githubusercontent.com/godotengine/godot/4.7.2-stable/COPYRIGHT.txt \
      -o /opt/godot-notices/GODOT_COPYRIGHT.txt

COPY export.py /opt/export.py
ENTRYPOINT ["python3", "/opt/export.py"]
