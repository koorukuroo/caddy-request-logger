package request_logger

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(RequestLogger{})
	httpcaddyfile.RegisterHandlerDirective("request_logger", parseCaddyfile)
}

// parseSize parses a size string (e.g., "1MB", "512KB", "2GB") and returns the size in bytes
func parseSize(sizeStr string) (int, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Convert to uppercase for case-insensitive matching
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Handle just numeric values (assume bytes)
	if val, err := strconv.Atoi(sizeStr); err == nil {
		return val, nil
	}

	// Extract number and unit
	var num string
	var unit string
	
	for i, char := range sizeStr {
		if char >= '0' && char <= '9' || char == '.' {
			num += string(char)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if num == "" {
		return 0, fmt.Errorf("no numeric value found in size string: %s", sizeStr)
	}

	// Parse the numeric part
	val, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", num)
	}

	// Convert based on unit
	switch unit {
	case "B", "":
		return int(val), nil
	case "KB":
		return int(val * 1024), nil
	case "MB":
		return int(val * 1024 * 1024), nil
	case "GB":
		return int(val * 1024 * 1024 * 1024), nil
	case "TB":
		return int(val * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// parseCaddyfile parses the Caddyfile configuration for request_logger
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var rl RequestLogger
	
	// Parse the Caddyfile configuration
	err := rl.UnmarshalCaddyfile(h.Dispenser)
	if err != nil {
		return nil, err
	}
	
	return &rl, nil
}

// RequestLogger implements an HTTP middleware that logs request details
type RequestLogger struct {
	// Logger name for structured logging
	LoggerName string `json:"logger_name,omitempty"`
	
	// Log level: debug, info, warn, error
	LogLevel string `json:"log_level,omitempty"`
	
	// Include request body in logs
	IncludeRequestBody bool `json:"include_request_body,omitempty"`
	
	// Include all request headers in logs
	IncludeAllHeaders bool `json:"include_all_headers,omitempty"`
	
	// Maximum body size to log (in bytes)
	MaxBodySize int `json:"max_body_size,omitempty"`
	
	// Skip logging for specific methods
	SkipMethods []string `json:"skip_methods,omitempty"`
	
	// Skip logging for specific paths
	SkipPaths []string `json:"skip_paths,omitempty"`
	
	// Specific headers to include in logs (if not include_all_headers)
	IncludeHeaders []string `json:"include_headers,omitempty"`
	
	// Headers to exclude from logging (when include_all_headers is true)
	ExcludeHeaders []string `json:"exclude_headers,omitempty"`
	
	// Skip logging for specific content types
	SkipContentTypes []string `json:"skip_content_types,omitempty"`
	
	// Base64 encode request body (useful for binary data)
	Base64EncodeBody bool `json:"base64_encode_body,omitempty"`
	
	logger *zap.Logger
}

// CaddyModule returns the module information.
func (RequestLogger) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.request_logger",
		New: func() caddy.Module { return new(RequestLogger) },
	}
}

// Provision sets up the module
func (rl *RequestLogger) Provision(ctx caddy.Context) error {
	// Set defaults
	if rl.LoggerName == "" {
		rl.LoggerName = "request_logger"
	}
	if rl.LogLevel == "" {
		rl.LogLevel = "info"
	}
	if rl.MaxBodySize == 0 {
		rl.MaxBodySize = 1024 * 1024 // 1MB default
	}
	
	// Get logger
	rl.logger = ctx.Logger(rl)
	
	return nil
}

// shouldSkipMethod checks if the request method should be skipped
func (rl *RequestLogger) shouldSkipMethod(method string) bool {
	for _, skipMethod := range rl.SkipMethods {
		if strings.EqualFold(method, skipMethod) {
			return true
		}
	}
	return false
}

// shouldSkipPath checks if the request path should be skipped
func (rl *RequestLogger) shouldSkipPath(path string) bool {
	for _, skipPath := range rl.SkipPaths {
		if strings.Contains(path, skipPath) {
			return true
		}
	}
	return false
}

// shouldSkipContentType checks if the request content type should be skipped
func (rl *RequestLogger) shouldSkipContentType(contentType string) bool {
	for _, skipType := range rl.SkipContentTypes {
		if strings.Contains(strings.ToLower(contentType), strings.ToLower(skipType)) {
			return true
		}
	}
	return false
}

// isHeaderExcluded checks if a header should be excluded from logging
func (rl *RequestLogger) isHeaderExcluded(headerName string) bool {
	for _, excludeHeader := range rl.ExcludeHeaders {
		if strings.EqualFold(headerName, excludeHeader) {
			return true
		}
	}
	return false
}

// ServeHTTP implements the middleware interface
func (rl *RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Check if we should skip logging for this method
	if rl.shouldSkipMethod(r.Method) {
		return next.ServeHTTP(w, r)
	}
	
	// Check if we should skip logging for this path
	if rl.shouldSkipPath(r.URL.Path) {
		return next.ServeHTTP(w, r)
	}
	
	// Check if we should skip logging for this content type
	contentType := r.Header.Get("Content-Type")
	if rl.shouldSkipContentType(contentType) {
		return next.ServeHTTP(w, r)
	}
	
	start := time.Now()
	
	// Read request body if needed
	var requestBody []byte
	if rl.IncludeRequestBody && r.Body != nil {
		requestBody, _ = io.ReadAll(io.LimitReader(r.Body, int64(rl.MaxBodySize)))
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	
	// Prepare log fields
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("query", r.URL.RawQuery),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("referer", r.Referer()),
		zap.String("host", r.Host),
		zap.String("proto", r.Proto),
		zap.String("content_type", contentType),
		zap.Int64("content_length", r.ContentLength),
		zap.Time("timestamp", start),
	}
	
	// Add request headers
	if rl.IncludeAllHeaders {
		headers := make(map[string][]string)
		for name, values := range r.Header {
			if !rl.isHeaderExcluded(name) {
				headers[name] = values
			}
		}
		if len(headers) > 0 {
			fields = append(fields, zap.Any("headers", headers))
		}
	} else if len(rl.IncludeHeaders) > 0 {
		headers := make(map[string]string)
		for _, headerName := range rl.IncludeHeaders {
			if value := r.Header.Get(headerName); value != "" {
				headers[headerName] = value
			}
		}
		if len(headers) > 0 {
			fields = append(fields, zap.Any("headers", headers))
		}
	}
	
	// Add request body if included
	if rl.IncludeRequestBody && len(requestBody) > 0 {
		if rl.Base64EncodeBody {
			encoded := base64.StdEncoding.EncodeToString(requestBody)
			fields = append(fields, zap.String("request_body_b64", encoded))
		} else {
			fields = append(fields, zap.ByteString("request_body", requestBody))
		}
	}
	
	// Log the request
	message := fmt.Sprintf("Request: %s %s", r.Method, r.URL.Path)
	
	switch rl.LogLevel {
	case "debug":
		rl.logger.Debug(message, fields...)
	case "info":
		rl.logger.Info(message, fields...)
	case "warn":
		rl.logger.Warn(message, fields...)
	case "error":
		rl.logger.Error(message, fields...)
	default:
		rl.logger.Info(message, fields...)
	}
	
	// Call next handler
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (rl *RequestLogger) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "logger_name":
				if !d.Args(&rl.LoggerName) {
					return d.ArgErr()
				}
			case "log_level":
				if !d.Args(&rl.LogLevel) {
					return d.ArgErr()
				}
			case "include_request_body":
				rl.IncludeRequestBody = true
			case "include_all_headers":
				rl.IncludeAllHeaders = true
			case "base64_encode_body":
				rl.Base64EncodeBody = true
			case "max_body_size":
				var sizeStr string
				if !d.Args(&sizeStr) {
					return d.ArgErr()
				}
				var err error
				rl.MaxBodySize, err = parseSize(sizeStr)
				if err != nil {
					return d.Errf("invalid size: %v", err)
				}
			case "skip_methods":
				rl.SkipMethods = append(rl.SkipMethods, d.RemainingArgs()...)
			case "skip_paths":
				rl.SkipPaths = append(rl.SkipPaths, d.RemainingArgs()...)
			case "include_headers":
				rl.IncludeHeaders = append(rl.IncludeHeaders, d.RemainingArgs()...)
			case "exclude_headers":
				rl.ExcludeHeaders = append(rl.ExcludeHeaders, d.RemainingArgs()...)
			case "skip_content_types":
				rl.SkipContentTypes = append(rl.SkipContentTypes, d.RemainingArgs()...)
			default:
				return d.Errf("unknown directive: %s", d.Val())
			}
		}
	}
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*RequestLogger)(nil)
	_ caddyhttp.MiddlewareHandler = (*RequestLogger)(nil)
	_ caddyfile.Unmarshaler       = (*RequestLogger)(nil)
) 