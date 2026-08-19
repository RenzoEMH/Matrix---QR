package matrix

import (
	"errors"
	"math"
	"testing"
)

const qrTol = 1e-8

func TestFactorizeQR_ReconstructsAndOrthogonal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    Matrix
	}{
		{name: "1x1", a: Matrix{{5}}},
		{name: "1x1 negative", a: Matrix{{-3}}},
		{name: "2x2", a: Matrix{{1, 2}, {3, 4}}},
		{name: "identity 3x3", a: Matrix{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}},
		{name: "3x2 tall", a: Matrix{{1, 2}, {3, 4}, {5, 6}}},
		{name: "2x3 wide", a: Matrix{{1, 2, 3}, {4, 5, 6}}},
		{name: "zeros", a: Matrix{{0, 0}, {0, 0}}},
		{name: "rank deficient", a: Matrix{{1, 2}, {2, 4}}},
		{name: "known 3x3", a: Matrix{
			{12, -51, 4},
			{6, 167, -68},
			{-4, 24, -41},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertValidQR(t, tt.a)
		})
	}
}

func TestFactorizeQR_DoesNotMutateInput(t *testing.T) {
	t.Parallel()

	in := Matrix{{1, 2}, {3, 4}, {5, 6}}
	snapshot := clone(in)
	if _, err := FactorizeQR(in); err != nil {
		t.Fatal(err)
	}
	if !AlmostEqual(in, snapshot, 0) {
		t.Fatalf("input mutated: got %v, want %v", in, snapshot)
	}
}

func TestFactorizeQR_RejectsInvalid(t *testing.T) {
	t.Parallel()

	if _, err := FactorizeQR(Matrix{{1, 2}, {3}}); !errors.Is(err, ErrNotRectangular) {
		t.Fatalf("error = %v, want %v", err, ErrNotRectangular)
	}
}

func TestFactorizeQR_ThinShapes(t *testing.T) {
	t.Parallel()

	tall := Matrix{{1, 2}, {3, 4}, {5, 6}}
	qr, err := FactorizeQR(tall)
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, "tall Q", qr.Q, 3, 2)
	assertShape(t, "tall R", qr.R, 2, 2)

	wide := Matrix{{1, 2, 3}, {4, 5, 6}}
	qr, err = FactorizeQR(wide)
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, "wide Q", qr.Q, 2, 2)
	assertShape(t, "wide R", qr.R, 2, 3)
}

func TestMatMul_AndTranspose(t *testing.T) {
	t.Parallel()

	a := Matrix{{1, 2}, {3, 4}}
	id := Matrix{{1, 0}, {0, 1}}

	got, err := MatMul(a, id)
	if err != nil {
		t.Fatal(err)
	}
	if !AlmostEqual(got, a, 0) {
		t.Fatalf("A I = %v, want %v", got, a)
	}

	if !AlmostEqual(Transpose(Transpose(a)), a, 0) {
		t.Fatal("transpose is not involutive")
	}
}

func TestMatMul_DimensionMismatch(t *testing.T) {
	t.Parallel()

	_, err := MatMul(Matrix{{1, 2}}, Matrix{{1}, {2}, {3}})
	if !errors.Is(err, ErrDimensionMismatch) {
		t.Fatalf("error = %v, want %v", err, ErrDimensionMismatch)
	}
}

func assertValidQR(t *testing.T, a Matrix) {
	t.Helper()

	qr, err := FactorizeQR(a)
	if err != nil {
		t.Fatalf("FactorizeQR() error = %v", err)
	}

	m, n := Dims(a)
	_, qCols := Dims(qr.Q)
	rRows, rCols := Dims(qr.R)

	if m >= n {
		assertShape(t, "Q", qr.Q, m, n)
		assertShape(t, "R", qr.R, n, n)
	} else {
		assertShape(t, "Q", qr.Q, m, m)
		assertShape(t, "R", qr.R, m, n)
	}

	product, err := MatMul(qr.Q, qr.R)
	if err != nil {
		t.Fatalf("Q R multiply: %v", err)
	}
	if !AlmostEqual(product, a, qrTol) {
		t.Fatalf("Q R != A\nQ R = %v\nA   = %v", product, a)
	}

	qtq, err := MatMul(Transpose(qr.Q), qr.Q)
	if err != nil {
		t.Fatalf("Qᵀ Q multiply: %v", err)
	}
	if !AlmostEqual(qtq, identity(qCols), qrTol) {
		t.Fatalf("Qᵀ Q != I: %v", qtq)
	}

	if !isUpperTriangular(qr.R, qrTol) {
		t.Fatalf("R is not upper triangular: %v", qr.R)
	}

	for j := 0; j < min(rRows, rCols); j++ {
		if qr.R[j][j] < -qrTol {
			t.Fatalf("R[%d %d] = %v, want >= 0", j, j, qr.R[j][j])
		}
	}
}

func assertShape(t *testing.T, name string, a Matrix, rows, cols int) {
	t.Helper()
	gotRows, gotCols := Dims(a)
	if gotRows != rows || gotCols != cols {
		t.Fatalf("%s shape = %d×%d, want %d×%d", name, gotRows, gotCols, rows, cols)
	}
}

func isUpperTriangular(a Matrix, eps float64) bool {
	rows, cols := Dims(a)
	for i := 1; i < rows; i++ {
		for j := 0; j < i && j < cols; j++ {
			if math.Abs(a[i][j]) > eps {
				return false
			}
		}
	}
	return true
}
