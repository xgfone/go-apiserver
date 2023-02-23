// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

// IndexFunc returns the first index i satisfying equal(vs[i]), or -1.
func IndexFunc[S ~[]E, E any](vs S, equal func(E) bool) int {
	for i, e := range vs {
		if equal(e) {
			return i
		}
	}
	return -1
}
