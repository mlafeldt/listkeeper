package main

import (
	"sort"

	mapset "github.com/deckarep/golang-set"
)

//nolint:forcetypeassert
// diffInt64Slices compares two int64 slices and returns any differences between them.
func diffInt64Slices(x, y []int64) (eq bool, xd, yd []int64) {
	xs := mapset.NewSet()
	for _, i := range x {
		xs.Add(i)
	}

	ys := mapset.NewSet()
	for _, i := range y {
		ys.Add(i)
	}

	if eq = xs.Equal(ys); eq {
		return
	}

	for v := range xs.Difference(ys).Iterator().C {
		xd = append(xd, v.(int64))
	}
	for v := range ys.Difference(xs).Iterator().C {
		yd = append(yd, v.(int64))
	}

	// Make results stable
	sort.Slice(xd, func(i, j int) bool { return xd[i] < xd[j] })
	sort.Slice(yd, func(i, j int) bool { return yd[i] < yd[j] })

	return
}
