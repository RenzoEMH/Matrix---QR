import { formatNumber } from "../matrix";

export function MatrixTable({ title, rows }) {
  if (!rows?.length) {
    return null;
  }
  return (
    <section className="card table-card">
      <h2>{title}</h2>
      <div className="table-wrap">
        <table>
          <tbody>
            {rows.map((row, i) => (
              <tr key={i}>
                {row.map((cell, j) => (
                  <td key={j}>{formatNumber(cell)}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
