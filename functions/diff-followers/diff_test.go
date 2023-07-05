package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDiffInt64Slices(t *testing.T) {
	tests := []struct {
		x     []int64
		y     []int64
		left  []int64
		right []int64
		eq    bool
	}{
		{
			x:     nil,
			y:     nil,
			left:  nil,
			right: nil,
			eq:    true,
		},
		{
			x:     []int64{},
			y:     []int64{},
			left:  nil,
			right: nil,
			eq:    true,
		},
		{
			x:     []int64{123},
			y:     []int64{123},
			left:  nil,
			right: nil,
			eq:    true,
		},
		{
			x:     []int64{},
			y:     []int64{123},
			left:  nil,
			right: []int64{123},
			eq:    false,
		},
		{
			x:     []int64{456},
			y:     []int64{},
			left:  []int64{456},
			right: nil,
			eq:    false,
		},
		{
			x:     []int64{123, 456},
			y:     []int64{123},
			left:  []int64{456},
			right: nil,
			eq:    false,
		},
		{
			x:     []int64{456},
			y:     []int64{456, 789},
			left:  nil,
			right: []int64{789},
			eq:    false,
		},
		{
			x:     []int64{123, 456},
			y:     []int64{456, 789},
			left:  []int64{123},
			right: []int64{789},
			eq:    false,
		},
		{
			x:     []int64{123, 789},
			y:     []int64{123, 456, 789},
			left:  nil,
			right: []int64{456},
			eq:    false,
		},
		{
			x:     []int64{123, 456, 789},
			y:     []int64{321, 654, 987},
			left:  []int64{123, 456, 789},
			right: []int64{321, 654, 987},
			eq:    false,
		},
	}

	for _, test := range tests {
		eq, left, right := diffInt64Slices(test.x, test.y)

		if diff := cmp.Diff(test.eq, eq); diff != "" {
			t.Errorf("mismatch eq: %s", diff)
		}
		if diff := cmp.Diff(test.left, left); diff != "" {
			t.Errorf("mismatch left: %s", diff)
		}
		if diff := cmp.Diff(test.right, right); diff != "" {
			t.Errorf("mismatch right: %s", diff)
		}
	}
}
