package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/services"
	"github.com/ownafarm/ownafarm-backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetNonce_MissingWalletAddress(t *testing.T) {
	router := gin.New()

	// Create a handler with nil dependencies (we're testing validation only)
	handler := &AuthHandler{}
	router.GET("/auth/nonce", handler.GetNonce)

	req, _ := http.NewRequest("GET", "/auth/nonce", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "wallet_address is required", response["message"])
}

func TestGetNonce_InvalidWalletAddressFormat(t *testing.T) {
	router := gin.New()

	handler := &AuthHandler{}
	router.GET("/auth/nonce", handler.GetNonce)

	tests := []struct {
		name          string
		walletAddress string
	}{
		{"too short", "0x1234"},
		{"no 0x prefix", "1234567890abcdef1234567890abcdef12345678"},
		{"invalid characters", "0xGGGGGGGGGGabcdef1234567890abcdef12345678"},
		{"too long", "0x1234567890abcdef1234567890abcdef1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/auth/nonce?wallet_address="+tt.walletAddress, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "Invalid wallet address format", response["message"])
		})
	}
}

func TestLogin_InvalidRequestBody(t *testing.T) {
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/auth/login", handler.Login)

	// Test with an empty body
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid request body", response["message"])
}

func TestLogin_MissingFields(t *testing.T) {
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/auth/login", handler.Login)

	tests := []struct {
		name string
		body map[string]string
	}{
		{
			"missing wallet_address",
			map[string]string{"signature": "0x123", "nonce": "abc"},
		},
		{
			"missing signature",
			map[string]string{"wallet_address": "0x1234567890abcdef1234567890abcdef12345678", "nonce": "abc"},
		},
		{
			"missing nonce",
			map[string]string{"wallet_address": "0x1234567890abcdef1234567890abcdef12345678", "signature": "0x123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response["status"])
		})
	}
}

func TestLogin_InvalidWalletAddressFormat(t *testing.T) {
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/auth/login", handler.Login)

	body := map[string]string{
		"wallet_address": "invalid-wallet",
		"signature":      "0x123",
		"nonce":          "abc123",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid wallet address format", response["message"])
}

func TestWalletAddressRegex(t *testing.T) {
	validAddresses := []string{
		"0x1234567890abcdef1234567890abcdef12345678",
		"0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
	}

	invalidAddresses := []string{
		"1234567890abcdef1234567890abcdef12345678",    // no 0x prefix
		"0x1234567890abcdef1234567890abcdef1234567",   // too short
		"0x1234567890abcdef1234567890abcdef123456789", // too long
		"0xGGGGGGGGGGabcdef1234567890abcdef12345678",  // invalid hex chars
		"", // empty
	}

	for _, addr := range validAddresses {
		assert.True(t, walletAddressRegex.MatchString(addr), "Expected %s to be valid", addr)
	}

	for _, addr := range invalidAddresses {
		assert.False(t, walletAddressRegex.MatchString(addr), "Expected %s to be invalid", addr)
	}
}

// Mock implementations for testing

type mockNonceService struct {
	GenerateNonceFunc          func(ctx context.Context, walletAddress string) (string, error)
	ValidateAndDeleteNonceFunc func(ctx context.Context, walletAddress, nonce string) error
	BuildSignMessageFunc       func(nonce string) string
}

func (m *mockNonceService) GenerateNonce(ctx context.Context, walletAddress string) (string, error) {
	if m.GenerateNonceFunc != nil {
		return m.GenerateNonceFunc(ctx, walletAddress)
	}
	return "test-nonce-123", nil
}

func (m *mockNonceService) ValidateAndDeleteNonce(ctx context.Context, walletAddress, nonce string) error {
	if m.ValidateAndDeleteNonceFunc != nil {
		return m.ValidateAndDeleteNonceFunc(ctx, walletAddress, nonce)
	}
	return nil
}

func (m *mockNonceService) BuildSignMessage(nonce string) string {
	if m.BuildSignMessageFunc != nil {
		return m.BuildSignMessageFunc(nonce)
	}
	return "Sign this message to login to OwnaFarm.\n\nNonce: " + nonce
}

type mockAuthService struct {
	VerifySignatureFunc        func(walletAddress, signature, message string) error
	NormalizeWalletAddressFunc func(walletAddress string) string
}

func (m *mockAuthService) VerifySignature(walletAddress, signature, message string) error {
	if m.VerifySignatureFunc != nil {
		return m.VerifySignatureFunc(walletAddress, signature, message)
	}
	return nil
}

func (m *mockAuthService) NormalizeWalletAddress(walletAddress string) string {
	if m.NormalizeWalletAddressFunc != nil {
		return m.NormalizeWalletAddressFunc(walletAddress)
	}
	return walletAddress
}

type mockUserRepositoryForAuth struct {
	GetByIDFunc            func(id string) (*models.User, error)
	GetByWalletAddressFunc func(walletAddress string) (*models.User, error)
	CreateFunc             func(user *models.User) error
	UpdateLastLoginFunc    func(userID string) error
}

func (m *mockUserRepositoryForAuth) GetByID(id string) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *mockUserRepositoryForAuth) GetByWalletAddress(walletAddress string) (*models.User, error) {
	if m.GetByWalletAddressFunc != nil {
		return m.GetByWalletAddressFunc(walletAddress)
	}
	return nil, nil
}

