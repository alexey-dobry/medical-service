package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/google/uuid"
)

func (r *Repository) AddDoctor(doctorData model.DoctorSearchParams) error {
	doc := map[string]interface{}{
		"id":          doctorData.ID,
		"first_name":  doctorData.FirstName,
		"middle_name": doctorData.MiddleName,
		"last_name":   doctorData.LastName,
		"sex":         doctorData.Sex,
		"specialty":   doctorData.Specialty,
		"service":     doctorData.Service,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("addDoctor: marshal document: %w", err)
	}

	res, err := r.db.Index(
		r.index,
		bytes.NewReader(body),
		r.db.Index.WithDocumentID(doctorData.ID),
		r.db.Index.WithContext(context.Background()),
		r.db.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("addDoctor: index request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("addDoctor: elasticsearch error [%s]", res.Status())
	}

	return nil
}

func (r *Repository) SearchDoctor(searchParams model.DoctorSearchParams) (uuid.UUID, error) {
	query := buildSearchQuery(searchParams)

	body, err := json.Marshal(query)
	if err != nil {
		return uuid.Nil, fmt.Errorf("searchDoctor: marshal query: %w", err)
	}

	res, err := r.db.Search(
		r.db.Search.WithIndex(r.index),
		r.db.Search.WithBody(bytes.NewReader(body)),
		r.db.Search.WithSize(1),
		r.db.Search.WithContext(context.Background()),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("searchDoctor: search request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return uuid.Nil, fmt.Errorf("searchDoctor: elasticsearch error [%s]", res.Status())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source struct {
					ID string `json:"id"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("searchDoctor: decode response: %w", err)
	}

	if len(result.Hits.Hits) == 0 {
		return uuid.Nil, fmt.Errorf("searchDoctor: no results found")
	}

	id, err := uuid.Parse(result.Hits.Hits[0].Source.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("searchDoctor: parse uuid %q: %w", result.Hits.Hits[0].Source.ID, err)
	}

	return id, nil
}

func buildSearchQuery(p model.DoctorSearchParams) map[string]interface{} {
	var should []map[string]interface{}

	addTerm := func(field, value string) {
		if strings.TrimSpace(value) != "" {
			should = append(should, map[string]interface{}{
				"term": map[string]interface{}{field: value},
			})
		}
	}

	addMatch := func(field, value string) {
		if strings.TrimSpace(value) != "" {
			should = append(should, map[string]interface{}{
				"match": map[string]interface{}{
					field: map[string]interface{}{
						"query":     value,
						"fuzziness": "AUTO",
					},
				},
			})
		}
	}

	addTerm("id", p.ID)
	addTerm("sex", p.Sex)
	addMatch("first_name", p.FirstName)
	addMatch("middle_name", p.MiddleName)
	addMatch("last_name", p.LastName)
	addMatch("specialty", p.Specialty)
	addMatch("service", p.Service)

	if len(should) == 0 {
		return map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               should,
				"minimum_should_match": 1,
			},
		},
	}
}
