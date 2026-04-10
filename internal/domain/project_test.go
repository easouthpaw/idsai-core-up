package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetakePenaltyPercent(t *testing.T) {
	require.Equal(t, 0, RetakePenaltyPercent(0))
	require.Equal(t, 5, RetakePenaltyPercent(1))
	require.Equal(t, 25, RetakePenaltyPercent(10))
}

func TestReviewScoreFromPercent(t *testing.T) {
	require.Equal(t, "0.0", ReviewScoreFromPercent(-10))
	require.Equal(t, "2.5", ReviewScoreFromPercent(50))
	require.Equal(t, "5.0", ReviewScoreFromPercent(999))
}
