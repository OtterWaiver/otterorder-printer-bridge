#!/usr/bin/env bash
set -euo pipefail

# ------------- Config -------------
APP_NAME="Otter Order Printer Bridge"   # Your .app display name
BIN_NAME="otter-order-printer-bridge"   # Your CLI/binary name (no spaces)
VERSION="${1:-}"                        # Optional version suffix like v1.0.7
# -----------------------------------

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build/bin"
OUT_DIR="$ROOT_DIR/release"

bold()   { printf "\033[1m%s\033[0m\n" "$*"; }
info()   { printf "➜ %s\n" "$*"; }
warn()   { printf "⚠️  %s\n" "$*"; }
die()    { printf "❌ %s\n" "$*" >&2; exit 1; }

require_cmd() { command -v "$1" >/dev/null 2>&1 || die "Missing tool: $1"; }

# Tools we expect
require_cmd go
require_cmd wails

# On macOS we use ditto to zip .app bundles
if [[ "$(uname)" == "Darwin" ]]; then
  require_cmd ditto
fi

# Make output folder fresh
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

# Optional clean
rm -rf "$ROOT_DIR/build"
mkdir -p "$BUILD_DIR"

SUFFIX() {
  # Adds optional -<version> suffix to filenames
  if [[ -n "$VERSION" ]]; then
    echo "-$1-$VERSION"
  else
    echo "-$1"
  fi
}

# ---------- Helpers ----------

pkg_macos () {
  local arch="$1"   # macos-arm64 | macos-x86_64
  local root="$ROOT_DIR/build/bin"

  # Search recursively for any .app produced by Wails
  local bundle
  bundle="$(find "$root" -type d -name '*.app' -print | head -n 1)"

  if [[ -z "$bundle" ]]; then
    warn "No .app found for $arch under $root"
    echo "Contents of $root:"
    find "$root" -maxdepth 3 -print
    die "No .app found for $arch"
  fi

  # Ensure inner binary is executable
  chmod +x "$bundle/Contents/MacOS/"* || true

  local safe_name="OtterOrderPrinterBridge"
  local zip_name="$OUT_DIR/${safe_name}$(SUFFIX "$arch").app.zip"
  info "Zipping $bundle -> $zip_name"
  ditto -c -k --keepParent "$bundle" "$zip_name"

  # Optional: pick up a .pkg if Wails generated one
  local pkg
  pkg="$(find "$root" -maxdepth 3 -type f -name '*.pkg' -print | head -n 1 || true)"
  if [[ -n "$pkg" ]]; then
    cp "$pkg" "$OUT_DIR/${safe_name}$(SUFFIX "$arch").pkg"
  fi
}
pkg_windows () {
  local arch="windows-amd64"
  local exe
  exe="$(find "$BUILD_DIR" -type f -name '*.exe' | head -n 1)"
  [[ -n "$exe" ]] || die "No .exe found for $arch"
  local safe_name="OtterOrderPrinterBridge"
  local zip_name="$OUT_DIR/${safe_name}$(SUFFIX "$arch").zip"
  info "Zipping $exe -> $zip_name"
  (cd "$(dirname "$exe")" && zip -q -9 "$zip_name" "$(basename "$exe")")
}

pkg_linux () {
  local arch="linux-amd64"
  local bin
  # Find first executable file (ELF) in build/bin
  bin="$(find "$BUILD_DIR" -maxdepth 1 -type f -perm -u+x -print | head -n 1)"
  [[ -n "$bin" ]] || die "No executable found for $arch"
  # Normalize to predictable name inside archive
  cp "$bin" "$BUILD_DIR/$BIN_NAME"
  local safe_name="OtterOrderPrinterBridge"
  local tgz_name="$OUT_DIR/${safe_name}$(SUFFIX "$arch").tar.gz"
  info "Tarring $BUILD_DIR/$BIN_NAME -> $tgz_name"
  tar -C "$BUILD_DIR" -czf "$tgz_name" "$BIN_NAME"
}

# ---------- Builds ----------

build_platform () {
  local plat="$1"
  info "Building Wails for: $plat"
  # CGO is required by Wails; allow wails to handle it
  # Add --clean to ensure fresh build each time
  wails build -clean -platform "$plat" \
    -ldflags "-s -w" \
    -o "$BIN_NAME" \
    -trimpath
}

# macOS arm64
build_macos_arm64 () {
  rm -rf "$BUILD_DIR"
  mkdir -p "$BUILD_DIR"
  build_platform "darwin/arm64"
  pkg_macos "macos-arm64"
}

# macOS x86_64
build_macos_amd64 () {
  rm -rf "$BUILD_DIR"
  mkdir -p "$BUILD_DIR"
  build_platform "darwin/amd64"
  pkg_macos "macos-x86_64"
}

# Windows amd64
build_windows_amd64 () {
  rm -rf "$BUILD_DIR"
  mkdir -p "$BUILD_DIR"
  build_platform "windows/amd64"
  pkg_windows
}

# Linux amd64
build_linux_amd64 () {
  rm -rf "$BUILD_DIR"
  mkdir -p "$BUILD_DIR"
  build_platform "linux/amd64"
  pkg_linux
}

# ---------- Run ----------

bold "Building ${APP_NAME} ${VERSION:+($VERSION)} for macOS (arm64)…"
build_macos_arm64

bold "Building ${APP_NAME} ${VERSION:+($VERSION)} for macOS (x86_64)…"
build_macos_amd64

bold "Building ${APP_NAME} ${VERSION:+($VERSION)} for Windows (amd64)…"
build_windows_amd64

bold "Building ${APP_NAME} ${VERSION:+($VERSION)} for Linux (amd64)…"
build_linux_amd64

bold "All artifacts are in: $OUT_DIR"
ls -lh "$OUT_DIR"
echo
info "Upload these files to your GitHub Release."