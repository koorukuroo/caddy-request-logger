#!/bin/bash

# Request Logger 테스트 스크립트

echo "🧪 Request Logger 테스트 시작..."

# 기본 GET 요청
echo "📋 1. 기본 GET 요청 테스트"
curl -v http://localhost:8080/ \
    -H "User-Agent: Test-Agent/1.0" \
    -H "X-Request-ID: test-123"

echo -e "\n\n"

# POST 요청 with JSON body
echo "📋 2. POST 요청 with JSON body 테스트"
curl -v -X POST http://localhost:8080/api/users \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer test-token" \
    -H "X-API-Key: secret-key" \
    -d '{"name":"홍길동","email":"hong@example.com","age":30}'

echo -e "\n\n"

# PUT 요청 with form data
echo "📋 3. PUT 요청 with form data 테스트"
curl -v -X PUT http://localhost:8080/api/profile \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Cookie: session=abc123" \
    -d "name=김철수&email=kim@example.com&age=25"

echo -e "\n\n"

# 큰 JSON 요청
echo "📋 4. 큰 JSON 요청 테스트"
curl -v -X POST http://localhost:8080/api/data \
    -H "Content-Type: application/json" \
    -d '{
        "data": {
            "users": [
                {"id": 1, "name": "사용자1", "email": "user1@example.com"},
                {"id": 2, "name": "사용자2", "email": "user2@example.com"},
                {"id": 3, "name": "사용자3", "email": "user3@example.com"}
            ],
            "metadata": {
                "total": 3,
                "page": 1,
                "limit": 10,
                "timestamp": "2024-01-15T10:30:00Z"
            }
        }
    }'

echo -e "\n\n"

# 건강 체크 요청 (skip_paths에 포함됨)
echo "📋 5. 건강 체크 요청 테스트 (로그 생략됨)"
curl -v http://localhost:8080/health

echo -e "\n\n"

# OPTIONS 요청 (skip_methods에 포함됨)
echo "📋 6. OPTIONS 요청 테스트 (로그 생략됨)"
curl -v -X OPTIONS http://localhost:8080/api/users \
    -H "Access-Control-Request-Method: POST" \
    -H "Access-Control-Request-Headers: Content-Type"

echo -e "\n\n"

echo "✅ 테스트 완료! Caddy 로그를 확인해보세요."
echo "💡 로그 확인 명령어:"
echo "  tail -f caddy.log" 