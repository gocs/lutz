package lutz

import (
	"slices"
	"testing"
)

func Test_insert(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		sorted []string
		line   string
		want   []string
	}{
		{
			name:   "insert into empty slice",
			sorted: []string{},
			line:   "c",
			want:   []string{"c"},
		},
		{
			name:   "insert into non-empty slice",
			sorted: []string{"c"},
			line:   "f",
			want:   []string{"c", "f"},
		},
		{
			name:   "insert into sorted slice",
			sorted: []string{"c", "f"},
			line:   "d",
			want:   []string{"c", "d", "f"},
		},
		{
			name:   "insert infront of a sorted slice",
			sorted: []string{"c", "d", "f"},
			line:   "b",
			want:   []string{"b", "c", "d", "f"},
		},
		{
			name:   "insert at the end of a sorted slice",
			sorted: []string{"b", "c", "d", "f"},
			line:   "i",
			want:   []string{"b", "c", "d", "f", "i"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insert(tt.sorted, tt.line)
			if !slices.Equal(got, tt.want) {
				t.Errorf("testing: [%s] insert() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
