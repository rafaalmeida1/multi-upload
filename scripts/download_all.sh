#!/usr/bin/env bash
# download_all.sh — Baixa TODO o conteúdo desta API (manifest + arquivos)
# preservando a estrutura `YYYY/MM/DD/uuid.ext`. Idempotente: pula arquivos
# já baixados que tenham o mesmo tamanho.
#
# Uso:
#   API_URL=https://api.exemplo.com EXPORT_TOKEN=xxxx ./scripts/download_all.sh ./export
#
# Variáveis aceitas:
#   API_URL       URL base da API (default: http://localhost:8082)
#   EXPORT_TOKEN  Token configurado em EXPORT_TOKEN do servidor (opcional)
#   CONCURRENCY   Quantidade de downloads paralelos (default: 6)

set -euo pipefail

API_URL="${API_URL:-http://localhost:8082}"
TOKEN_HEADER=()
if [[ -n "${EXPORT_TOKEN:-}" ]]; then
  TOKEN_HEADER=(-H "X-Export-Token: ${EXPORT_TOKEN}")
fi
OUT_DIR="${1:-./export}"
CONCURRENCY="${CONCURRENCY:-6}"

mkdir -p "$OUT_DIR"
MANIFEST_PATH="$OUT_DIR/manifest.json"

command -v jq >/dev/null 2>&1 || {
  echo "ERRO: jq não está instalado. Instale com 'apt install jq' ou similar." >&2
  exit 1
}
command -v curl >/dev/null 2>&1 || {
  echo "ERRO: curl não está instalado." >&2
  exit 1
}

echo "==> Baixando manifest de $API_URL/api/v1/export/manifest"
curl -fsSL "${TOKEN_HEADER[@]}" "$API_URL/api/v1/export/manifest" -o "$MANIFEST_PATH"

TOTAL=$(jq '.total' "$MANIFEST_PATH")
echo "==> $TOTAL arquivos no manifest"

if [[ "$TOTAL" -eq 0 ]]; then
  echo "Nada a baixar. Saindo."
  exit 0
fi

# Lê stats opcionalmente.
if curl -fsSL "${TOKEN_HEADER[@]}" "$API_URL/api/v1/export/stats" -o "$OUT_DIR/stats.json" 2>/dev/null; then
  echo "==> Stats: $(jq -c '{files: .total_files, size: .total_human, by_type: .by_type}' "$OUT_DIR/stats.json")"
fi

# Gera lista "<file_size>\t<file_path>" para baixar em paralelo via xargs.
TMP_LIST="$(mktemp)"
trap 'rm -f "$TMP_LIST"' EXIT
jq -r '.items[] | "\(.file_size)\t\(.file_path)"' "$MANIFEST_PATH" > "$TMP_LIST"

download_one() {
  local size="$1"
  local rel="$2"
  local out="$OUT_DIR/files/$rel"
  mkdir -p "$(dirname "$out")"

  if [[ -f "$out" ]]; then
    local existing
    existing=$(stat -c %s "$out" 2>/dev/null || stat -f %z "$out" 2>/dev/null || echo 0)
    if [[ "$existing" == "$size" ]]; then
      echo "skip $rel ($size bytes)"
      return 0
    fi
  fi

  echo "get  $rel"
  # Usa --range opcional? Mais simples baixar inteiro. -L segue redirects.
  curl -fsSL ${TOKEN_HEADER[@]+"${TOKEN_HEADER[@]}"} \
    "$API_URL/api/v1/files/$rel" -o "$out.part" \
    && mv "$out.part" "$out"
}
export -f download_one
export API_URL OUT_DIR
export EXPORT_TOKEN="${EXPORT_TOKEN:-}"
# Reexporta header como string parseável:
export TOKEN_HEADER_STR="${EXPORT_TOKEN:+-H X-Export-Token: $EXPORT_TOKEN}"

# xargs paralelo (linhas TAB-separadas).
awk -F'\t' '{print $1"\t"$2}' "$TMP_LIST" | \
  xargs -P "$CONCURRENCY" -I{} bash -c '
    IFS=$'"'"'\t'"'"' read -r size rel <<<"{}"
    download_one "$size" "$rel"
  '

echo ""
echo "==> Concluído. Saída em: $OUT_DIR"
echo "    - manifest.json  ($TOTAL itens)"
echo "    - files/         (estrutura YYYY/MM/DD/uuid.ext)"
echo ""
echo "Para migrar para o novo sistema (Cloudflare Worker):"
echo "  cd ../app_multi_upload_cf"
echo "  node scripts/migrate.mjs --from-dir $OUT_DIR \\"
echo "      --to https://<seu-worker>.workers.dev \\"
echo "      --user <username> --pass <senha>"
