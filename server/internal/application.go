package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/oegegr/gophkeeper/proto"
	"github.com/oegegr/gophkeeper/server/internal/adapter/postgresql"
	"github.com/oegegr/gophkeeper/server/internal/config"
	grpcapp "github.com/oegegr/gophkeeper/server/internal/input/grpc"
	"github.com/samber/do/v2"
)

// Application управляет жизненным циклом приложения
type Application struct {
	config    *config.Config
	server    *grpc.Server
	lis       net.Listener
	isRunning bool
	stopChan  chan struct{}
	stopFNs   []func(context.Context) error
}

// NewApplication создает новое приложение
func NewApplication(cfg *config.Config) *Application {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &Application{
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

// Start запускает сервер
func (app *Application) Start(appctx context.Context) error {
	// Создаем listener
	lis, err := net.Listen("tcp", app.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	app.lis = lis

	// Регистрируем сервис
	app.server, err = app.buildServer(appctx)
	if err != nil {
		return fmt.Errorf("failed to register services: %v", err)
	}

	// Включаем reflection
	if app.config.EnableReflection {
		reflection.Register(app.server)
		log.Println("gRPC reflection enabled")
	}

	// Запускаем сервер
	app.isRunning = true
	log.Printf("Server starting on %s", app.config.Addr)

	go func() {
		if err := app.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Printf("Server error: %v", err)
		}
		app.isRunning = false
		close(app.stopChan)
	}()

	return nil
}

// Stop останавливает сервер
func (app *Application) Stop() error {
	if app.server == nil || !app.isRunning {
		return nil
	}

	log.Println("Shutting down server...")

	// Graceful shutdown с таймаутом
	done := make(chan struct{})
	go func() {
		app.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Server stopped gracefully")
		return nil
	case <-time.After(app.config.ShutdownTimeout):
		app.server.Stop()
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// Wait блокируется до остановки сервера
func (app *Application) Wait() <-chan struct{} {
	return app.stopChan
}

// registerServices регистрирует все gRPC сервисы
func (app *Application) buildServer(appctx context.Context) (*grpc.Server, error) {

	pgconn, err := postgresql.NewPGConn(appctx, app.config.Dsa)
	if err != nil {
		return nil, err
	}
	registerServices(app.config, pgconn)
	// Создаем зависимости
	handler := do.MustInvokeAs[pb.GophKeeperServer](di)
	authInterceptor := grpcapp.ResolveAuthInterceptor(di)

	interceptors := []grpc.UnaryServerInterceptor{
		authInterceptor,
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors...),

		// Дополнительные опции
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
		grpc.ConnectionTimeout(30 * time.Second),
	}

	server := grpc.NewServer(opts...)

	// Регистрируем сервисы
	pb.RegisterGophKeeperServer(server, handler)

	return server, nil
}
