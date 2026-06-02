#!/bin/sh
set -e

REPO="belousovsergey56/tn"
BINARY="tn"

# 1. Определяем операционную систему
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  darwin)  OS="darwin" ;;
  linux)   OS="linux" ;;
  msys*|mingw*|cygwin*) OS="windows" ;;
  *) echo "Error: Unsupported OS: ${OS}"; exit 1 ;;
esac

# 2. Определяем архитектуру процессора
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Error: Unsupported architecture: ${ARCH}"; exit 1 ;;
esac

# 3. Получаем тег самого свежего релиза через GitHub API
echo "Fetching the latest release version..."
TAG=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG" ]; then
  echo "Error: Could not fetch the latest release tag. Please check your internet connection or repository name."
  exit 1
fi

# 4. Определяем расширение файла (в .goreleaser.yaml мы настроили zip для Windows)
EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

# 5. Формируем URL для скачивания (шаблон совпадает с тем, что генерирует GoReleaser)
URL="https://github.com/${REPO}/releases/download/${TAG}/${BINARY}_${OS}_${ARCH}.${EXT}"

# 6. Скачиваем архив
echo "Downloading ${BINARY} ${TAG} for ${OS}/${ARCH}..."
curl -sSfL "$URL" -o "${BINARY}.${EXT}"

# 7. Распаковываем бинарник
if [ "$EXT" = "zip" ]; then
  unzip -q -o "${BINARY}.${EXT}" "${BINARY}.exe"
  rm "${BINARY}.${EXT}"
else
  tar -xzf "${BINARY}.${EXT}" "${BINARY}"
  rm "${BINARY}.${EXT}"
fi

chmod +x "${BINARY}"

# 8. Складываем бинарник в локальную папку ./bin (это безопасно и не требует прав sudo)
INSTALL_DIR="./bin"
mkdir -p "$INSTALL_DIR"
mv "${BINARY}" "$INSTALL_DIR/"

echo ""
echo "🎉 ${BINARY} successfully installed to ${INSTALL_DIR}/${BINARY}"
echo "To use it anywhere, move it to your PATH, for example:"
echo "  sudo mv ${INSTALL_DIR}/${BINARY} /usr/local/bin/"
