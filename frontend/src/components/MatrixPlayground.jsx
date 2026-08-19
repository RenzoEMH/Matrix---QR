import { useEffect, useState } from "react";
import { ApiError, analyzeMatrix, checkHealth } from "../api";
import { parseMatrixJSON } from "../matrix";
import { MatrixTable } from "./MatrixTable";
import { StatsCards } from "./StatsCards";

const SAMPLE = "[[1, 2], [3, 4], [5, 6]]";

export function MatrixPlayground({ token, onLogout, onUnauthorized }) {
  const [raw, setRaw] = useState(SAMPLE);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [result, setResult] = useState(null);
  const [healthy, setHealthy] = useState(null);

  useEffect(() => {
    let cancelled = false;
    async function ping() {
      try {
        await checkHealth();
        if (!cancelled) setHealthy(true);
      } catch {
        if (!cancelled) setHealthy(false);
      }
    }
    ping();
    const id = setInterval(ping, 15000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  async function handleSubmit(event) {
    event.preventDefault();
    setError("");
    let matrix;
    try {
      matrix = parseMatrixJSON(raw);
    } catch (err) {
      setError(err.message);
      return;
    }

    setPending(true);
    try {
      const data = await analyzeMatrix(matrix, token);
      setResult(data);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        onUnauthorized();
        return;
      }
      setError(err.message);
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Matrix QR</p>
          <h1>Factorización y rotación</h1>
        </div>
        <div className="topbar-actions">
          <span className={healthy ? "status ok" : "status down"}>
            {healthy === null ? "API…" : healthy ? "API ok" : "API caída"}
          </span>
          <button type="button" className="ghost" onClick={onLogout}>
            Salir
          </button>
        </div>
      </header>

      <form className="card editor" onSubmit={handleSubmit}>
        <label>
          Matriz JSON
          <textarea
            value={raw}
            onChange={(e) => setRaw(e.target.value)}
            rows={8}
            spellCheck="false"
          />
        </label>
        {error ? <p className="error">{error}</p> : null}
        <button type="submit" disabled={pending}>
          {pending ? "Calculando…" : "Calcular"}
        </button>
      </form>

      {result ? (
        <>
          <StatsCards stats={result.stats} />
          <div className="tables">
            <MatrixTable title="Original" rows={result.original} />
            <MatrixTable title="Rotada 90°" rows={result.rotated} />
            <MatrixTable title="Q" rows={result.qr?.q} />
            <MatrixTable title="R" rows={result.qr?.r} />
          </div>
        </>
      ) : null}
    </div>
  );
}
