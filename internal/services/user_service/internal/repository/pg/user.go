package pg

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/google/uuid"
)

func (ur *UserRepository) AddPatient(userData model.User) error {
	_, err := ur.db.Exec(
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

func (ur *UserRepository) GetPatient(ID uuid.UUID) (model.User, error) {
	row := ur.db.QueryRow("SELECT first_name,middle_name,last_name,phone,email,sex FROM user WHERE id = $1", ID)

	var u model.User
	err := row.Scan(&u.FirstName, &u.MiddleName, &u.LastName, &u.Phone, &u.Email, &u.Sex)
	if err != nil {
		return model.User{}, err
	}

	return u, nil
}
