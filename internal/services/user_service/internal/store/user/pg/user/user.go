package user

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const insertUserQuery = `INSERT INTO users (id,first_name,middle_name,last_name,phone,email,sex,birth_date,photo_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

func userArgs(u model.User) []any {
	var photoID *string
	if u.PhotoID != "" {
		photoID = &u.PhotoID
	}

	return []any{u.ID, u.FirstName, u.MiddleName, u.LastName, u.Phone, u.Email, u.Sex, u.BirthDate, photoID}
}

func (r *Repository) AddPatient(ctx context.Context, userData model.User) error {
	if _, err := r.db.Pool.Exec(ctx, insertUserQuery, userArgs(userData)...); err != nil {
		return fmt.Errorf("userRepository - AddPatient(): %w", mapError(err))
	}

	return nil
}

func (r *Repository) GetPatient(ctx context.Context, ID uuid.UUID) (model.User, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT id,first_name,middle_name,last_name,phone,email,sex,birth_date,COALESCE(photo_id::text,'')
		FROM users WHERE id = $1`, ID)

	var u model.User
	err := row.Scan(&u.ID, &u.FirstName, &u.MiddleName, &u.LastName, &u.Phone, &u.Email, &u.Sex, &u.BirthDate, &u.PhotoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, store.ErrNotFound
	} else if err != nil {
		return model.User{}, fmt.Errorf("userRepository - GetPatient(): %w", err)
	}

	return u, nil
}

func (r *Repository) UpdateUser(ctx context.Context, ID uuid.UUID, updateData map[string]interface{}) error {
	if len(updateData) == 0 {
		return nil
	}

	// Имена колонок нельзя передать параметром — подставляем только разрешённые поля.
	allowedFields := map[string]bool{
		"middle_name": true,
		"first_name":  true,
		"last_name":   true,
		"phone":       true,
		"sex":         true,
	}

	keys := make([]string, 0, len(updateData))
	for k := range updateData {
		if !allowedFields[k] {
			return fmt.Errorf("%w: %q", store.ErrInvalidField, k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	setClauses := make([]string, 0, len(keys))
	args := make([]interface{}, 0, len(keys)+1)

	for i, key := range keys {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, i+1))
		args = append(args, updateData[key])
	}

	args = append(args, ID)

	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		len(args),
	)

	result, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("userRepository - UpdateUser(): %w", err)
	}

	if result.RowsAffected() == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, ID uuid.UUID) error {
	// doctors и photo_meta уйдут каскадом (ON DELETE CASCADE).
	if _, err := r.db.Pool.Exec(ctx, "DELETE FROM users WHERE id = $1", ID); err != nil {
		return fmt.Errorf("userRepository - DeleteUser(): %w", err)
	}

	return nil
}
