package matrix

// Rotate90Clockwise returns a new matrix rotated 90° clockwise.
// An m×n input becomes n×m. The original is not modified.
func Rotate90Clockwise(a Matrix) (Matrix, error) {
	if err := Validate(a); err != nil {
		return nil, err
	}

	rows, cols := Dims(a)
	out := make(Matrix, cols)
	for j := 0; j < cols; j++ {
		out[j] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			out[j][rows-1-i] = a[i][j]
		}
	}
	return out, nil
}
