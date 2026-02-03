// Package server provides HTTP server for API Gateway
package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	"net/http"
	"time"

	stdredis "github.com/redis/go-redis/v9"
	"github.com/wangjialin/myops/api-gateway/internal/config"
	"github.com/wangjialin/myops/api-gateway/internal/handler"
	"github.com/wangjialin/myops/api-gateway/internal/middleware"
	"github.com/wangjialin/myops/api-gateway/internal/service"
	"github.com/wangjialin/myops/pkg/auth/jwt"
	ldapauth "github.com/wangjialin/myops/pkg/auth/ldap"
	"github.com/wangjialin/myops/pkg/auth/redis"
	"github.com/wangjialin/myops/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	config     *config.Config
	db         *gorm.DB
	redis      *stdredis.Client
}

// New creates a new HTTP server
func New(cfg *config.Config, logger *zap.Logger) *Server {
	mux := http.NewServeMux()

	// TODO: Initialize dependencies
	// For now, create nil dependencies (service will handle nil gracefully)
	var gormDB *gorm.DB
	var redisClient *stdredis.Client
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey

	// Create repositories
	var userRepo *db.UserRepository
	if gormDB != nil {
		userRepo = db.NewUserRepository(gormDB)
	}

	// Create JWT manager
	jwtManager := jwt.NewManager(privateKey, publicKey, jwt.Config{
		Issuer:          cfg.JWT.Issuer,
		AccessDuration:  cfg.JWT.AccessDuration,
		RefreshDuration: cfg.JWT.RefreshDuration,
	})

	// Create refresh token repository
	var tokenRepo *redis.RefreshTokenRepository
	if redisClient != nil {
		tokenRepo = redis.NewRefreshTokenRepository(redisClient)
	}

	// Create LDAP client if configured
	var ldapClient *ldapauth.Client
	if cfg.LDAP.URL != "" {
		ldapConfig := ldapauth.Config{
			URL:               cfg.LDAP.URL,
			BindDN:            cfg.LDAP.BindDN,
			BindPassword:      cfg.LDAP.BindPassword,
			BaseDN:            cfg.LDAP.BaseDN,
			SearchFilter:      cfg.LDAP.UserFilter,
			UserAttributes:    []string{"uid", "cn", "mail", "displayName"},
			UseTLS:            false,
			InsecureSkipVerify: false,
			Timeout:           10 * time.Second,
		}
		ldapClient = ldapauth.NewClient(ldapConfig)
	}

	// Create services (pass nil for now, will be properly initialized in production)
	authService := service.NewAuthService(userRepo, jwtManager, tokenRepo, ldapClient)

	// Create host, scan, agent, SSH and file handlers (requires database)
	var hostHandler *handler.HostHandler
	var scanHandler *handler.ScanHandler
	var agentHandler *handler.AgentHandler
	var sshWSHandler *handler.SSHWebSocketHandler
	var fileHandler *handler.FileTransferHandler
	if gormDB != nil {
		hostHandler = handler.NewHostHandler(gormDB)
		scanHandler = handler.NewScanHandler(gormDB)
		agentHandler = handler.NewAgentHandler(gormDB)
		sshWSHandler = handler.NewSSHWebSocketHandler(gormDB, nil) // TODO: pass proper logger
		fileHandler = handler.NewFileTransferHandler(gormDB)
	}

	// Register handlers
	mux.Handle("/api/v1/auth/register", handler.NewRegisterHandler(authService))
	mux.Handle("/api/v1/auth/login", handler.NewLoginHandler(authService))
	mux.Handle("/api/v1/auth/ldap-login", handler.NewLDAPLoginHandler(authService))
	mux.Handle("/api/v1/auth/refresh", handler.NewRefreshTokenHandler(authService))
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/api/", handler.API) // Catch-all for API routes

	// Register SSH WebSocket handler (before middleware)
	if sshWSHandler != nil {
		mux.Handle("/ws/ssh/", sshWSHandler)
	}

	// Register host, scan and agent handlers
	handler.RegisterHandlers(hostHandler, scanHandler, agentHandler)

	// Register file transfer handler
	if fileHandler != nil {
		handler.RegisterFileHandler(fileHandler)
	}

	// Apply middleware chain
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:5173"}
	h := middleware.Chain(
		middleware.Recovery(logger),
		middleware.Logger(logger),
		middleware.RateLimit,
		middleware.CORS(allowedOrigins),
		middleware.Auth,
	)(mux)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		config:     cfg,
		db:         gormDB,
		redis:      redisClient,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting API Gateway",
		zap.String("addr", s.httpServer.Addr),
		zap.Duration("read_timeout", s.httpServer.ReadTimeout),
		zap.Duration("write_timeout", s.httpServer.WriteTimeout),
	)

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return s.httpServer.Serve(listener)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down API Gateway")

	// Close Redis connection if available
	if s.redis != nil {
		if err := s.redis.Close(); err != nil {
			s.logger.Error("failed to close redis connection", zap.Error(err))
		}
	}

	// Close database connection if available
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	return s.httpServer.Shutdown(ctx)
}
