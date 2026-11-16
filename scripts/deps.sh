#!/usr/bin/env bash
#
# Checks or installs the development toolchain for My World Cup App.
#
#   scripts/deps.sh check     # report what's installed/missing (exit 1 if anything is missing)
#   scripts/deps.sh install   # install OS packages + mise + every tool pinned in mise.toml
#
# Supported platforms: Ubuntu 22.04 / 24.04 / 26.04 (amd64) and macOS (amd64 / arm64).
# Other platforms get a warning but the script still tries its best.
#
# Docker Engine, Docker Compose v2 and Docker Buildx are NOT installed here (they
# need root daemons / Docker Desktop and per-distro repos); `check` only verifies them.
#
# Must stay compatible with macOS's stock bash 3.2: no associative arrays,
# no mapfile, no ${var,,}.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MISE_TOML="${ROOT_DIR}/mise.toml"

# OS packages needed before mise can take over.
BASE_TOOLS="git make curl"

if [ -t 1 ]; then
  C_OK=$'\033[32m'; C_WARN=$'\033[33m'; C_ERR=$'\033[31m'; C_DIM=$'\033[2m'; C_RST=$'\033[0m'
else
  C_OK=""; C_WARN=""; C_ERR=""; C_DIM=""; C_RST=""
fi

missing=0

ok()      { printf "  ${C_OK}[OK]${C_RST}      %-16s %s\n" "$1" "$2"; }
miss()    { printf "  ${C_ERR}[MISSING]${C_RST} %-16s %s\n" "$1" "$2"; missing=1; }
warn()    { printf "${C_WARN}WARNING:${C_RST} %s\n" "$*" >&2; }
info()    { printf "${C_DIM}==>${C_RST} %s\n" "$*"; }
die()     { printf "${C_ERR}ERROR:${C_RST} %s\n" "$*" >&2; exit 1; }
have()    { command -v "$1" >/dev/null 2>&1; }

# ---------------------------------------------------------------------------
# Platform detection
# ---------------------------------------------------------------------------

OS="$(uname -s)"
ARCH="$(uname -m)"
DISTRO=""
DISTRO_VERSION=""

detect_platform() {
  case "$OS" in
    Linux)
      if [ -r /etc/os-release ]; then
        # shellcheck disable=SC1091
        DISTRO="$(. /etc/os-release && echo "${ID:-}")"
        # shellcheck disable=SC1091
        DISTRO_VERSION="$(. /etc/os-release && echo "${VERSION_ID:-}")"
      fi
      ;;
  esac
}

platform_label() {
  case "$OS" in
    Darwin) echo "macOS $(sw_vers -productVersion 2>/dev/null || echo '?') ($ARCH)" ;;
    Linux)  echo "${DISTRO:-linux} ${DISTRO_VERSION:-?} ($ARCH)" ;;
    *)      echo "$OS ($ARCH)" ;;
  esac
}

check_supported_platform() {
  case "$OS" in
    Darwin)
      case "$ARCH" in
        x86_64|arm64) return 0 ;;
      esac
      ;;
    Linux)
      if [ "$DISTRO" = "ubuntu" ] && [ "$ARCH" = "x86_64" ]; then
        case "$DISTRO_VERSION" in
          22.04|24.04|26.04) return 0 ;;
        esac
      fi
      ;;
  esac
  warn "$(platform_label) is not an officially supported platform (see REQUIREMENTS.md)."
  warn "Supported: Ubuntu 22.04/24.04/26.04 amd64, macOS amd64/arm64. Continuing anyway."
}

# ---------------------------------------------------------------------------
# mise helpers
# ---------------------------------------------------------------------------

# find_mise prints the mise binary path (it may be installed but not yet on PATH).
find_mise() {
  if have mise; then
    command -v mise
  elif [ -x "${HOME}/.local/bin/mise" ]; then
    echo "${HOME}/.local/bin/mise"
  elif [ -x /opt/homebrew/bin/mise ]; then
    echo /opt/homebrew/bin/mise
  elif [ -x /usr/local/bin/mise ]; then
    echo /usr/local/bin/mise
  fi
}

