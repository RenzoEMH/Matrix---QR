class ValidationError extends Error {
  constructor(message) {
    super(message);
    this.name = "ValidationError";
  }
}

const MAX_DIM = 50;
const MAX_MATRICES = 16;

function parseMatrices(body) {
  if (body == null || typeof body !== "object" || Array.isArray(body)) {
    throw new ValidationError("body must be a JSON object");
  }

  const { matrices } = body;
  if (matrices == null || typeof matrices !== "object" || Array.isArray(matrices)) {
    throw new ValidationError("matrices must be an object");
  }

  const names = Object.keys(matrices);
  if (names.length === 0) {
    throw new ValidationError("matrices must not be empty");
  }
  if (names.length > MAX_MATRICES) {
    throw new ValidationError(`at most ${MAX_MATRICES} matrices allowed`);
  }

  const out = {};
  for (const name of names) {
    out[name] = validateMatrix(name, matrices[name]);
  }
  return out;
}

function validateMatrix(name, matrix) {
  if (!Array.isArray(matrix) || matrix.length === 0) {
    throw new ValidationError(`matrix "${name}" is empty`);
  }

  if (!Array.isArray(matrix[0]) || matrix[0].length === 0) {
    throw new ValidationError(`matrix "${name}" is empty`);
  }

  const rows = matrix.length;
  const cols = matrix[0].length;
  if (rows > MAX_DIM || cols > MAX_DIM) {
    throw new ValidationError(`matrix "${name}" exceeds ${MAX_DIM}×${MAX_DIM} limit`);
  }

  for (let i = 0; i < rows; i++) {
    const row = matrix[i];
    if (!Array.isArray(row) || row.length !== cols) {
      throw new ValidationError(`matrix "${name}" is not rectangular`);
    }
    for (let j = 0; j < cols; j++) {
      if (!isFiniteNumber(row[j])) {
        throw new ValidationError(`matrix "${name}" contains a non-finite number at [${i} ${j}]`);
      }
    }
  }

  return matrix;
}

function isFiniteNumber(value) {
  return typeof value === "number" && Number.isFinite(value);
}

module.exports = { ValidationError, parseMatrices, MAX_DIM, MAX_MATRICES };
