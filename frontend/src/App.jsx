import { useCallback, useState } from "react";
import { login as loginRequest } from "./api";
import { clearToken, getToken, setToken } from "./auth";
import { LoginForm } from "./components/LoginForm";
import { MatrixPlayground } from "./components/MatrixPlayground";

export default function App() {
  const [token, setTokenState] = useState(() => getToken());

  const handleLogin = useCallback(async (username, password) => {
    const data = await loginRequest(username, password);
    setToken(data.token);
    setTokenState(data.token);
  }, []);

  const handleLogout = useCallback(() => {
    clearToken();
    setTokenState(null);
  }, []);

  if (!token) {
    return <LoginForm onLogin={handleLogin} />;
  }

  return (
    <MatrixPlayground
      token={token}
      onLogout={handleLogout}
      onUnauthorized={handleLogout}
    />
  );
}
