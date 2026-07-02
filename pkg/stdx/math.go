package stdx

import "cmp"

func Clamp[T cmp.Ordered](v, a, b T) T {
	if v < a {
		return a
	}
	if v > b {
		return b
	}
	return v
}
