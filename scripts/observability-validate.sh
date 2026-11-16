#!/usr/bin/env bash
#
# Validates the local observability stack's configuration without starting it:
#
#   - docker-compose.yml + docker-compose.observability.yml merge cleanly
#   - Alloy, Prometheus, Loki, and Tempo accept their config files (each is
#     checked by the tool itself, using the exact image version pinned in
#     docker-compose.observability.yml)
#   - every Grafana dashboard is valid JSON with a unique uid
#
# Needs Docker. Must stay compatible with macOS's stock bash 3.2.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OBS_DIR="${ROOT_DIR}/observability"
COMPOSE_FILES=(-f "${ROOT_DIR}/docker-compose.yml" -f "${ROOT_DIR}/docker-compose.observability.yml")

if [ -t 1 ]; then
  C_OK=$'\033[32m'; C_ERR=$'\033[31m'; C_RST=$'\033[0m'
else
  C_OK=""; C_ERR=""; C_RST=""
fi

failures=0

# run_check <description> <command...>: runs the command quietly, printing its
# output only on failure.
run_check() {
  local desc="$1"; shift
  local out
  if out="$("$@" 2>&1)"; then
    printf '  %sok%s    %s\n' "$C_OK" "$C_RST" "$desc"
  else
    printf '  %sFAIL%s  %s\n' "$C_ERR" "$C_RST" "$desc"
    printf '%s\n' "$out" | sed 's/^/        /'
    failures=$((failures + 1))
  fi
}

# image_of <service>: the image pinned for a Compose service in
# docker-compose.observability.yml (the first "image:" line under it).
image_of() {
  awk -v svc="  $1:" '$0 == svc { found = 1; next } found && /^    image:/ { print $2; exit }' \
    "${ROOT_DIR}/docker-compose.observability.yml"
}

command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 1; }

echo "Validating the observability stack configuration..."

run_check "docker compose files merge" \
  docker compose "${COMPOSE_FILES[@]}" config --quiet

run_check "Alloy config (observability/alloy/config.alloy)" \
  docker run --rm -v "${OBS_DIR}/alloy:/cfg:ro" "$(image_of alloy)" fmt /cfg/config.alloy

run_check "Prometheus config (observability/prometheus/prometheus.yml)" \
  docker run --rm --entrypoint promtool -v "${OBS_DIR}/prometheus:/cfg:ro" "$(image_of prometheus)" \
  check config /cfg/prometheus.yml

run_check "Loki config (observability/loki/loki.yml)" \
  docker run --rm -v "${OBS_DIR}/loki:/cfg:ro" "$(image_of loki)" \
  -config.file=/cfg/loki.yml -verify-config

run_check "Tempo config (observability/tempo/tempo.yml)" \
  docker run --rm -v "${OBS_DIR}/tempo:/cfg:ro" "$(image_of tempo)" \
  -config.file=/cfg/tempo.yml -config.verify=true

check_dashboards() {
  local uids="" file uid
  for file in "${OBS_DIR}"/grafana/dashboards/*.json; do
    uid="$(python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); print(d["uid"])' "$file")" || return 1
    case " ${uids} " in
      *" ${uid} "*) echo "duplicate dashboard uid '${uid}' in ${file}"; return 1 ;;
    esac
    uids="${uids} ${uid}"
  done
}
run_check "Grafana dashboards are valid JSON with unique uids" check_dashboards

if [ "$failures" -gt 0 ]; then
  echo "${failures} check(s) failed."
  exit 1
fi
echo "All observability configuration checks passed."
