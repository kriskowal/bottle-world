package sim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSumOfLifeAbout(t *testing.T) {
	var f Field
	f[0][0].Life = 1
	assert.Equal(t, 0, SumOfLifeAbout(f, IntVec2{0, 0}))
	assert.Equal(t, 1, SumOfLifeAbout(f, IntVec2{1, 1}))
	assert.Equal(t, 0, SumOfLifeAbout(f, IntVec2{2, 2}))
	assert.Equal(t, 1, SumOfLifeAbout(f, IntVec2{Width - 1, Height - 1}))
	assert.Equal(t, 0, SumOfLifeAbout(f, IntVec2{Width - 2, Height - 2}))
}
