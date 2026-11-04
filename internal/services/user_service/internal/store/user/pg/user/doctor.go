package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) AddDoctor(ctx context.Context, userData model.User, doctorData model.DoctorAdditionalData) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("userRepository - AddDoctor() - begin: %w", err)
	}
	// После успешного Commit Rollback ничего не делает.
	defer func() { _ = tx.Rollback(ctx) }() //nolint:errcheck // ошибка rollback не важна ни после commit, ни после предыдущей ошибки

	if _, err = tx.Exec(ctx, insertUserQuery, userArgs(userData)...); err != nil {
		return fmt.Errorf("userRepository - AddDoctor() - insert user: %w", mapError(err))
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO doctors (user_id,specialty,work_experience,description,services) VALUES ($1,$2,$3,$4,$5)",
		userData.ID,
		doctorData.Specialty,
		doctorData.WorkExperience,
		doctorData.Description,
		doctorData.Services,
	)
	if err != nil {
		return fmt.Errorf("userRepository - AddDoctor() - insert doctor: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("userRepository - AddDoctor() - commit: %w", err)
	}

	return nil
}

func (r *Repository) GetDoctor(ctx context.Context, ID uuid.UUID) (model.User, model.DoctorAdditionalData, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT u.id,u.first_name,u.middle_name,u.last_name,u.phone,u.email,u.sex,u.birth_date,COALESCE(u.photo_id::text,''),
		       d.specialty,d.work_experience,d.description,d.services
		FROM users u
		JOIN doctors d ON d.user_id = u.id
		WHERE u.id = $1`, ID)

	var u model.User
	var d model.DoctorAdditionalData
	err := row.Scan(
		&u.ID, &u.FirstName, &u.MiddleName, &u.LastName, &u.Phone, &u.Email, &u.Sex, &u.BirthDate, &u.PhotoID,
		&d.Specialty, &d.WorkExperience, &d.Description, &d.Services,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.DoctorAdditionalData{}, store.ErrNotFound
	} else if err != nil {
		return model.User{}, model.DoctorAdditionalData{}, fmt.Errorf("userRepository - GetDoctor(): %w", err)
	}
	d.UserID = u.ID

	return u, d, nil
}
