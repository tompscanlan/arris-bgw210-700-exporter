package main

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSamplesToIncrement64(t *testing.T) {

	samples := []int64{2, 1, 0, -1, -2, math.MaxInt64 - 1, math.MaxInt64, -math.MaxInt64}

	for _, past := range samples {

		fmt.Printf("%d,\t%d\t->\t\t%d\n", past, past+1, samplesToIncrement(past, past+1))
		assert.Equal(t, int64(1), samplesToIncrement(past, past+1), fmt.Sprintf("failed for base sample %d", past))
	}

}
