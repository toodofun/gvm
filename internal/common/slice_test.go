package common

import (
	"reflect"
	"testing"
)

func TestReverseSlice(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		s := []int{}
		ReverseSlice(s)
		expected := []int{}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})

	t.Run("single element", func(t *testing.T) {
		s := []string{"a"}
		ReverseSlice(s)
		expected := []string{"a"}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})

	t.Run("two elements", func(t *testing.T) {
		s := []int{1, 2}
		ReverseSlice(s)
		expected := []int{2, 1}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})

	t.Run("odd number of elements", func(t *testing.T) {
		s := []int{1, 2, 3}
		ReverseSlice(s)
		expected := []int{3, 2, 1}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})

	t.Run("even number of elements", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		ReverseSlice(s)
		expected := []int{4, 3, 2, 1}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})

	t.Run("float64 slice", func(t *testing.T) {
		s := []float64{1.1, 2.2, 3.3}
		ReverseSlice(s)
		expected := []float64{3.3, 2.2, 1.1}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})

	t.Run("bool slice", func(t *testing.T) {
		s := []bool{true, false, true}
		ReverseSlice(s)
		expected := []bool{true, false, true}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})

	t.Run("struct slice", func(t *testing.T) {
		type Point struct {
			X int
			Y int
		}
		s := []Point{{1, 2}, {3, 4}, {5, 6}}
		ReverseSlice(s)
		expected := []Point{{5, 6}, {3, 4}, {1, 2}}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected %v, got %v", expected, s)
		}
	})
}
