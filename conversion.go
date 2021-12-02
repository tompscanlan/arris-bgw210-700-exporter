package main

import "math"

func samplesToIncrement(past, present int64) int64 {
	if present > past {
		return present - past
	} else {
		return (math.MaxInt64 - past) + (present - math.MinInt64 + 1)

	}
}
