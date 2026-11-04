package elasticsearch

import (
	"testing"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSearchQuery_NoFiltersMatchesAll(t *testing.T) {
	q := buildSearchQuery(model.DoctorSearchParams{})
	assert.Equal(t, map[string]interface{}{"match_all": map[string]interface{}{}}, q["query"])
}

func TestBuildSearchQuery_FiltersAreCombinedWithAnd(t *testing.T) {
	q := buildSearchQuery(model.DoctorSearchParams{Specialty: "cardiologist", Sex: "F", Services: []string{"ecg", " "}})

	boolQuery, ok := q["query"].(map[string]interface{})["bool"].(map[string]interface{})
	require.True(t, ok)
	assert.NotContains(t, boolQuery, "should", "семантика OR расширяла бы результаты с каждым фильтром")

	must, ok := boolQuery["must"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, must, 3, "специальность, пол и одна непустая услуга")
	assert.Contains(t, must, map[string]interface{}{"term": map[string]interface{}{"sex": "F"}})
}
