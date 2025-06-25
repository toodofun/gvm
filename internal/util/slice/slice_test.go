// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http:www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slice

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
