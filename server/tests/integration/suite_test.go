package integration

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/oegegr/gophkeeper/proto"
	"github.com/oegegr/gophkeeper/server/internal"
	"github.com/oegegr/gophkeeper/server/internal/config"
)

type TestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer testcontainers.Container
	app         *internal.Application
	client      pb.GophKeeperClient
}

func (s *TestSuite) SetupSuite() {
	var dsa, addr string
	s.ctx = context.Background()
	s.pgContainer, dsa = s.startPostgresContainer()
	s.app, addr = s.startApplication(dsa)
	s.client = s.createGRPCClient(addr)
}

func (s *TestSuite) TearDownSuite() {
	stopTimeout := 5 * time.Second
	if s.app != nil {
		s.app.Stop()
	}

	if s.pgContainer != nil {
		s.pgContainer.Stop(s.ctx, &stopTimeout)
	}
}

func TestRoutes(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

// startPostgresContainer запускает Postgres в контейнере
func (s *TestSuite) startPostgresContainer() (testcontainers.Container, string) {
	require := s.Require()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_DB":       "gophkeeper_test",
		},
		WaitingFor: wait.ForExec([]string{"pg_isready", "-U", "test_user"}).
			WithPollInterval(2 * time.Second).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}),
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(err, "Failed to start Postgres container")

	// Получаем порт
	port, err := container.MappedPort(s.ctx, "5432")
	require.NoError(err)

	// Формируем DSN
	host, err := container.Host(s.ctx)
	require.NoError(err)

	pgDSN := fmt.Sprintf("postgres://test_user:test_password@%s:%s/gophkeeper_test?sslmode=disable",
		host, port.Port())

	time.Sleep(5 * time.Second)
	log.Printf("Postgres container started at: %s", pgDSN)

	return container, pgDSN
}

// startGRPCServer запускает gRPC сервер
func (s *TestSuite) startApplication(dsa string) (*internal.Application, string) {
	// Разбираем DSN для конфига
	// Для простоты создаем конфиг с локальными значениями
	serverAddr := "127.0.0.1:8080"
	cfg := &config.Config{
		Addr:             "127.0.0.1:8080", // 0 = случайный свободный порт
		EnableReflection: true,
		ShutdownTimeout:  10 * time.Second,
		TokenSecret:      "test-secret-key-12345",
		Dsa:              dsa, 
	}

	// Создаем приложение
	app := internal.NewApplication(cfg)

	// Запускаем приложение с нашим listener'ом
	go func() {
		if err := app.Start(s.ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Даем время на запуск
	time.Sleep(2 * time.Second)

	return app, serverAddr
}

// createGRPCClient создает gRPC клиент
func (s *TestSuite) createGRPCClient(addr string) pb.GophKeeperClient {
	require := s.Require()
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(err, "Failed to create gRPC client")

	return pb.NewGophKeeperClient(conn)
}
