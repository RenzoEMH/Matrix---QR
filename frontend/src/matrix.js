export function parseMatrixJSON(raw) {
  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch {
    throw new Error("JSON inválido");
  }
  if (!Array.isArray(parsed) || parsed.length === 0) {
    throw new Error("la matriz debe ser un array de filas");
  }
  const cols = Array.isArray(parsed[0]) ? parsed[0].length : -1;
  if (cols < 1) {
    throw new Error("cada fila debe ser un array de números");
  }
  for (const row of parsed) {
    if (!Array.isArray(row) || row.length !== cols) {
      throw new Error("la matriz no es rectangular");
    }
    for (const cell of row) {
      if (typeof cell !== "number" || !Number.isFinite(cell)) {
        throw new Error("solo se permiten números finitos");
      }
    }
  }
  return parsed;
}

export function formatNumber(value) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "—";
  }
  if (Number.isInteger(value)) {
    return String(value);
  }
  return value.toLocaleString("en-US", {
    maximumFractionDigits: 6,
    useGrouping: false,
  });
}