# pinned_tools prints "tool version" pairs from mise.toml's [tools] table.
pinned_tools() {
  awk '
    /^\[/ { in_tools = ($0 == "[tools]"); next }
    in_tools && /^[[:space:]]*[A-Za-z0-9_.:\/-]+[[:space:]]*=/ {
      line = $0
      sub(/#.*/, "", line)
      split(line, kv, "=")
      name = kv[1]; ver = kv[2]
      gsub(/[[:space:]"]/, "", name); gsub(/[[:space:]"]/, "", ver)
      if (name != "" && ver != "") print name, ver
    }
  ' "$MISE_TOML"
}

# ---------------------------------------------------------------------------
# check
# ---------------------------------------------------------------------------

check_base_tools() {
  info "Base tools (installed by the OS package manager)"
  local tool
  for tool in $BASE_TOOLS; do
    if have "$tool"; then ok "$tool" "$(command -v "$tool")"; else miss "$tool" "run: make install-deps"; fi
  done
}

check_mise_tools() {
  info "Toolchain pinned in mise.toml"
  local mise_bin
  mise_bin="$(find_mise)"
  if [ -z "$mise_bin" ]; then
    miss "mise" "run: make install-deps  (or see https://mise.jdx.dev/getting-started.html)"
    pinned_tools | while read -r tool version; do
      printf "  ${C_DIM}[SKIPPED]${C_RST} %-16s %s (needs mise)\n" "$tool" "$version"
    done
    return
  fi
  ok "mise" "$("$mise_bin" --version 2>/dev/null | awk '{print $1}') ($mise_bin)"

  # Loop in the current shell (not a pipe) so `missing` propagates.
  local tool version where
  while read -r tool version; do
    if where="$(cd "$ROOT_DIR" && "$mise_bin" where "${tool}@${version}" 2>/dev/null)" && [ -n "$where" ]; then
      ok "$tool" "$version"
    else
      miss "$tool" "$version not installed — run: make install-deps"
    fi
  done <<EOF
$(pinned_tools)
EOF

  if ! have go || ! have helm; then
    warn "mise tools are installed but not on PATH in this shell. Activate mise:"
    warn "  bash: echo 'eval \"\$(mise activate bash)\"' >> ~/.bashrc"
    warn "  zsh:  echo 'eval \"\$(mise activate zsh)\"'  >> ~/.zshrc"
  fi
}

check_docker() {
  info "Container tooling (install separately — not managed by this script)"
  if have docker; then
    ok "docker" "$(docker version --format '{{.Client.Version}}' 2>/dev/null || command -v docker)"
    if ! docker info >/dev/null 2>&1; then
      warn "docker is installed but the daemon isn't reachable (not running, or your user isn't in the 'docker' group)."
    fi
  else
    miss "docker" "https://docs.docker.com/engine/install/ubuntu/ | https://docs.docker.com/desktop/setup/install/mac-install/"
  fi
  if docker compose version >/dev/null 2>&1; then
    ok "docker compose" "$(docker compose version --short 2>/dev/null)"
  else
    miss "docker compose" "https://docs.docker.com/compose/install/"
  fi
  if docker buildx version >/dev/null 2>&1; then
    ok "docker buildx" "$(docker buildx version 2>/dev/null | awk '{print $2}')"
  else
    miss "docker buildx" "https://docs.docker.com/build/concepts/overview/#install-buildx"
  fi
}

cmd_check() {
  detect_platform
  info "Platform: $(platform_label)"
  check_supported_platform
  echo
  check_base_tools
  echo
  check_mise_tools
  echo
  check_docker
  echo
  if [ "$missing" -ne 0 ]; then
    printf "%sOne or more required tools are missing.%s Run 'make install-deps' (Docker must be installed separately).\n" "$C_ERR" "$C_RST"
    exit 1
  fi
  printf "%sAll required tools are installed.%s\n" "$C_OK" "$C_RST"
}

# ---------------------------------------------------------------------------
# install
# ---------------------------------------------------------------------------

sudo_cmd() {
  if [ "$(id -u)" -eq 0 ]; then "$@"; else sudo "$@"; fi
}

install_base_ubuntu() {
  local pkgs="" tool
  for tool in $BASE_TOOLS; do
    have "$tool" || pkgs="$pkgs $tool"
  done
  if [ -z "$pkgs" ]; then
    info "Base tools already installed ($BASE_TOOLS)"
    return
  fi
  info "Installing base packages via apt:${pkgs}"
  sudo_cmd apt-get update -y
  # shellcheck disable=SC2086
  sudo_cmd apt-get install -y --no-install-recommends ca-certificates $pkgs
}

install_base_macos() {
  if ! xcode-select -p >/dev/null 2>&1; then
    info "Installing Xcode Command Line Tools (provides git and make)..."
    xcode-select --install || true
    die "Finish the Xcode Command Line Tools installer dialog, then re-run 'make install-deps'."
  fi
  info "Xcode Command Line Tools present (git, make, curl)"
}

install_mise() {
  if [ -n "$(find_mise)" ]; then
    info "mise already installed: $(find_mise)"
    return
  fi
  if [ "$OS" = "Darwin" ] && have brew; then
    info "Installing mise via Homebrew..."
    brew install mise
  else
    info "Installing mise via the official installer (https://mise.run) into ~/.local/bin ..."
    curl -fsSL https://mise.run | sh
  fi
  [ -n "$(find_mise)" ] || die "mise installation failed; see https://mise.jdx.dev/getting-started.html"
}

install_mise_tools() {
  local mise_bin
  mise_bin="$(find_mise)"
  info "Installing the toolchain pinned in mise.toml..."
  (cd "$ROOT_DIR" && "$mise_bin" trust --yes "$MISE_TOML" && "$mise_bin" install --yes)
}

cmd_install() {
  detect_platform
  info "Platform: $(platform_label)"
  check_supported_platform

  case "$OS" in
    Linux)
      if have apt-get; then
        install_base_ubuntu
      else
        warn "apt-get not found; install manually if missing: $BASE_TOOLS"
      fi
      ;;
    Darwin) install_base_macos ;;
    *) warn "Unknown OS; install manually if missing: $BASE_TOOLS" ;;
  esac

  install_mise
  install_mise_tools

  echo
  info "Done. If 'go'/'helm' aren't found in new shells, activate mise once:"
  echo "    bash: echo 'eval \"\$(mise activate bash)\"' >> ~/.bashrc"
  echo "    zsh:  echo 'eval \"\$(mise activate zsh)\"'  >> ~/.zshrc"
  echo
  info "Docker / Docker Compose / Buildx are not installed by this script; see REQUIREMENTS.md."
  echo
  missing=0
  cmd_check
}

case "${1:-}" in
  check)   cmd_check ;;
  install) cmd_install ;;
  *)       echo "usage: $0 {check|install}" >&2; exit 2 ;;
esac
