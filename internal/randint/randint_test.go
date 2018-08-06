package randint

import "testing"

func TestNewRandomIntString(t *testing.T) {
	lenghts := []int{-1, 0, 1, 5, 20}

	for _, l := range lenghts {
		expectedLength := l
		if expectedLength < 0 {
			expectedLength = 0
		}

		v := NewRandomIntString(l)
		if len(v) != expectedLength {
			t.Errorf("Length should be %d but got %d", expectedLength, len(v))
		}
	}
}
