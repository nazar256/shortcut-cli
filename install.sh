#!/bin/sh

set -eu

OWNER="nazar256"
REPO="shortcut-cli"
PROJECT_NAME="shortcut-cli"
BINARY_NAME="shortcut"
BASE_URL_OVERRIDE="${SHORTCUT_INSTALL_BASE_URL:-}"

usage() {
  cat <<'EOF'
Install shortcut-cli from the latest GitHub release.

Usage:
  install.sh [--version <tag>] [--install-dir <dir>]

Options:
  --version <tag>      Install a specific release tag (for example v1.0.0)
  --install-dir <dir>  Install into a specific directory
  -h, --help           Show this help message
EOF
}

log() {
  printf '%s\n' "$*" >&2
}

fail() {
  log "Error: $*"
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

detect_os() {
  case "$(uname -s)" in
    Linux) printf 'linux' ;;
    Darwin) printf 'darwin' ;;
    *) fail "unsupported operating system: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac
}

resolve_latest_tag() {
  header_file="$1"
  curl -fsSLI "https://github.com/${OWNER}/${REPO}/releases/latest" >"${header_file}"
  location_line=$(grep -i '^location:' "${header_file}" | tail -n 1 || true)
  [ -n "${location_line}" ] || fail "could not resolve latest release"
  tag=$(printf '%s' "${location_line}" | tr -d '\r' | sed -nE 's#.*releases/tag/([^[:space:]]+).*#\1#p')
  [ -n "${tag}" ] || fail "could not parse latest release tag"
  printf '%s' "${tag}"
}

is_writable_dir() {
  dir="$1"
  [ -d "${dir}" ] || return 1
  [ -w "${dir}" ] || return 1
}

dir_in_path() {
  dir="$1"
  case ":${PATH}:" in
    *:"${dir}":*) return 0 ;;
    *) return 1 ;;
  esac
}

canonicalize_dir() {
  dir="$1"
  [ -d "${dir}" ] || return 1
  (
    cd "${dir}" 2>/dev/null && pwd -P
  )
}

canonical_dirs_equal() {
  candidate="$1"
  other="$2"
  canonical_other=$(canonicalize_dir "${other}" || true)
  [ -n "${canonical_other}" ] || return 1
  [ "${candidate}" = "${canonical_other}" ]
}

canonical_dir_within() {
  candidate="$1"
  other="$2"
  canonical_other=$(canonicalize_dir "${other}" || true)
  [ -n "${canonical_other}" ] || return 1
  path_is_within "${candidate}" "${canonical_other}"
}

nvm_dir() {
  if [ -n "${NVM_DIR:-}" ]; then
    printf '%s' "${NVM_DIR}"
    return
  fi
  printf '%s' "${HOME}/.nvm"
}

npm_global_bin_dir() {
  if [ -n "${NPM_CONFIG_PREFIX:-}" ]; then
    printf '%s/bin' "${NPM_CONFIG_PREFIX}"
    return
  fi
  printf '%s' "${HOME}/.npm-global/bin"
}

yarn_global_bin_dir() {
  if [ -n "${YARN_GLOBAL_FOLDER:-}" ]; then
    printf '%s/bin' "${YARN_GLOBAL_FOLDER}"
    return
  fi
  printf '%s' "${HOME}/.yarn/bin"
}

pnpm_home_dir() {
  if [ -n "${PNPM_HOME:-}" ]; then
    printf '%s' "${PNPM_HOME}"
    return
  fi
  printf '%s' "${HOME}/.local/share/pnpm"
}

cargo_bin_dir() {
  if [ -n "${CARGO_HOME:-}" ]; then
    printf '%s/bin' "${CARGO_HOME}"
    return
  fi
  printf '%s' "${HOME}/.cargo/bin"
}

