package model

var DoctorsIndexMapping = map[string]interface{}{
	"mappings": map[string]interface{}{
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type": "keyword",
			},
			"sex": map[string]interface{}{
				"type": "keyword",
			},
			"first_name": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "keyword"},
				},
			},
			"middle_name": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "keyword"},
				},
			},
			"last_name": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "keyword"},
				},
			},
			"specialty": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "keyword"},
				},
			},
			"services": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "keyword"},
				},
			},
		},
	},
}
