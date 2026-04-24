package meta

import "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"

func (r *Repository) Create(photo model.Photo) error {
	_, err := r.db.Exec(
		"INSERT INTO meta (id,name,mime_type,size,user_id,storage_key) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		photo.ID,
		photo.Name,
		photo.MimeType,
		photo.Size,
		photo.UserID,
		photo.StorageKey,
	)

	return err
}

func (r *Repository) GetByID(ID string) (model.Photo, error) {
	row := r.db.QueryRow("SELECT id,name,mime_type,size,user_id,storage_key FROM meta WHERE id = $1", ID)

	var p model.Photo
	err := row.Scan(&p.ID, &p.Name, &p.MimeType, &p.Size, &p.UserID, &p.StorageKey)
	if err != nil {
		return model.Photo{}, err
	}

	return p, nil
}

func (r *Repository) Delete(ID string) error {
	_, err := r.db.Exec("DELETE FROM meta WHERE user_id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
