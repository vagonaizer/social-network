package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"social-network/auth-service/internal/config"
	"social-network/auth-service/internal/service"
	"social-network/auth-service/internal/transport/grpc/handlers"
	pb "social-network/auth-service/pkg/api/proto/auth/v1"
	"social-network/auth-service/pkg/logger"
)

type Server struct {
	server *grpc.Server
	logger logger.Logger
	config *config.Config
}

func NewServer(
	cfg *config.Config,
	authService *service.AuthService,
	jwtService *service.JWTService,
	validationService *service.ValidationService,
	logger logger.Logger,
) *Server {
	// gRPC server options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.Server.GRPC.MaxReceiveSize),
		grpc.MaxSendMsgSize(cfg.Server.GRPC.MaxSendSize),
		grpc.ConnectionTimeout(cfg.Server.GRPC.ConnectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    cfg.Server.GRPC.KeepaliveTime,
			Timeout: cfg.Server.GRPC.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	server := grpc.NewServer(opts...)

	// Register services
	authHandler := handlers.NewAuthHandler(authService, jwtService, validationService, logger)
	pb.RegisterAuthServiceServer(server, authHandler)

	// Enable reflection for gRPC testing (always enabled for development)
	reflection.Register(server)

	logger.Info("gRPC server created with reflection enabled")

	return &Server{
		server: server,
		logger: logger,
		config: cfg,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.config.Server.GRPC.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.config.Server.GRPC.Port, err)
	}

	s.logger.Info("Starting gRPC server",
		logger.String("port", s.config.Server.GRPC.Port),
		logger.String("address", lis.Addr().String()),
		logger.Bool("reflection_enabled", true),
	)

	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping gRPC server")

	// Graceful stop with timeout
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		s.logger.Warn("gRPC server force stopped due to timeout")
		s.server.Stop()
		return ctx.Err()
	}
}
