import { formatNumber } from "../matrix";

export function StatsCards({ stats }) {
  if (!stats) {
    return null;
  }

  const items = [
    { label: "Máximo", value: formatNumber(stats.max) },
    { label: "Mínimo", value: formatNumber(stats.min) },
    { label: "Promedio", value: formatNumber(stats.average) },
    { label: "Suma", value: formatNumber(stats.sum) },
  ];

  const diagonal = stats.diagonal ?? {};

  return (
    <section className="stats-grid">
      {items.map((item) => (
        <article className="card stat-card" key={item.label}>
          <p className="muted">{item.label}</p>
          <p className="stat-value">{item.value}</p>
        </article>
      ))}
      {Object.entries(diagonal).map(([name, isDiag]) => (
        <article className="card stat-card" key={name}>
          <p className="muted">¿{name} diagonal?</p>
          <p className={isDiag ? "pill ok" : "pill no"}>{isDiag ? "Sí" : "No"}</p>
        </article>
      ))}
    </section>
  );
}
