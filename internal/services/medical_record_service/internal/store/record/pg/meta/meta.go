package meta

import (
	"github.com/alexey-dobry/medical-service/internal/services/medical_record_service/internal/domain/model"
	"github.com/google/uuid"
)

func (r *Repository) Add(documentMeta model.DocumentMeta) error {
	_, err := r.db.Exec(
		"INSERT INTO meta (id,name,mime_type,size,record_id,storage_key) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		documentMeta.ID,
		documentMeta.Name,
		documentMeta.MimeType,
		documentMeta.Size,
		documentMeta.RecordID,
		documentMeta.StorageKey,
	)

	return err
}

func (r *Repository) Get(id uuid.UUID) (model.DocumentMeta, error) {
	row := r.db.QueryRow("SELECT id,name,mime_type,size,record_id,storage_key FROM meta WHERE id = $1", id)

	var p model.DocumentMeta
	err := row.Scan(&p.ID, &p.Name, &p.MimeType, &p.Size, &p.RecordID, &p.StorageKey)
	if err != nil {
		return model.DocumentMeta{}, err
	}

	return p, nil
}

func (r *Repository) Delete(id uuid.UUID) (string, error) {
	var key string
	err := r.db.QueryRow("DELETE FROM meta WHERE record_id = $1 RETURNING storage_key", id).Scan(key)
	if err != nil {
		return "", err
	}

	return key, nil
}
