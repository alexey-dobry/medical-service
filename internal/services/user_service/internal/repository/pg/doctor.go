package pg

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/google/uuid"
)

func (ur *UserRepository) AddDoctor(userData model.User, doctorData model.DoctorAdditionalData) error {
	tx, err := ur.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var userID string

	err = tx.QueryRow(
		"INSERT INTO user (first_name,middle_name,last_name,phone,email,sex,birth_date) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id",
		userData.FirstName,
		userData.MiddleName,
		userData.LastName,
		userData.Phone,
		userData.Email,
		userData.Sex,
		userData.BirthDate,
	).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO doctor (user_id,specialty,work_experience,description) VALUES ($1,$2,$3,$4)",
		userID,
		doctorData.Specialty,
		doctorData.WorkExperience,
		doctorData.Description,
	)
	if err != nil {
		return err
	}

	return nil
}

func (ur *UserRepository) GetDoctor(ID uuid.UUID) (model.User, model.DoctorAdditionalData, error) {
	tx, err := ur.db.Begin()
	if err != nil {
		return model.User{}, model.DoctorAdditionalData{}, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var u model.User
	var d model.DoctorAdditionalData

	userRow := tx.QueryRow("SELECT first_name,middle_name,last_name,phone,email,sex FROM user WHERE id = $1", ID)

	err = userRow.Scan(&u.FirstName, &u.MiddleName, &u.LastName, &u.Phone, &u.Email, &u.Sex)
	if err != nil {
		return model.User{}, model.DoctorAdditionalData{}, err
	}

	doctorAdditionalDataRow := tx.QueryRow("SELECT specialty, work_experience, description FROM doctor WHERE user_id = $1", ID)

	err = doctorAdditionalDataRow.Scan(&d.Specialty, &d.WorkExperience, &d.Description)
	if err != nil {
		return model.User{}, model.DoctorAdditionalData{}, err
	}

	return u, d, nil
}
