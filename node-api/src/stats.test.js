const { computeStats, isDiagonal, DIAGONAL_EPSILON } = require("./stats");

describe("isDiagonal", () => {
  test("identity is diagonal", () => {
    expect(isDiagonal([[1, 0], [0, 1]])).toBe(true);
  });

  test("1x1 is diagonal", () => {
    expect(isDiagonal([[7]])).toBe(true);
  });

  test("non-square is never diagonal", () => {
    expect(isDiagonal([[1, 2, 3], [4, 5, 6]])).toBe(false);
  });

  test("off-diagonal above epsilon is not diagonal", () => {
    expect(isDiagonal([[1, 1], [0, 1]])).toBe(false);
  });

  test("tiny off-diagonal within epsilon is diagonal", () => {
    expect(isDiagonal([[1, DIAGONAL_EPSILON / 10], [0, 1]])).toBe(true);
  });

  test("zeros matrix is diagonal", () => {
    expect(isDiagonal([[0, 0], [0, 0]])).toBe(true);
  });
});

describe("computeStats", () => {
  test("aggregates every value and flags each matrix", () => {
    const got = computeStats({
      q: [[1, 0], [0, 1]],
      r: [[2, 3], [0, 4]],
      rotated: [[5, 6]],
    });

    expect(got.min).toBe(0);
    expect(got.max).toBe(6);
    expect(got.sum).toBe(22);
    expect(got.average).toBeCloseTo(22 / 10);
    expect(got.diagonal).toEqual({ q: true, r: false, rotated: false });
  });

  test("single matrix matches its entries", () => {
    const got = computeStats({
      rotated: [[1, 2], [3, 4], [5, 6]],
    });

    expect(got).toEqual({
      max: 6,
      min: 1,
      average: 3.5,
      sum: 21,
      diagonal: { rotated: false },
    });
  });
});
