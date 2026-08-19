const DIAGONAL_EPSILON = 1e-9;

function isDiagonal(matrix, epsilon = DIAGONAL_EPSILON) {
  const rows = matrix.length;
  const cols = matrix[0].length;
  if (rows !== cols) {
    return false;
  }

  for (let i = 0; i < rows; i++) {
    for (let j = 0; j < cols; j++) {
      if (i !== j && Math.abs(matrix[i][j]) >= epsilon) {
        return false;
      }
    }
  }
  return true;
}

function computeStats(matrices) {
  let min = Infinity;
  let max = -Infinity;
  let sum = 0;
  let count = 0;
  const diagonal = {};

  for (const [name, matrix] of Object.entries(matrices)) {
    diagonal[name] = isDiagonal(matrix);
    for (const row of matrix) {
      for (const value of row) {
        if (value < min) {
          min = value;
        }
        if (value > max) {
          max = value;
        }
        sum += value;
        count += 1;
      }
    }
  }

  return {
    max,
    min,
    average: sum / count,
    sum,
    diagonal,
  };
}

module.exports = { computeStats, isDiagonal, DIAGONAL_EPSILON };
