package engine

import "testing"

func TestSplitStringIntoSliceOfInts(t *testing.T) {
	tables := []struct {
		stringToSplit string
		sep           string
		res           []uint64
	}{
		// test basic functionality
		{
			"100,200",
			",",
			[]uint64{100, 200},
		},
		// test a different separator
		{
			"11-22-33",
			"-",
			[]uint64{11, 22, 33},
		},
		// test with unwanted empty spaces
		{
			" 1 ,2, 3,4 ",
			",",
			[]uint64{1, 2, 3, 4},
		},
		// test invalid input
		{
			"91,92,",
			",",
			[]uint64{91, 92},
		},
	}

	for _, table := range tables {
		got := splitStringIntoSliceOfInts(table.stringToSplit, table.sep)
		if !sliceEquals(table.res, got) {
			t.Errorf("slice of Ints didn't match. got:%v expected:%v", got, table.res)
		}
	}
}

func sliceEquals(a []uint64, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