func (m *mockUserRepositoryForAuth) Create(user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *mockUserRepositoryForAuth) UpdateLastLogin(userID string) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(userID)
	}
	return nil
}

// TestGetNonce_Success tests successful nonce generation
func TestGetNonce_Success(t *testing.T) {
	router := gin.New()

	mockNonce := &mockNonceService{
		GenerateNonceFunc: func(ctx context.Context, walletAddress string) (string, error) {
			return "abc123def456", nil
		},
		BuildSignMessageFunc: func(nonce string) string {
			return "Sign this message to login to OwnaFarm.\n\nNonce: " + nonce
		},
	}

	handler := &AuthHandler{
		nonceService: mockNonce,
	}
	router.GET("/auth/nonce", handler.GetNonce)

	walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
	req, _ := http.NewRequest("GET", "/auth/nonce?wallet_address="+walletAddress, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "abc123def456", data["nonce"])
	assert.Contains(t, data["message"], "Sign this message to login to OwnaFarm")
}

// TestLogin_Success_ExistingUser tests successful login for an existing user
func TestLogin_Success_ExistingUser(t *testing.T) {
	router := gin.New()

	existingUser := &models.User{
		ID:            "user-uuid-123",
		WalletAddress: "0x1234567890abcdef1234567890abcdef12345678",
	}

	mockUserRepo := &mockUserRepositoryForAuth{
		GetByWalletAddressFunc: func(walletAddress string) (*models.User, error) {
			return existingUser, nil
		},
		UpdateLastLoginFunc: func(userID string) error {
			return nil
		},
	}

	mockNonce := &mockNonceService{
		ValidateAndDeleteNonceFunc: func(ctx context.Context, walletAddress, nonce string) error {
			return nil
		},
		BuildSignMessageFunc: func(nonce string) string {
			return "Sign this message to login to OwnaFarm.\n\nNonce: " + nonce
		},
	}

	mockAuth := &mockAuthService{
		VerifySignatureFunc: func(walletAddress, signature, message string) error {
			return nil
		},
		NormalizeWalletAddressFunc: func(walletAddress string) string {
			return walletAddress
		},
	}

	jwtCfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-jwt-testing-32chars",
		ExpirationHours: 24,
	}
	jwtUtil := utils.NewJWTUtil(jwtCfg)

	handler := &AuthHandler{
		userRepo:     mockUserRepo,
		nonceService: mockNonce,
		authService:  mockAuth,
		jwtUtil:      jwtUtil,
	}
	router.POST("/auth/login", handler.Login)

	body := map[string]string{
		"wallet_address": "0x1234567890abcdef1234567890abcdef12345678",
		"signature":      "0x" + string(make([]byte, 130)), // 65 bytes hex encoded
		"nonce":          "test-nonce-123",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
	assert.NotNil(t, data["user"])
}

