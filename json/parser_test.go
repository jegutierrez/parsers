package main

import (
	"reflect"
	"testing"
)

// TestJSONParse is a functional test of the parser.
func TestJSONParse(t *testing.T) {
	tt := []struct {
		name       string
		stringJSON string
		parsedJSON interface{}
		err        error
	}{
		{
			name:       "simple json",
			stringJSON: `{"key": "value"}`,
			parsedJSON: map[string]interface{}{"key": "value"},
			err:        nil,
		},
		{
			name:       "simple json",
			stringJSON: `{"key": value"}`,
			parsedJSON: map[string]interface{}{},
			err:        ErrorInvalidJSON,
		},
		{
			name:       "fun json",
			stringJSON: `{"value":1.2,"bool":true,"array":[1,2,5,7]}`,
			parsedJSON: map[string]interface{}{
				"value": float64(1.2),
				"bool":  true,
				"array": []interface{}{
					int64(1), int64(2), int64(5), int64(7),
				},
			},
			err: nil,
		},
		{
			name:       "fun json array",
			stringJSON: `[{"value":1.2,"bool":true,"array":[1,2,5,7]},{"value":1.2,"bool":true,"array":[1,2,5,7]}]`,
			parsedJSON: []interface{}{
				map[string]interface{}{
					"value": float64(1.2),
					"bool":  true,
					"array": []interface{}{
						int64(1), int64(2), int64(5), int64(7),
					},
				},
				map[string]interface{}{
					"value": float64(1.2),
					"bool":  true,
					"array": []interface{}{
						int64(1), int64(2), int64(5), int64(7),
					},
				},
			},
			err: nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := ParseJSON(tc.stringJSON)
			if tc.err == nil && !reflect.DeepEqual(res, tc.parsedJSON) {
				t.Errorf("unexpected response, want: %+v, got: %+v", tc.parsedJSON, res)
			}
			if tc.err != nil && tc.err != err {
				t.Errorf("unexpected error returned, want: %+v, got: %+v", tc.err, err)
			}
		})
	}
}
