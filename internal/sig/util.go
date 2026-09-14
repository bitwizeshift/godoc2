package sig

import (
	"iter"
	"strconv"
)

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

// enumerate pairs each value of seq with its index.
func enumerate[T any](seq iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for v := range seq {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}
