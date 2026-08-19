package matrix

import "math"

type reflector struct {
	row0 int
	v    []float64
	beta float64
}

// FactorizeQR computes a thin QR factorization of A using Householder
// reflections. The input is not modified.
func FactorizeQR(a Matrix) (QR, error) {
	if err := Validate(a); err != nil {
		return QR{}, err
	}

	m, n := Dims(a)
	work := clone(a)
	reflectors := make([]reflector, 0, min(m, n))

	for k := 0; k < min(m, n); k++ {
		v, beta := householder(columnTail(work, k, k))
		applyHouseholderLeft(work, k, k, v, beta)
		reflectors = append(reflectors, reflector{row0: k, v: v, beta: beta})
	}

	qFull := identity(m)
	for i := len(reflectors) - 1; i >= 0; i-- {
		r := reflectors[i]
		applyHouseholderLeft(qFull, r.row0, 0, r.v, r.beta)
	}

	qr := extractThin(qFull, work, m, n)
	zeroStrictLower(qr.R)
	normalizeDiagonalSigns(qr.Q, qr.R)
	return qr, nil
}

// householder builds a unit reflector P = I - beta v vᵀ that zeros x[1:].
// The sign of v[0] is chosen to avoid cancellation (Golub–Van Loan).
func householder(x []float64) (v []float64, beta float64) {
	v = append([]float64(nil), x...)
	sigma := norm2(x)
	if sigma == 0 {
		return v, 0
	}

	v[0] += math.Copysign(sigma, x[0])
	vn := norm2(v)
	if vn == 0 {
		return v, 0
	}
	for i := range v {
		v[i] /= vn
	}
	return v, 2
}

func applyHouseholderLeft(a Matrix, row0, col0 int, v []float64, beta float64) {
	if beta == 0 {
		return
	}
	cols := len(a[0])
	for j := col0; j < cols; j++ {
		dot := 0.0
		for i, vi := range v {
			dot += vi * a[row0+i][j]
		}
		scale := beta * dot
		for i, vi := range v {
			a[row0+i][j] -= scale * vi
		}
	}
}

func extractThin(qFull, rFull Matrix, m, n int) QR {
	if m >= n {
		return QR{Q: takeColumns(qFull, n), R: takeRows(rFull, n)}
	}
	return QR{Q: qFull, R: clone(rFull)}
}

func takeColumns(a Matrix, n int) Matrix {
	out := make(Matrix, len(a))
	for i, row := range a {
		out[i] = append([]float64(nil), row[:n]...)
	}
	return out
}

func takeRows(a Matrix, n int) Matrix {
	out := make(Matrix, n)
	for i := 0; i < n; i++ {
		out[i] = append([]float64(nil), a[i]...)
	}
	return out
}

func zeroStrictLower(a Matrix) {
	rows, cols := Dims(a)
	for i := 1; i < rows; i++ {
		limit := min(i, cols)
		for j := 0; j < limit; j++ {
			a[i][j] = 0
		}
	}
}

// normalizeDiagonalSigns flips column j of Q and row j of R when R[j,j] < 0
// so the conventional non-negative diagonal does not change the product Q R.
func normalizeDiagonalSigns(q, r Matrix) {
	_, qCols := Dims(q)
	rRows, rCols := Dims(r)
	for j := 0; j < min(qCols, min(rRows, rCols)); j++ {
		if r[j][j] >= 0 {
			continue
		}
		for k := 0; k < rCols; k++ {
			r[j][k] = -r[j][k]
		}
		for i := 0; i < len(q); i++ {
			q[i][j] = -q[i][j]
		}
	}
}

func columnTail(a Matrix, row0, col int) []float64 {
	x := make([]float64, len(a)-row0)
	for i := row0; i < len(a); i++ {
		x[i-row0] = a[i][col]
	}
	return x
}

func norm2(x []float64) float64 {
	n := 0.0
	for _, v := range x {
		n = math.Hypot(n, v)
	}
	return n
}
