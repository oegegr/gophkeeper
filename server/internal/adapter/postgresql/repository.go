package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oegegr/gophkeeper/server/internal/domain"

	// "github.com/oegegr/gophkeeper/server/internal/usecases"
	"github.com/samber/do/v2"
)

// var _ usecases.AuthRepository = (*PostgresRepository)(nil)
// var _ usecases.SecretsRepository = (*PostgresRepository)(nil)
// var _ usecases.SyncRepository = (*PostgresRepository)(nil)

// PostgresRepository реализация хранилища для PostgreSQL
type PostgresRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func ResolvePosgresRepository(i do.Injector) *PostgresRepository {
	db := do.MustInvoke[*sql.DB](i)
	return NewPostgresRepository(db)
}

// NewPostgresRepository создает новое PostgreSQL хранилище
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &PostgresRepository{
		db:   db,
		psql: psql,
	}
}

// FindSecretByID ищет секрет по ID
func (s *PostgresRepository) FindSecretByID(ctx context.Context, secretID string) (domain.Secret, error) {
	query := s.psql.Select(
		ColumnID,
		ColumnUserID,
		ColumnType,
		ColumnData,
		ColumnMeta,
		ColumnVersion,
		ColumnCreatedAt,
		ColumnUpdatedAt,
	).
		From(TableSecrets).
		Where(sq.Eq{ColumnID: secretID}).
		Limit(1)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return domain.Secret{}, fmt.Errorf("failed to build SQL: %w", err)
	}

	var secret domain.Secret
	var metaData []byte
	var createdAt, updatedAt time.Time
	var secretType string
	var dataString string

	// Используем QueryRowContext вместо QueryRow
	row := s.db.QueryRowContext(ctx, sqlStr, args...)
	err = row.Scan(
		&secret.ID,
		&secret.UserID,
		&secretType,
		&dataString,
		&metaData,
		&secret.Version,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Secret{}, domain.ErrNotFound
		}
		return domain.Secret{}, fmt.Errorf("failed to query secret: %w", err)
	}


	// Конвертируем тип
	secret.Type = domain.SecretType(secretType)
	secret.CreatedAt = createdAt
	secret.UpdatedAt = updatedAt

	// Парсим метаданные если они есть
	if len(metaData) > 0 {
		secret.Meta = string(metaData)
	}

	return secret, nil
}

// FindSecretsByUserID ищет секреты по ID пользователя
func (s *PostgresRepository) FindSecretsByUserID(ctx context.Context, userID string) ([]domain.Secret, error) {
	query := s.psql.Select(
		ColumnID,
		ColumnUserID,
		ColumnType,
		ColumnData,
		ColumnMeta,
		ColumnVersion,
		ColumnCreatedAt,
		ColumnUpdatedAt,
	).
		From(TableSecrets).
		Where(sq.Eq{ColumnUserID: userID}).
		OrderBy(ColumnCreatedAt + " DESC")
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	// Используем QueryContext вместо Query
	rows, err := s.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer rows.Close()

	var secrets []domain.Secret

	for rows.Next() {
		var secret domain.Secret
		var metaData []byte
		var createdAt, updatedAt time.Time
		var secretType string
		var dataString string

		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secretType,
			&dataString,
			&metaData,
			&secret.Version,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret row: %w", err)
		}

		secret.Type = domain.SecretType(secretType)
		secret.CreatedAt = createdAt
		secret.UpdatedAt = updatedAt

		if len(metaData) > 0 {
			secret.Meta = string(metaData)
		}

		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return secrets, nil
}

// DeleteSecret удаляет секрет
func (s *PostgresRepository) DeleteSecret(ctx context.Context, secretID string) error {
	query := s.psql.Delete(TableSecrets).
		Where(sq.Eq{ColumnID: secretID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL: %w", err)
	}

	// Используем ExecContext вместо Exec
	result, err := s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateSecret создает новый секрет
func (s *PostgresRepository) CreateSecret(ctx context.Context, secret domain.Secret) error {
	query := s.psql.Insert(TableSecrets).
		Columns(
			ColumnID,
			ColumnUserID,
			ColumnType,
			ColumnData,
			ColumnMeta,
			ColumnVersion,
			ColumnCreatedAt,
			ColumnUpdatedAt,
		).
		Values(
			secret.ID,
			secret.UserID,
			string(secret.Type),
			secret.Data,
			secret.Meta,
			secret.Version,
			time.Now(),
			time.Now(),
		)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL: %w", err)
	}

	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				return domain.ErrAlreadyExists
			case "23503": // foreign_key_violation
				return domain.ErrUserNotFound
			}
		}
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// UpdateSecret обновляет существующий секрет
func (s *PostgresRepository) UpdateSecret(ctx context.Context, secret domain.Secret) error {
	query := s.psql.Update(TableSecrets).
		Set(ColumnType, string(secret.Type)).
		Set(ColumnData, secret.Data).
		Set(ColumnMeta, secret.Meta).
		Set(ColumnVersion, sq.Expr(ColumnVersion+" + 1")).
		Set(ColumnUpdatedAt, time.Now()).
		Where(sq.Eq{ColumnID: secret.ID}).
		Where(sq.Eq{ColumnVersion: secret.Version}) // optimistic locking

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL: %w", err)
	}

	result, err := s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Проверяем, существует ли секрет
		_, err := s.FindSecretByID(ctx, secret.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound
			}
			return err
		}
		// Если секрет существует, значит версия изменилась
		return domain.ErrVersionConflict
	}

	// Инкрементируем версию
	secret.Version++
	secret.UpdatedAt = time.Now()

	return nil
}

// FindByLogin ищет пользователя по логину
func (s *PostgresRepository) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := s.psql.Select(
		ColumnID,
		ColumnLogin,
		ColumnPasswordHash,
		ColumnCreatedAt,
		ColumnUpdatedAt,
	).
		From(TableUsers).
		Where(sq.Eq{ColumnLogin: login}).
		Limit(1)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	var user domain.User
	var createdAt, updatedAt time.Time

	row := s.db.QueryRowContext(ctx, sqlStr, args...)
	err = row.Scan(
		&user.ID,
		&user.Login,
		&user.Password, // исправлено: PasswordHash вместо Password
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return &user, nil
}

// CreateUser создает нового пользователя
func (s *PostgresRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := s.psql.Insert(TableUsers).
		Columns(
			ColumnID,
			ColumnLogin,
			ColumnPasswordHash,
			ColumnCreatedAt,
			ColumnUpdatedAt,
		).
		Values(
			user.ID,
			user.Login,
			user.Password, // исправлено: PasswordHash вместо Password
			time.Now(),
			time.Now(),
		)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL: %w", err)
	}

	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// BeginTx начинает транзакцию с помощью sql.DB
func (s *PostgresRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}
