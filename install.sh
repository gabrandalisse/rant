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

HOOK_LINE='command -v rant >/dev/null 2>&1 && eval "$(rant shell-init)"'

configure_rc() {
  RC_FILE="$1"
  if [ -f "$RC_FILE" ]; then
    if ! grep -Fq "rant shell-init" "$RC_FILE"; then
      printf '\n# Evitar guardar comandos rant en el historial\n%s\n' "$HOOK_LINE" >> "$RC_FILE"
      echo "Protección de historial agregada en ${RC_FILE}"
    fi
  fi
}

configure_rc "${HOME}/.zshrc"
configure_rc "${HOME}/.bashrc"

echo "Instalado en ${INSTALL_DIR}/${BINARY}"
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "Agregá ${INSTALL_DIR} a tu PATH para ejecutar '${BINARY}' desde cualquier lugar." ;;
esac