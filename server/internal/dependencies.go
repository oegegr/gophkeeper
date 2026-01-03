package internal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/oegegr/gophkeeper/server/internal/adapter/postgresql"
	"github.com/oegegr/gophkeeper/server/internal/adapter/tokens"
	"github.com/oegegr/gophkeeper/server/internal/config"
	"github.com/oegegr/gophkeeper/server/internal/input/grpc"
	"github.com/oegegr/gophkeeper/server/internal/usecases"
	"github.com/samber/do/v2"
)

var di *do.RootScope

func init() {
	di = do.New()
}

func registerServices(
	cfg *config.Config,
	pgconn *sql.DB,
) func(context.Context) error {
	do.ProvideValue(di, cfg)
	do.ProvideValue(di, pgconn)

	do.ProvideValue(di, tokens.ResolveContextUserProvider(di))
	do.ProvideValue(di, tokens.ResolveJWTTokenManager(di))
	do.ProvideValue(di, postgresql.ResolvePosgresRepository(di))
	do.ProvideValue(di, usecases.ResolveAuthUseCase(di))
	do.ProvideValue(di, usecases.ResolveSecretsUseCase(di))
	do.ProvideValue(di, usecases.ResolveSyncUseCase(di))
	do.ProvideValue(di, grpc.ResolveGophKeeperServer(di))

	fmt.Printf("DEBUG: Available services after: %v\n", di.ListProvidedServices())
	return MakeStopFn(di)
}

func MakeStopFn(di *do.RootScope) func(context.Context) error {
	return func(ctx context.Context) error {
		if errs := di.ShutdownWithContext(ctx); errs != nil {
			return errs
		}
		return nil
	}
}