// TestLogin_Success_NewUser tests successful login and user creation for a new user
func TestLogin_Success_NewUser(t *testing.T) {
	router := gin.New()

	mockUserRepo := &mockUserRepositoryForAuth{
		GetByWalletAddressFunc: func(walletAddress string) (*models.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
		CreateFunc: func(user *models.User) error {
			user.ID = "new-user-uuid-456"
			return nil
		},
		UpdateLastLoginFunc: func(userID string) error {
			return nil
		},
	}

	mockNonce := &mockNonceService{
		ValidateAndDeleteNonceFunc: func(ctx context.Context, walletAddress, nonce string) error {
			return nil
		},
		BuildSignMessageFunc: func(nonce string) string {
			return "Sign this message to login to OwnaFarm.\n\nNonce: " + nonce
		},
	}

	mockAuth := &mockAuthService{
		VerifySignatureFunc: func(walletAddress, signature, message string) error {
			return nil
		},
		NormalizeWalletAddressFunc: func(walletAddress string) string {
			return walletAddress
		},
	}

	jwtCfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-jwt-testing-32chars",
		ExpirationHours: 24,
	}
	jwtUtil := utils.NewJWTUtil(jwtCfg)

	handler := &AuthHandler{
		userRepo:     mockUserRepo,
		nonceService: mockNonce,
		authService:  mockAuth,
		jwtUtil:      jwtUtil,
	}
	router.POST("/auth/login", handler.Login)

	body := map[string]string{
		"wallet_address": "0xabcdef1234567890abcdef1234567890abcdef12",
		"signature":      "0x" + string(make([]byte, 130)),
		"nonce":          "test-nonce-456",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
	assert.NotNil(t, data["user"])
}

// TestLogin_InvalidNonce tests login with invalid or expired nonce
func TestLogin_InvalidNonce(t *testing.T) {
	router := gin.New()

	mockNonce := &mockNonceService{
		ValidateAndDeleteNonceFunc: func(ctx context.Context, walletAddress, nonce string) error {
			return services.ErrNonceNotFound
		},
	}

	handler := &AuthHandler{
		nonceService: mockNonce,
	}
	router.POST("/auth/login", handler.Login)

	body := map[string]string{
		"wallet_address": "0x1234567890abcdef1234567890abcdef12345678",
		"signature":      "0x" + string(make([]byte, 130)),
		"nonce":          "invalid-nonce",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid or expired nonce", response["message"])
}

// TestLogin_SignatureMismatch tests login with signature that doesn't match wallet address
func TestLogin_SignatureMismatch(t *testing.T) {
	router := gin.New()

	mockNonce := &mockNonceService{
		ValidateAndDeleteNonceFunc: func(ctx context.Context, walletAddress, nonce string) error {
			return nil
		},
		BuildSignMessageFunc: func(nonce string) string {
			return "Sign this message to login to OwnaFarm.\n\nNonce: " + nonce
		},
	}

	mockAuth := &mockAuthService{
		VerifySignatureFunc: func(walletAddress, signature, message string) error {
			return services.ErrSignatureMismatch
		},
	}

	handler := &AuthHandler{
		nonceService: mockNonce,
		authService:  mockAuth,
	}
	router.POST("/auth/login", handler.Login)

	body := map[string]string{
		"wallet_address": "0x1234567890abcdef1234567890abcdef12345678",
		"signature":      "0x" + string(make([]byte, 130)),
		"nonce":          "test-nonce-789",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid signature", response["message"])
}
