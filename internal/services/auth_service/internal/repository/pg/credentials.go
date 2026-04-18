package pg

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/google/uuid"
)

func (cr *CredentialsRepository) Add(credentials model.Credentials) error {
	_, err := cr.db.Exec("INSERT INTO credentials (user_id,email,password_hash) VALUES ($1,$2,$3)", credentials.UserID, credentials.Email, credentials.PasswordHash)
	return err
}

func (cr *CredentialsRepository) GetOneByMail(email string) (model.Credentials, error) {
	row := cr.db.QueryRow("SELECT user_id,password_hash FROM credentials WHERE email = $1", email)

	var c model.Credentials
	err := row.Scan(&c.UserID, &c.PasswordHash)
	if err != nil {
		return model.Credentials{}, err
	}

	return c, nil
}

func (cr *CredentialsRepository) GetOneByID(ID uuid.UUID) (model.Credentials, error) {
	row := cr.db.QueryRow("SELECT email,password_hash FROM credentials WHERE user_id = $1", ID)

	var c model.Credentials
	err := row.Scan(&c.Email, &c.PasswordHash)
	if err != nil {
		return model.Credentials{}, err
	}

	return c, nil
}
