# Caddy Request Logger

Caddy의 HTTP 요청 헤더와 body를 자세히 로깅하는 미들웨어입니다.

## 주요 기능

-   **Request Body 로깅**: 요청 본문 내용을 로그에 포함
-   **Request Headers 로깅**: 모든 헤더 또는 특정 헤더만 로깅
-   **유연한 필터링**: 특정 경로, HTTP 메서드, Content-Type 제외 가능
-   **크기 제한**: 로깅할 최대 body 크기 설정
-   **Base64 인코딩**: 바이너리 데이터 안전하게 로깅
-   **보안**: 민감한 헤더 제외 가능

## 설치

1. 이 저장소를 클론합니다:

```bash
git clone https://github.com/koorukuroo/caddy-request-logger.git
cd caddy-request-logger
```

2. 의존성을 설치합니다:

```bash
go mod tidy
```

3. Caddy를 빌드합니다:

```bash
go build -o caddy ./cmd/caddy
```

## 사용법

### 기본 설정

```caddy
:8080 {
    request_logger {
        include_request_body
        include_all_headers
    }

    file_server
}
```

### 고급 설정

```caddy
:8080 {
    request_logger {
        log_level debug
        include_request_body
        include_all_headers
        max_body_size 1MB
        base64_encode_body
        exclude_headers Authorization X-API-Key
        skip_paths /health /metrics
        skip_methods OPTIONS HEAD
        skip_content_types image/ video/ application/octet-stream
    }

    reverse_proxy localhost:3000
}
```

## 설정 옵션

| 옵션                   | 타입     | 기본값  | 설명                                        |
| ---------------------- | -------- | ------- | ------------------------------------------- |
| `log_level`            | string   | `info`  | 로그 레벨 (debug, info, warn, error)        |
| `include_request_body` | bool     | `false` | 요청 본문을 로그에 포함                     |
| `include_all_headers`  | bool     | `false` | 모든 헤더를 로그에 포함                     |
| `max_body_size`        | string   | `1MB`   | 로깅할 최대 본문 크기 (예: 1MB, 512KB, 2GB) |
| `base64_encode_body`   | bool     | `false` | 요청 본문을 Base64로 인코딩                 |
| `include_headers`      | []string | `[]`    | 포함할 특정 헤더 목록                       |
| `exclude_headers`      | []string | `[]`    | 제외할 헤더 목록                            |
| `skip_paths`           | []string | `[]`    | 로깅하지 않을 경로 목록                     |
| `skip_methods`         | []string | `[]`    | 로깅하지 않을 HTTP 메서드 목록              |
| `skip_content_types`   | []string | `[]`    | 로깅하지 않을 Content-Type 목록             |

## 로그 출력 예시

```json
{
    "level": "info",
    "ts": "2024-01-15T10:30:00.000Z",
    "logger": "request_logger",
    "msg": "Request: POST /api/users",
    "method": "POST",
    "path": "/api/users",
    "query": "page=1&limit=10",
    "remote_addr": "192.168.1.100:54321",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "host": "example.com",
    "proto": "HTTP/1.1",
    "content_type": "application/json",
    "content_length": 156,
    "timestamp": "2024-01-15T10:30:00.000Z",
    "headers": {
        "Content-Type": ["application/json"],
        "Authorization": ["Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."],
        "X-Request-ID": ["req-123456"]
    },
    "request_body": "{\"name\":\"John Doe\",\"email\":\"john@example.com\"}"
}
```

## 보안 고려사항

-   민감한 헤더는 `exclude_headers`로 제외하세요:

    ```caddy
    exclude_headers Authorization X-API-Key Cookie
    ```

-   개인정보가 포함된 경로는 `skip_paths`로 제외하세요:

    ```caddy
    skip_paths /login /register /reset-password
    ```

-   큰 파일 업로드는 `skip_content_types`로 제외하세요:
    ```caddy
    skip_content_types multipart/form-data application/octet-stream
    ```

## 성능 최적화

-   `max_body_size`를 적절히 설정하여 메모리 사용량을 제한하세요
-   불필요한 경로는 `skip_paths`로 제외하세요
-   정적 파일은 `skip_content_types`로 제외하세요

## 문제 해결

### 로그가 출력되지 않는 경우

1. Caddy 로그 레벨 확인:

    ```bash
    ./caddy run --config Caddyfile --adapter caddyfile --verbose
    ```

2. 모듈이 올바르게 로드되었는지 확인:
    ```bash
    ./caddy list-modules | grep request_logger
    ```

### 메모리 사용량이 높은 경우

-   `max_body_size`를 줄이세요
-   `skip_paths`와 `skip_content_types`를 활용하세요
-   큰 파일 업로드 경로는 제외하세요

## 기여

버그 리포트나 기능 요청은 GitHub Issues를 통해 제출해주세요.

## 라이센스

MIT License
