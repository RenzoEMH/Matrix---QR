package matrix

import (
	"errors"
	"fmt"
	"math"
)

// Matrix is a rectangular numeric matrix stored row-major.
type Matrix [][]float64

// QR holds a thin factorization A = Q × R.
//
// If A is m×n and m >= n, Q is m×n (orthonormal columns) and R is n×n.
// If m < n, Q is m×m and R is m×n.
type QR struct {
	Q Matrix
	R Matrix
}

var (
	ErrEmpty             = errors.New("matrix is empty")
	ErrNotRectangular    = errors.New("matrix is not rectangular")
	ErrNonFinite         = errors.New("matrix contains NaN or Inf")
	ErrDimensionMismatch = errors.New("matrix dimension mismatch")
)

// Dims returns the number of rows and columns. Jagged matrices are not detected.
func Dims(a Matrix) (rows, cols int) {
	if len(a) == 0 {
		return 0, 0
	}
	return len(a), len(a[0])
}

// Validate reports whether a is a non-empty rectangular matrix of finite numbers.
func Validate(a Matrix) error {
	if len(a) == 0 || len(a[0]) == 0 {
		return ErrEmpty
	}

	cols := len(a[0])
	for i, row := range a {
		if len(row) != cols {
			return fmt.Errorf("%w: row %d has %d columns, want %d", ErrNotRectangular, i, len(row), cols)
		}
		for j, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("%w at [%d %d]", ErrNonFinite, i, j)
			}
		}
	}
	return nil
}

func clone(a Matrix) Matrix {
	out := make(Matrix, len(a))
	for i, row := range a {
		out[i] = append([]float64(nil), row...)
	}
	return out
}

func identity(n int) Matrix {
	out := make(Matrix, n)
	for i := range out {
		out[i] = make([]float64, n)
		out[i][i] = 1
	}
	return out
}
