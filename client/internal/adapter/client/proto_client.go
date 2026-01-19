package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/oegegr/gophkeeper/client/internal/config"
	"github.com/oegegr/gophkeeper/client/internal/domain"
	"github.com/oegegr/gophkeeper/client/internal/usecases"
	pb "github.com/oegegr/gophkeeper/proto"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)

var _ usecases.SessionClient = (*ProtoClient)(nil)
var _ usecases.SecretClient = (*ProtoClient)(nil)

type ProtoClient struct {
	conn    *grpc.ClientConn
	client  pb.GophKeeperClient
}

func ResolveClient(i do.Injector) *ProtoClient {
	cfg := do.MustInvoke[*config.Config](i)
	client, err := NewProtoClient(cfg.ServerAddress)
	if err != nil {
		panic(errors.Wrap(err, "failed to create client"))
	}
	return client
}

func NewProtoClient(addr string) (*ProtoClient, error) {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50*1024*1024)))

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to server %s", addr)
	}

	return &ProtoClient{
		conn:   conn,
		client: pb.NewGophKeeperClient(conn),
	}, nil
}

// Login выполняет аутентификацию
func (c *ProtoClient) Login(ctx context.Context, username, password string) (domain.Session, error) {
	req := &pb.LoginRequest{
		Login:    username,
		Password: password,
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return domain.Session{}, errors.Wrap(err, "failed to login")
	}

	return domain.Session{UserID: resp.UserId, Username: username, Token: resp.Token}, nil
}

// Register регистрирует нового пользователя
func (c *ProtoClient) Register(ctx context.Context, username, password string) (domain.Session, error) {
	req := &pb.RegisterRequest{
		Login:    username,
		Password: password,
	}

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return domain.Session{}, errors.Wrap(err, "failed to register")
	}

	return domain.Session{UserID: resp.UserId, Username: username, Token: resp.Token}, nil
}

// Sync синхронизирует данные
func (c *ProtoClient) Sync(ctx context.Context, secrets []domain.Secret, session domain.Session) ([]domain.Secret, error) {
	authCtx := c.getAuthContext(ctx, session.Token)

	localSecrets := make([]*pb.Secret, len(secrets))
	for i, secret := range secrets {
		localSecrets[i] = fromDomain(&secret)
	}

	req := &pb.SyncRequest{
		Secrets: localSecrets,
	}

	resp, err := c.client.Sync(authCtx, req)
	if err != nil {
		if c.isUnauthenticatedError(err) {
			return nil, domain.ErrUnauthenticated
		}
		return []domain.Secret{}, errors.Wrap(err, "failed to sync")
	}

	syncSecrets := make([]domain.Secret, len(resp.Secrets))
	for i, secret := range resp.Secrets {
		syncSecrets[i] = *toDomain(secret)
	}

	return syncSecrets, nil
}

// GetSecrets получает все секреты с сервера
func (c *ProtoClient) GetSecrets(ctx context.Context, session domain.Session) ([]domain.Secret, error) {
	authCtx := c.getAuthContext(ctx, session.Token)

	req := &pb.GetSecretsRequest{}

	resp, err := c.client.GetSecrets(authCtx, req)
	if err != nil {
		if c.isUnauthenticatedError(err) {
			return nil, domain.ErrUnauthenticated
		}
		return []domain.Secret{}, errors.Wrap(err, "failed to get secrets")
	}

	secrets := make([]domain.Secret, len(resp.Secrets))
	for i, secret := range resp.Secrets {
		secrets[i] = *toDomain(secret)
	}

	return secrets, nil
}

// AddSecret добавляет секрет
func (c *ProtoClient) AddSecret(ctx context.Context, secret domain.Secret, session domain.Session) (*domain.Secret, error) {
	authCtx := c.getAuthContext(ctx, session.Token)

	s := fromDomain(&secret)
	req := &pb.AddSecretRequest{
		Type: s.Type,
		Data: s.Data,
		Meta: s.Meta,
	}

	resp, err := c.client.AddSecret(authCtx, req)
	if err != nil {
		if c.isUnauthenticatedError(err) {
			return nil, domain.ErrUnauthenticated
		}
		return nil, errors.Wrap(err, "failed to add secret")
	}

	return toDomain(resp.Secret), nil
}

// UpdateSecret обновляет секрет
func (c *ProtoClient) UpdateSecret(ctx context.Context, secret domain.Secret, session domain.Session) (domain.Secret, error) {
	authCtx := c.getAuthContext(ctx, session.Token)

	s := fromDomain(&secret)
	req := &pb.UpdateSecretRequest{
		Id:      s.Id,
		Type:    s.Type,
		Data:    s.Data,
		Meta:    s.Meta,
		Version: s.Version,
	}

	resp, err := c.client.UpdateSecret(authCtx, req)
	if err != nil {
		if c.isUnauthenticatedError(err) {
			return domain.Secret{}, domain.ErrUnauthenticated
		}
		return domain.Secret{}, errors.Wrap(err, "failed to update secret")
	}

	updatedSecret := toDomain(resp.Secret)

	return *updatedSecret, nil
}

// DeleteSecret удаляет секрет
func (c *ProtoClient) DeleteSecret(ctx context.Context, secret domain.Secret, session domain.Session) error {
	authCtx := c.getAuthContext(ctx, session.Token)

	s := fromDomain(&secret)
	req := &pb.DeleteSecretRequest{
		Id: s.Id,
	}

	_, err := c.client.DeleteSecret(authCtx, req)
	if err != nil {
		if c.isUnauthenticatedError(err) {
			return domain.ErrUnauthenticated
		}
		return errors.Wrap(err, "failed to delete secret")
	}

	return nil
}

// Close закрывает соединение
func (c *ProtoClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *ProtoClient) getAuthContext(ctx context.Context, token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer " + token)
	return metadata.NewOutgoingContext(ctx, md)
}

// isUnauthenticatedError проверяет, является ли ошибка Unauthenticated
func (c *ProtoClient) isUnauthenticatedError(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.Unauthenticated
	}
	return false
}

func toDomain(p *pb.Secret) *domain.Secret {
	return &domain.Secret{
		ID:        p.Id,
		Type:      pb.DataType_name[int32(p.Type)],
		Data:      p.Data,
		Meta:      p.Meta,
		Version:   p.Version,
		CreatedAt: time.Now(), // TODO: парсить из protobuf
		UpdatedAt: time.Now(), // TODO: парсить из protobuf
	}
}

// Mapping from domain.Secret to proto.Secret
func fromDomain(d *domain.Secret) *pb.Secret {
	return &pb.Secret{
		Id:        d.ID,
		Type:      pb.DataType(pb.DataType_value[d.Type]),
		Data:      d.Data,
		Meta:      d.Meta,
		Version:   int32(d.Version),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}
}
