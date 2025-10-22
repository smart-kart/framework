package generic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smart-kart/framework/response"
	"github.com/smart-kart/framework/utils/generic"
)

func TestReturnZero(t *testing.T) {
	type args struct {
		x any
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "Should return int zero value",
			args: args{5},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generic.ReturnZero(tt.args.x)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContains(t *testing.T) {
	type args struct {
		elems []int
		v     int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should return true if exists",
			args: args{[]int{1, 2, 3}, 2},
			want: true,
		},
		{
			name: "Should return false if not exists",
			args: args{[]int{1, 2, 3}, 10},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generic.Contains(tt.args.elems, tt.args.v)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		val   string
		want  []string
	}{
		{
			name:  "should able to remove if there are multiple occurrences of val",
			input: []string{"a", "b", "c", "a", "b"},
			val:   "a",
			want:  []string{"b", "c", "b"},
		},
		{
			name:  "should not remove any elem if there is no occurrence of val",
			input: []string{"a", "b", "c", "d"},
			val:   "e",
			want:  []string{"a", "b", "c", "d"},
		},
		{
			name:  "should return empty slice for the empty slice input",
			input: []string{},
			val:   "a",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		got := generic.Remove(tt.input, tt.val)
		assert.Equal(t, tt.want, got)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	type args struct {
		array1 []string
		array2 []int
	}
	tests := []struct {
		name  string
		args  args
		want1 []string
		want2 []int
	}{
		{
			name: "should able to remove if there are multiple occurrences of an element",
			args: args{
				array1: []string{"a", "b", "c", "a", "a", "c"},
				array2: []int{1, 2, 3, 1, 1},
			},
			want1: []string{"a", "b", "c"},
			want2: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1 := generic.RemoveDuplicates(tt.args.array1)
			got2 := generic.RemoveDuplicates(tt.args.array2)

			assert.Equal(t, tt.want1, got1)
			assert.Equal(t, tt.want2, got2)
		})
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		args any
		want bool
	}{
		{"should return true for nil", nil, true},
		{"should return true for response.Empty", response.Empty, true},
		{"should return true for nil interface", (any)(nil), true},
		{"should return true for nil pointer", (*struct{})(nil), true},
		{"should return false for non-nil pointer to empty struct", &struct{}{}, false},
		{"should return true for empty string", "", true},
		{"should return false for non-empty string", "non-empty", false},
		{"should return true for zero int", 0, true},
		{"should return false for non-zero int", 1, false},
		{"should return true for zero float", 0.0, true},
		{"should return false for non-zero float", 1.1, false},
		{"should return true for nil slice", []int(nil), true},
		{"should return false for non-empty slice", []int{1}, false},
		{"should return true for nil map", map[int]int(nil), true},
		{"should return false for non-empty map", map[int]int{1: 1}, false},
		{"should return true for zero value struct", struct{ a int }{}, true},
		{"should return false for non-zero value struct", struct{ a int }{a: 1}, false},
		{"should return false for true bool", true, false},
		{"should return true for false bool", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generic.IsZero(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}
