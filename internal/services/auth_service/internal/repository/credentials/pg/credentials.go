package pg

import (
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/google/uuid"
)

func (cr *PostgresCredentialsRepository) Add(credentials model.Credentials) error {
	_, err := cr.db.Exec(
		"INSERT INTO credentials (user_id,email,password_hash,role) VALUES ($1,$2,$3,$4)",
		credentials.UserID,
		credentials.Email,
		credentials.PasswordHash,
		credentials.Role,
	)

	return err
}

func (cr *PostgresCredentialsRepository) GetOneByMail(email string) (model.Credentials, error) {
	row := cr.db.QueryRow("SELECT user_id,password_hash,role FROM credentials WHERE email = $1", email)

	var c model.Credentials
	if err := row.Scan(&c.UserID, &c.PasswordHash, &c.Role); err != nil {
		return model.Credentials{}, err
	}

	return c, nil
}

func (cr *PostgresCredentialsRepository) GetOneByID(ID uuid.UUID) (model.Credentials, error) {
	row := cr.db.QueryRow("SELECT email,password_hash,role FROM credentials WHERE user_id = $1", ID)

	var c model.Credentials
	if err := row.Scan(&c.Email, &c.PasswordHash, &c.Role); err != nil {
		return model.Credentials{}, err
	}

	return c, nil
}

func (cr *PostgresCredentialsRepository) Delete(ID uuid.UUID) error {
	_, err := cr.db.Exec("DELETE FROM credentials WHERE user_id = $1", ID)

	return err
}
