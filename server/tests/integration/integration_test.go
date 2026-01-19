package integration

import (
	"context"
	"fmt"

	pb "github.com/oegegr/gophkeeper/proto"
	"google.golang.org/grpc/metadata"
)

// testRegisterAndLogin тестирует регистрацию и логин
func (s *TestSuite) TestRegisterAndLogin() {
	require := s.Require()
	// Тест 1: Регистрация
	registerResp, err := s.client.Register(s.ctx, &pb.RegisterRequest{
		Login:    "testuser",
		Password: "TestPass123!",
	})
	require.NoError(err, "Register should succeed")
	require.NotEmpty(registerResp.GetToken(), "Token should not be empty")
	require.NotEmpty(registerResp.GetUserId(), "UserID should not be empty")

	// Тест 2: Логин с правильным паролем
	loginResp, err := s.client.Login(s.ctx, &pb.LoginRequest{
		Login:    "testuser",
		Password: "TestPass123!",
	})
	require.NoError(err, "Login should succeed")
	require.NotEmpty(loginResp.GetToken(), "Login token should not be empty")
	require.Equal(registerResp.GetUserId(), loginResp.GetUserId(), "UserID should match")

	// Тест 3: Логин с неправильным паролем
	_, err = s.client.Login(s.ctx, &pb.LoginRequest{
		Login:    "testuser",
		Password: "WrongPassword",
	})
	require.Error(err, "Login with wrong password should fail")
}

// testCreateAndGetSecret тестирует создание и получение секрета
func (s *TestSuite) TestCreateAndGetSecret() {
	require := s.Require()
	// 1. Сначала регистрируем пользователя
	registerResp, err := s.client.Register(s.ctx, &pb.RegisterRequest{
		Login:    "secretuser",
		Password: "SecretPass123!",
	})
	require.NoError(err)
	token := registerResp.GetToken()

	// 3. Создаем секрет
	secretData := []byte("my secret password")
	addResp, err := s.client.AddSecret(metadataContext(s.ctx, token), &pb.AddSecretRequest{
		Type: pb.DataType_LOGIN_PASSWORD,
		Data: secretData,
		Meta: "Test secret",
	})
	require.NoError(err, "Should add secret successfully")
	require.NotEmpty(addResp.GetSecret().GetId(), "Secret should have ID")
	require.Equal(pb.DataType_LOGIN_PASSWORD, addResp.GetSecret().GetType())
	require.Equal(secretData, addResp.GetSecret().GetData())
	require.Equal("Test secret", addResp.GetSecret().GetMeta())

	secretID := addResp.GetSecret().GetId()

	// 4. Получаем все секреты
	getResp, err := s.client.GetSecrets(metadataContext(s.ctx, token), &pb.GetSecretsRequest{})
	require.NoError(err, "Should get secrets successfully")
	require.Len(getResp.GetSecrets(), 1, "Should have one secret")
	require.Equal(secretID, getResp.GetSecrets()[0].GetId(), "Secret IDs should match")

	// 5. Пробуем получить секреты без токена (должна быть ошибка)
	_, err = s.client.GetSecrets(s.ctx, &pb.GetSecretsRequest{})
	require.Error(err, "Should fail without token")
}

func metadataContext(ctx context.Context, token string) context.Context {
	md := metadata.MD{"authorization": []string{fmt.Sprintf("Bearer %s", token)}}
	return metadata.NewOutgoingContext(ctx, md)
}
