package postgres

import (
	"context"
	"fmt"
	"pulse-auth/internal/model"
	"pulse-auth/internal/utils"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// LoginUser takes a user login input and checks the database for a matching user.
func (s *Storage) LoginUser(ctx context.Context, userLogin *model.UserLogin) (*model.User, error) {
	sql, args, err := sq.Select(userFields...).
		From(UserTable).
		Where(sq.Eq{
			fieldUsername:       userLogin.Username,
			fieldHashedPassword: userLogin.HashedPassword,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, utils.WrapInternalError(fmt.Errorf("incorrect sql"))
	}

	var entity userEntity
	err = s.db.GetContext(ctx, &entity, sql, args...)
	if err != nil {
		return nil, utils.WrapSqlError(err)
	}

	return userEntityToModel(entity), nil
}

// CreateUser takes in user registration params, validates them and creates a new user in the database.
func (s *Storage) CreateUser(ctx context.Context, params *model.UserRegister) (*model.User, error) {
	err := params.Validate()
	if err != nil {
		return nil, fmt.Errorf("params validate: %w", err)
	}
	now := time.Now().Truncate(time.Millisecond)
	sql, args, err := sq.Insert(UserTable).
		Columns(userFields...).
		Values(params.ID, params.Username, params.HashedPassword, params.FirstName, params.SecondName,
			params.Sex, params.Birthdate, params.Biography, params.City, now,
		).
		Suffix(returningUser).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, utils.WrapInternalError(fmt.Errorf("incorrect sql"))
	}
	var entity userEntity
	err = s.db.GetContext(ctx, &entity, sql, args...)
	if err != nil {
		return nil, utils.WrapSqlError(err)
	}

	return userEntityToModel(entity), nil
}

// GetUserByID retrieve a user by ID from the storage with additional debug info logging.
func (s *Storage) GetUserByID(ctx context.Context, id model.UserID) (*model.User, error) {
	sql, args, err := sq.Select(userFields...).
		From(UserTable).
		Where(sq.Eq{
			fieldID:        id.String(),
			fieldDeletedAt: nil,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, utils.WrapInternalError(fmt.Errorf("incorrect sql"))
	}

	var entity userEntity
	err = s.db.GetContext(ctx, &entity, sql, args...)
	if err != nil {
		return nil, utils.WrapSqlError(err)
	}

	return userEntityToModel(entity), nil
}

// SearchUser searches for a user in the storage using first and last name.
func (s *Storage) SearchUser(ctx context.Context, firstName, lastName string) (*model.User, error) {
	sql, args, err := sq.Select(userFields...).
		From(UserTable).
		Where(sq.Eq{
			fieldFirstName:  firstName,
			fieldSecondName: lastName,
			fieldDeletedAt:  nil,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, utils.WrapInternalError(fmt.Errorf("incorrect sql"))
	}

	var entity userEntity
	err = s.db.GetContext(ctx, &entity, sql, args...)
	if err != nil {
		return nil, utils.WrapSqlError(err)
	}

	return userEntityToModel(entity), nil
}

type userEntity struct {
	ID             string    `db:"id"`
	Username       string    `db:"username"`
	HashedPassword string    `db:"hashed_password"`
	FirstName      string    `db:"first_name"`
	SecondName     string    `db:"second_name"`
	Sex            string    `db:"sex"`
	Birthdate      time.Time `db:"birthdate"`
	Biography      string    `db:"biography"`
	City           string    `db:"city"`
	CreatedAt      time.Time `db:"created_at"`
}

// userEntityToModel converts a user entity to a model User instance, mapping the attributes accordingly.
func userEntityToModel(entity userEntity) *model.User {
	return &model.User{
		UserID:     model.UserID(entity.ID),
		Username:   entity.Username,
		FirstName:  entity.FirstName,
		SecondName: entity.SecondName,
		Sex:        entity.Sex,
		Birthdate:  entity.Birthdate,
		Biography:  entity.Biography,
		City:       entity.City,
	}
}
