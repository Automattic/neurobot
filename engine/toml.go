package engine

import (
	"strconv"
)

type WorkflowStepTOML struct {
	Active      bool
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intSliceToStringSlice(a []uint64) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.FormatUint(v, 10)
	}

	return b
}
