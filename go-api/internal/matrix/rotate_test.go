package matrix

import (
	"errors"
	"math"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      Matrix
		wantErr error
	}{
		{name: "nil", in: nil, wantErr: ErrEmpty},
		{name: "no rows", in: Matrix{}, wantErr: ErrEmpty},
		{name: "no columns", in: Matrix{{}}, wantErr: ErrEmpty},
		{name: "jagged", in: Matrix{{1, 2}, {3}}, wantErr: ErrNotRectangular},
		{name: "NaN", in: Matrix{{math.NaN()}}, wantErr: ErrNonFinite},
		{name: "Inf", in: Matrix{{math.Inf(1)}}, wantErr: ErrNonFinite},
		{name: "ok", in: Matrix{{1}, {2}}, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.in)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRotate90Clockwise(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Matrix
		want Matrix
	}{
		{
			name: "1x1",
			in:   Matrix{{7}},
			want: Matrix{{7}},
		},
		{
			name: "2x2",
			in:   Matrix{{1, 2}, {3, 4}},
			want: Matrix{{3, 1}, {4, 2}},
		},
		{
			name: "3x2 tall",
			in:   Matrix{{1, 2}, {3, 4}, {5, 6}},
			want: Matrix{{5, 3, 1}, {6, 4, 2}},
		},
		{
			name: "2x3 wide",
			in:   Matrix{{1, 2, 3}, {4, 5, 6}},
			want: Matrix{{4, 1}, {5, 2}, {6, 3}},
		},
		{
			name: "identity 2x2",
			in:   Matrix{{1, 0}, {0, 1}},
			want: Matrix{{0, 1}, {1, 0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Rotate90Clockwise(tt.in)
			if err != nil {
				t.Fatalf("Rotate90Clockwise() error = %v", err)
			}
			if !AlmostEqual(got, tt.want, 0) {
				t.Fatalf("Rotate90Clockwise() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRotate90Clockwise_FourTurnsIsIdentity(t *testing.T) {
	t.Parallel()

	original := Matrix{{1, 2, 3}, {4, 5, 6}}
	cur := clone(original)
	for i := 0; i < 4; i++ {
		var err error
		cur, err = Rotate90Clockwise(cur)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !AlmostEqual(cur, original, 0) {
		t.Fatalf("after 4 rotations = %v, want %v", cur, original)
	}
}

func TestRotate90Clockwise_DoesNotMutateInput(t *testing.T) {
	t.Parallel()

	in := Matrix{{1, 2}, {3, 4}, {5, 6}}
	snapshot := clone(in)
	if _, err := Rotate90Clockwise(in); err != nil {
		t.Fatal(err)
	}
	if !AlmostEqual(in, snapshot, 0) {
		t.Fatalf("input mutated: got %v, want %v", in, snapshot)
	}
}

func TestRotate90Clockwise_RejectsInvalid(t *testing.T) {
	t.Parallel()

	if _, err := Rotate90Clockwise(nil); !errors.Is(err, ErrEmpty) {
		t.Fatalf("error = %v, want %v", err, ErrEmpty)
	}
}
