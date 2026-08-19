import { useState } from "react";
import { ApiError } from "../api";

export function LoginForm({ onLogin }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function handleSubmit(event) {
    event.preventDefault();
    setError("");
    setPending(true);
    try {
      await onLogin(username.trim(), password);
    } catch (err) {
      const message = err instanceof ApiError && err.status === 401
        ? "usuario o contraseña incorrectos"
        : err.message;
      setError(message);
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="auth-layout">
      <form className="card auth-card" onSubmit={handleSubmit}>
        <p className="eyebrow">Matrix QR</p>
        <h1>Iniciar sesión</h1>
        <label>
          Usuario
          <input
            autoComplete="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
        </label>
        <label>
          Contraseña
          <input
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>
        {error ? <p className="error">{error}</p> : null}
        <button type="submit" disabled={pending}>
          {pending ? "Entrando…" : "Entrar"}
        </button>
      </form>
    </main>
  );
}
