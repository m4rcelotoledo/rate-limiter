package limiter

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// MockStorage implementa StorageStrategy para testes
type MockStorage struct {
	data map[string]int64
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]int64),
	}
}

func (m *MockStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.data[key]++
	return m.data[key], nil
}

func (m *MockStorage) Get(ctx context.Context, key string) (int64, error) {
	return m.data[key], nil
}

func (m *MockStorage) Set(ctx context.Context, key string, value int64, expiration time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestRateLimiter_CheckLimit(t *testing.T) {
	config := &Config{
		IPRequestsPerSecond:       5,
		IPBlockDurationSeconds:    60,
		TokenRequestsPerSecond:    10,
		TokenBlockDurationSeconds: 120,
	}

	storage := NewMockStorage()
	limiter := NewRateLimiter(storage, config)

	tests := []struct {
		name       string
		identifier string
		limitType  string
		requests   int
		expected   bool
	}{
		{
			name:       "IP limit not exceeded",
			identifier: "192.168.1.1",
			limitType:  "ip",
			requests:   3,
			expected:   true,
		},
		{
			name:       "IP limit exceeded",
			identifier: "192.168.1.2",
			limitType:  "ip",
			requests:   6,
			expected:   false,
		},
		{
			name:       "Token limit not exceeded",
			identifier: "abc123",
			limitType:  "token",
			requests:   8,
			expected:   true,
		},
		{
			name:       "Token limit exceeded",
			identifier: "def456",
			limitType:  "token",
			requests:   12,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Simula múltiplas requisições
			for i := 0; i < tt.requests; i++ {
				result, err := limiter.CheckLimit(ctx, tt.identifier, tt.limitType)
				if err != nil {
					t.Errorf("CheckLimit() error = %v", err)
					return
				}

				// Verifica se o resultado está correto na última requisição
				if i == tt.requests-1 {
					if result.Allowed != tt.expected {
						t.Errorf("CheckLimit() = %v, expected %v", result.Allowed, tt.expected)
					}
				}
			}
		})
	}
}

func TestRateLimiter_ExtractTokenFromHeader(t *testing.T) {
	limiter := &RateLimiter{}

	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name:     "No API_KEY header",
			headers:  map[string]string{},
			expected: "",
		},
		{
			name:     "API_KEY header present",
			headers:  map[string]string{"API_KEY": "test-token-123"},
			expected: "test-token-123",
		},
		{
			name:     "API_KEY header with spaces",
			headers:  map[string]string{"API_KEY": "  test-token-456  "},
			expected: "test-token-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := limiter.ExtractTokenFromHeader(req)
			if result != tt.expected {
				t.Errorf("ExtractTokenFromHeader() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRateLimiter_GetClientIP(t *testing.T) {
	limiter := &RateLimiter{}

	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For header",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.100"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "192.168.1.100",
		},
		{
			name:       "X-Real-IP header",
			headers:    map[string]string{"X-Real-IP": "10.0.0.1"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "10.0.0.1",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.50:8080",
			expected:   "192.168.1.50:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := limiter.GetClientIP(req)
			if result != tt.expected {
				t.Errorf("GetClientIP() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
