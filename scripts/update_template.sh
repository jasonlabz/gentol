#!/bin/bash
# update_template.sh - 更新嵌入式模板
# 用法: bash scripts/update_template.sh [template_repo_url]
#
# 从模板仓库 clone，打包为 template.tar.gz，
# 放到 embedded/ 目录供 //go:embed 编译进 gentol 二进制

set -e

REPO_URL="${1:-https://github.com/jasonlabz/generate-example-project.git}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT="$PROJECT_DIR/embedded/template.tar.gz"
TMP_DIR="$(mktemp -d)"

echo "Cloning template from: $REPO_URL"
git clone --depth 1 "$REPO_URL" "$TMP_DIR/template"

echo "Creating template.tar.gz..."
cd "$TMP_DIR/template"

# 用 tar 打包，排除 .git 和 go.sum
tar czf "$OUTPUT" \
  --exclude='.git' \
  --exclude='go.sum' \
  .

echo "Template archive created: $OUTPUT"
echo "File count: $(tar tzf "$OUTPUT" | grep -v '/$' | wc -l)"
echo ""
echo "Now rebuild gentol to embed the updated template:"
echo "  cd $PROJECT_DIR && go build ."

# 清理
rm -rf "$TMP_DIR"
