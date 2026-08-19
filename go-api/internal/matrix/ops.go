package matrix

import (
	"fmt"
	"math"
)

// Transpose returns a new matrix with rows and columns swapped.
func Transpose(a Matrix) Matrix {
	rows, cols := Dims(a)
	out := make(Matrix, cols)
	for j := 0; j < cols; j++ {
		out[j] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			out[j][i] = a[i][j]
		}
	}
	return out
}

// MatMul returns the product a × b.
func MatMul(a, b Matrix) (Matrix, error) {
	if err := Validate(a); err != nil {
		return nil, fmt.Errorf("left matrix: %w", err)
	}
	if err := Validate(b); err != nil {
		return nil, fmt.Errorf("right matrix: %w", err)
	}

	aRows, aCols := Dims(a)
	bRows, bCols := Dims(b)
	if aCols != bRows {
		return nil, fmt.Errorf("%w: %d×%d times %d×%d", ErrDimensionMismatch, aRows, aCols, bRows, bCols)
	}

	out := make(Matrix, aRows)
	for i := 0; i < aRows; i++ {
		out[i] = make([]float64, bCols)
		for k := 0; k < aCols; k++ {
			aik := a[i][k]
			if aik == 0 {
				continue
			}
			for j := 0; j < bCols; j++ {
				out[i][j] += aik * b[k][j]
			}
		}
	}
	return out, nil
}

// AlmostEqual reports whether a and b have the same shape and every pair of
// entries differs by at most epsilon in absolute value.
func AlmostEqual(a, b Matrix, epsilon float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if math.Abs(a[i][j]-b[i][j]) > epsilon {
				return false
			}
		}
	}
	return true
}
