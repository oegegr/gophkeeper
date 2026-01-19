package internal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/oegegr/gophkeeper/server/internal/adapter/postgresql"
	"github.com/oegegr/gophkeeper/server/internal/adapter/tokens"
	"github.com/oegegr/gophkeeper/server/internal/config"
	"github.com/oegegr/gophkeeper/server/internal/input/grpc"
	"github.com/oegegr/gophkeeper/server/internal/usecases"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)

var di *do.RootScope

func init() {
	di = do.New()
}

type StopDI func(context.Context) error

func InitDependencies(
	cfg *config.Config,
	pgconn *sql.DB,
) (do.Injector, StopDI) {
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
	return di, MakeStopFn(di, cfg.ShutdownTimeout)
}

func MakeStopFn(di *do.RootScope, timeout time.Duration) StopDI {
	return func(ctx context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if errs := di.ShutdownWithContext(ctxWithTimeout); errs != nil {
			return errors.Wrap(errs, "failed to stop DI")
		}
		return nil
	}
}
