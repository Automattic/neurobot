package engine

import (
	"reflect"
	"testing"
)

func TestBoolToInt(t *testing.T) {
	tables := []struct {
		input  bool
		output int
	}{
		{
			input:  true,
			output: 1,
		},
		{
			input:  false,
			output: 0,
		},
	}

	for _, table := range tables {
		if table.output != boolToInt(table.input) {
			t.Errorf("boolToInt() failing")
		}
	}
}

func TestAsSha256(t *testing.T) {
	tables := []struct {
		input interface{}
		hash  string
	}{
		{
			input: 1,
			hash:  "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
		},
		{
			// Simple slice of workflow step, variations will follow
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "8bcf9ee916ac1c78319c410fbf8a1b8523ee3f6f613612501651e80073e943c9",
		},
		{
			// Active status is changed for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      false,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "e250badbf5822c4e2bd6cd5a2ed6a0d26e6a2ca79c31fe4c7e6de0802349c2cf",
		},
		{
			// Name is changed for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test My Workflow",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "658e9e3cd92e1c9677327fe60e7dbbef5801662b815261c50882b0eab6480045",
		},
		{
			// Description is changed for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is a different description",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "9af2b40f7be6929f5a6fa2126e52e0508d6f87b30475e2a12a60ec85c0b18805",
		},
		{
			// Different variety of a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is a different description",
						Variety:     "stdout",
						Meta:        map[string]string{},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "d5e01e68bdace2ffa0efc67a19de9baceeef0a5eca9f9cda6777c4dd71763876",
		},
		{
			// Meta value (diff value for a particular meta) for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "newvalue",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "56500e217ab3d583400b86304d1f183851c5b855d020d20c0429f2564a874ede",
		},
		{
			// New meta value in a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key3": "value3",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "55f1e881fa352e47779b6dc74ed01261bdbf2116eddadaa1daf0d398a8990698",
		},
		{
			// Different count of workflow steps
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "5b99e2e9b67ce4284628837aa55485ff783356cef5a5e8eb5da0ac6f2f327ae0",
		},
	}

	for _, table := range tables {
		got := asSha256(table.input)
		if got != table.hash {
			t.Errorf("asSha256 hash didn't match. Got: %s Expected: %s", got, table.hash)
		}
	}
}

func TestIntSliceToStringSlice(t *testing.T) {
	tables := []struct {
		intSlice    []uint64
		stringslice []string
	}{
		{
			intSlice:    []uint64{},
			stringslice: []string{},
		},
		{
			intSlice:    []uint64{1},
			stringslice: []string{"1"},
		},
		{
			intSlice:    []uint64{1, 2},
			stringslice: []string{"1", "2"},
		},
		{
			intSlice:    []uint64{1, 2, 3, 4},
			stringslice: []string{"1", "2", "3", "4"},
		},
	}

	for _, table := range tables {
		got := intSliceToStringSlice(table.intSlice)
		if !reflect.DeepEqual(got, table.stringslice) {
			t.Errorf("intSliceToStringSlice didn't generate the correct string slice. Got: %s Expected: %s", got, table.stringslice)
		}
	}
}