path_is_within() {
  path="$1"
  prefix="$2"
  case "${path}" in
    "${prefix}"|"${prefix}"/*) return 0 ;;
    *) return 1 ;;
  esac
}

is_language_managed_dir() {
  dir="$1"

  if canonical_dir_within "${dir}" "$(nvm_dir)"; then
    return 0
  fi

  case "${dir}" in
    */node_modules/.bin|*/node_modules/.bin/*) return 0 ;;
  esac

  if canonical_dirs_equal "${dir}" "$(npm_global_bin_dir)"; then
    return 0
  fi

  if canonical_dirs_equal "${dir}" "$(yarn_global_bin_dir)"; then
    return 0
  fi

  if canonical_dirs_equal "${dir}" "$(pnpm_home_dir)"; then
    return 0
  fi

  if canonical_dirs_equal "${dir}" "$(cargo_bin_dir)"; then
    return 0
  fi

  if [ -n "${GOBIN:-}" ] && canonical_dirs_equal "${dir}" "${GOBIN}"; then
    return 0
  fi

  old_ifs=${IFS}
  IFS=:
  for go_path in ${GOPATH:-${HOME}/go}; do
    [ -n "${go_path}" ] || continue
    if canonical_dirs_equal "${dir}" "${go_path}/bin"; then
      IFS=${old_ifs}
      return 0
    fi
  done
  IFS=${old_ifs}

  return 1
}

canonical_dir_in_path() {
  target="$1"

  old_ifs=${IFS}
  IFS=:
  for candidate in ${PATH}; do
    [ -n "${candidate}" ] || continue
    canonical_candidate=$(canonicalize_dir "${candidate}" || true)
    [ -n "${canonical_candidate}" ] || continue
    if [ "${canonical_candidate}" = "${target}" ]; then
      IFS=${old_ifs}
      return 0
    fi
  done
  IFS=${old_ifs}

  return 1
}

choose_install_dir() {
  if [ -n "${INSTALL_DIR}" ]; then
    mkdir -p "${INSTALL_DIR}"
    is_writable_dir "${INSTALL_DIR}" || fail "install dir is not writable: ${INSTALL_DIR}"
    printf '%s' "${INSTALL_DIR}"
    return
  fi

  old_ifs=${IFS}
  IFS=:
  for candidate in ${PATH}; do
    [ -n "${candidate}" ] || continue
    canonical_candidate=$(canonicalize_dir "${candidate}" || true)
    [ -n "${canonical_candidate}" ] || continue
    if is_language_managed_dir "${canonical_candidate}"; then
      continue
    fi
    if is_writable_dir "${canonical_candidate}"; then
      IFS=${old_ifs}
      printf '%s' "${canonical_candidate}"
      return
    fi
  done
  IFS=${old_ifs}

  for fallback in "${HOME}/.local/bin" "${HOME}/bin"; do
    mkdir -p "${fallback}"
    canonical_fallback=$(canonicalize_dir "${fallback}" || true)
    [ -n "${canonical_fallback}" ] || continue
    if is_writable_dir "${canonical_fallback}"; then
      printf '%s' "${canonical_fallback}"
      return
    fi
  done

  fail "could not find a writable install directory"
}

verify_checksum() {
  archive_path="$1"
  checksum_file="$2"
  archive_name=$(basename "${archive_path}")

  if command -v sha256sum >/dev/null 2>&1; then
    expected=$(grep "  ${archive_name}" "${checksum_file}" | awk '{print $1}')
    actual=$(sha256sum "${archive_path}" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    expected=$(grep "  ${archive_name}" "${checksum_file}" | awk '{print $1}')
    actual=$(shasum -a 256 "${archive_path}" | awk '{print $1}')
  else
    fail "need sha256sum or shasum to verify release checksum"
  fi

  [ -n "${expected}" ] || fail "checksum entry not found for ${archive_name}"
  [ "${expected}" = "${actual}" ] || fail "checksum verification failed for ${archive_name}"
}

VERSION=""
INSTALL_DIR=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || fail "missing value for --version"
      VERSION="$2"
      shift 2
      ;;
    --install-dir)
      [ "$#" -ge 2 ] || fail "missing value for --install-dir"
      INSTALL_DIR="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

require_cmd curl
require_cmd tar
require_cmd mktemp
require_cmd install

tmpdir=$(mktemp -d)
trap 'rm -rf "${tmpdir}"' EXIT INT TERM HUP

if [ -z "${VERSION}" ]; then
  VERSION=$(resolve_latest_tag "${tmpdir}/latest.headers")
fi

OS=$(detect_os)
ARCH=$(detect_arch)
ARCHIVE_NAME="${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
CHECKSUM_NAME="${PROJECT_NAME}_${VERSION}_checksums.txt"
BASE_URL="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}"
if [ -n "${BASE_URL_OVERRIDE}" ]; then
  BASE_URL="${BASE_URL_OVERRIDE}"
fi
ARCHIVE_PATH="${tmpdir}/${ARCHIVE_NAME}"
CHECKSUM_PATH="${tmpdir}/${CHECKSUM_NAME}"
TARGET_DIR=$(choose_install_dir)

log "Installing ${BINARY_NAME} ${VERSION} for ${OS}/${ARCH}"
log "Downloading release assets from ${BASE_URL}"
curl -fsSL "${BASE_URL}/${ARCHIVE_NAME}" -o "${ARCHIVE_PATH}"
curl -fsSL "${BASE_URL}/${CHECKSUM_NAME}" -o "${CHECKSUM_PATH}"

verify_checksum "${ARCHIVE_PATH}" "${CHECKSUM_PATH}"

tar -xzf "${ARCHIVE_PATH}" -C "${tmpdir}"
[ -f "${tmpdir}/${BINARY_NAME}" ] || fail "archive did not contain ${BINARY_NAME}"

install -m 0755 "${tmpdir}/${BINARY_NAME}" "${TARGET_DIR}/${BINARY_NAME}"

log "Installed to ${TARGET_DIR}/${BINARY_NAME}"
if canonical_dir_in_path "${TARGET_DIR}" || dir_in_path "${TARGET_DIR}"; then
  "${TARGET_DIR}/${BINARY_NAME}" version
else
  log "${TARGET_DIR} is not currently in PATH"
  log "Add it with: export PATH=\"${TARGET_DIR}:\$PATH\""
  "${TARGET_DIR}/${BINARY_NAME}" version
fi
