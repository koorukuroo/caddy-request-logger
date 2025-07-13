#!/bin/bash

# Caddy Request Logger 빌드 스크립트

echo "🚀 Caddy Request Logger 빌드 시작..."

# 의존성 다운로드
echo "📦 의존성 다운로드 중..."
go mod tidy

# Caddy 빌드
echo "🔨 Caddy 빌드 중..."
go build -o caddy ./cmd/caddy

# 빌드 확인
if [ -f "./caddy" ]; then
    echo "✅ 빌드 성공!"
    echo "📝 모듈 목록:"
    ./caddy list-modules | grep request_logger
    echo ""
    echo "🎯 사용 방법:"
    echo "  ./caddy run --config Caddyfile --adapter caddyfile"
    echo ""
    echo "🔍 디버그 모드:"
    echo "  ./caddy run --config Caddyfile --adapter caddyfile --verbose"
else
    echo "❌ 빌드 실패"
    exit 1
fi 