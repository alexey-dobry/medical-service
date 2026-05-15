package record

import (
	"github.com/alexey-dobry/medical-service/internal/services/medical_record_service/internal/domain/model"
	"github.com/google/uuid"
)

func (r *Repository) Add(medicalRecord model.MedicalRecord) error {
	_, err := r.db.Exec(
		"INSERT INTO medical_record (id,patient_id,doctor_id,type,conclusion,description,recommendations,date) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)",
		medicalRecord.ID,
		medicalRecord.PatientID,
		medicalRecord.DoctorID,
		medicalRecord.Type,
		medicalRecord.Conclusion,
		medicalRecord.Description,
		medicalRecord.Recommendations,
		medicalRecord.Date,
	)

	return err
}

func (r *Repository) GetMany(patientID uuid.UUID, limit, offset int) ([]model.MedicalRecord, error) {
	rows, err := r.db.Query(
		`SELECT id, patient_id, doctor_id, type, conclusion, date
			FROM medical_record
			WHERE patient_id = $1
			ORDER BY date DESC
			LIMIT $2 OFFSET $3`,
		patientID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]model.MedicalRecord, 0, limit)
	for rows.Next() {
		var p model.MedicalRecord
		if err := rows.Scan(
			&p.ID,
			&p.PatientID,
			&p.DoctorID,
			&p.Type,
			&p.Conclusion,
			&p.Date,
		); err != nil {
			return nil, err
		}
		records = append(records, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) GetOne(id uuid.UUID) (model.MedicalRecord, error) {
	rows, err := r.db.Query(
		`SELECT patient_id, doctor_id, type, conclusion, description, recommendations, date
			FROM medical_record
			WHERE id = $1`,
		id,
	)
	if err != nil {
		return model.MedicalRecord{}, err
	}
	defer rows.Close()

	var p model.MedicalRecord
	if err := rows.Scan(
		&p.ID,
		&p.PatientID,
		&p.DoctorID,
		&p.Type,
		&p.Conclusion,
		&p.Description,
		&p.Recommendations,
		&p.Date,
	); err != nil {
		return model.MedicalRecord{}, err
	}

	if err := rows.Err(); err != nil {
		return model.MedicalRecord{}, err
	}

	return p, nil
}

func (r *Repository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM medical_record WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}
