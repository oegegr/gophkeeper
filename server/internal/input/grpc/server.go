package grpc

import (
	"context"
	"log"

	_ "google.golang.org/grpc/reflection"

	pb "github.com/oegegr/gophkeeper/proto"
	"github.com/oegegr/gophkeeper/server/internal/domain"
	"github.com/samber/do/v2"
)

// AuthUseCase интерфейс для аутентификации
type AuthUseCase interface {
	Register(ctx context.Context, login, password string) (*domain.User, string, error)
	Login(ctx context.Context, login, password string) (*domain.User, string, error)
}

// SecretsUseCase интерфейс для работы с секретами
type SecretsUseCase interface {
	GetSecrets(ctx context.Context, userID string) ([]domain.Secret, error)
	GetSecret(ctx context.Context, userID, secretID string) (domain.Secret, error)
	AddSecret(ctx context.Context, userID string, secretType domain.SecretType, data []byte, meta string) (domain.Secret, error)
	UpdateSecret(ctx context.Context, userID, secretID string, secretType domain.SecretType, data []byte, meta string, version int) (domain.Secret, error)
	DeleteSecret(ctx context.Context, userID, secretID string) error
}

// SyncUseCase интерфейс для синхронизации
type SyncUseCase interface {
	Sync(ctx context.Context, userID string, clientSecrets []domain.Secret) ([]domain.Secret, error)
}

// UserProvider интерфейс для получения login из токена
type UserProvider interface {
	Get(ctx context.Context) (string, error)
}

type GophKeeperServer struct {
	pb.UnimplementedGophKeeperServer
	authUC       AuthUseCase
	secretUC     SecretsUseCase
	syncUC       SyncUseCase
	userProvider UserProvider
}

func ResolveGophKeeperServer(i do.Injector) *GophKeeperServer {
	authUC := do.MustInvokeAs[AuthUseCase](i)
	secretsUC := do.MustInvokeAs[SecretsUseCase](i)
	syncUC := do.MustInvokeAs[SyncUseCase](i)
	userProvider := do.MustInvokeAs[UserProvider](i)
	return NewGophKeeperServer(authUC, secretsUC, syncUC, userProvider)
}

func NewGophKeeperServer(authUC AuthUseCase, secretUC SecretsUseCase, syncUC SyncUseCase, userProvider UserProvider) *GophKeeperServer {
	return &GophKeeperServer{authUC: authUC, secretUC: secretUC, syncUC: syncUC, userProvider: userProvider}
}

// Register - регистрация пользователя
func (s *GophKeeperServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Register called with login: %s", req.GetLogin())

	user, token, err := s.authUC.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return NewMapper().ToProtoRegisterResponse(user, token), nil
}

// Login - аутентификация
func (s *GophKeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Login called with login: %s", req.GetLogin())

	user, token, err := s.authUC.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return NewMapper().ToProtoLoginResponse(user, token), nil
}

// GetSecrets - получение всех секретов
func (s *GophKeeperServer) GetSecrets(ctx context.Context, req *pb.GetSecretsRequest) (*pb.GetSecretsResponse, error) {
	log.Printf("GetSecrets called")

	userID, err := s.userProvider.Get(ctx)
	secrets, err := s.secretUC.GetSecrets(ctx, userID)
	if err != nil {
		return nil, err
	}

	protoSecrets, err := NewMapper().ToProtoSecrets(secrets)
	if err != nil {
		return nil, err
	}

	return &pb.GetSecretsResponse{
		Secrets: protoSecrets,
	}, nil
}

// AddSecret - добавление нового секрета
func (s *GophKeeperServer) AddSecret(ctx context.Context, req *pb.AddSecretRequest) (*pb.AddSecretResponse, error) {
	log.Printf("AddSecret called with type: %v, meta: %s", req.GetType(), req.GetMeta())

	userID, err := s.userProvider.Get(ctx)
	if err != nil {
		return nil, err
	}

	mapper := NewMapper()

	secretType, err := mapper.FromSecretType(req.Type)
	if err != nil {
		return nil, err 
	}

	secret, err := s.secretUC.AddSecret(ctx, userID, secretType, req.GetData(), req.GetMeta())
	if err != nil {
		return nil, err
	}

	protoSecret, err := mapper.ToProtoSecret(secret)
	if err != nil {
		return nil, err
	}

	return &pb.AddSecretResponse{
		Secret: protoSecret,
	}, nil
}

// UpdateSecret - обновление секрета
func (s *GophKeeperServer) UpdateSecret(ctx context.Context, req *pb.UpdateSecretRequest) (*pb.UpdateSecretResponse, error) {
	log.Printf("UpdateSecret called for id: %s", req.GetId())

	userID, err := s.userProvider.Get(ctx)
	if err != nil {
		return nil, err
	}

	secret, err := s.secretUC.UpdateSecret(ctx, userID, req.GetId(), domain.SecretType(req.GetType()), req.GetData(), req.GetMeta(), int(req.GetVersion()))
	if err != nil {
		return nil, err
	}

	protoSecret, err := NewMapper().ToProtoSecret(secret)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateSecretResponse{
		Secret: protoSecret,
	}, nil
}

// DeleteSecret - удаление секрета
func (s *GophKeeperServer) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	log.Printf("DeleteSecret called for id: %s", req.GetId())

	userID, err := s.userProvider.Get(ctx)
	if err != nil {
		return nil, err
	}

	err = s.secretUC.DeleteSecret(ctx, userID, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.DeleteSecretResponse{
		Success: true,
	}, nil
}

// Sync - синхронизация данных
func (s *GophKeeperServer) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	log.Printf("Sync called with %d secrets", len(req.GetSecrets()))

	userID, err := s.userProvider.Get(ctx)
	if err != nil {
		return nil, err
	}

	secrets, err := NewMapper().FromProtoSecrets(req.GetSecrets(), userID)
	if err != nil {
		return nil, err
	}

	syncedSecrets, err := s.syncUC.Sync(ctx, userID, secrets)
	if err != nil {
		return nil, err
	}

	protoSecrets, err := NewMapper().ToProtoSecrets(syncedSecrets)
	if err != nil {
		return nil, err
	}

	return &pb.SyncResponse{
		Secrets: protoSecrets,
	}, nil
}
