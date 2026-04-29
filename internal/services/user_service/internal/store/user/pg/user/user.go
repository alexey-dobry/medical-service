package user

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/google/uuid"
)

func (r *Repository) AddPatient(userData model.User) error {
	_, err := r.db.Exec(
		"INSERT INTO user (first_name,middle_name,last_name,phone,email,sex,birth_date) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		userData.FirstName,
		userData.MiddleName,
		userData.LastName,
		userData.Phone,
		userData.Email,
		userData.Sex,
		userData.BirthDate,
	)
	return err
}

func (r *Repository) GetPatient(ID uuid.UUID) (model.User, error) {
	row := r.db.QueryRow("SELECT first_name,middle_name,last_name,phone,email,sex FROM user WHERE id = $1", ID)

	var u model.User
	err := row.Scan(&u.FirstName, &u.MiddleName, &u.LastName, &u.Phone, &u.Email, &u.Sex)
	if err != nil {
		return model.User{}, err
	}

	return u, nil
}

func (r *Repository) UpdateUser(ID uuid.UUID, updateData map[string]interface{}) error {
	if len(updateData) == 0 {
		return nil
	}

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
			return fmt.Errorf("field %q is not allowed for update", k)
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

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %s not found", ID)
	}

	return nil
}

func (r *Repository) DeleteUser(ID uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM user WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
