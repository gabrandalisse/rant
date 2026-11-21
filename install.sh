#!/bin/sh
set -e

REPO="gabrandalisse/rant"
BINARY="rant"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Arquitectura no soportada: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Sistema operativo no soportado: $OS" >&2; exit 1 ;;
esac

URL="https://github.com/${REPO}/releases/latest/download/${BINARY}_${OS}_${ARCH}.tar.gz"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Descargando ${BINARY} para ${OS}/${ARCH}..."
curl -fsSL "$URL" -o "${TMP_DIR}/archive.tar.gz"
tar -xzf "${TMP_DIR}/archive.tar.gz" -C "$TMP_DIR"

INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"
install -m 755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "Instalado en ${INSTALL_DIR}/${BINARY}"
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "Agregá ${INSTALL_DIR} a tu PATH para ejecutar '${BINARY}' desde cualquier lugar." ;;
esac