import { spawn } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const port = process.env.PORT || "3000";
const serveBin = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  "node_modules",
  ".bin",
  process.platform === "win32" ? "serve.cmd" : "serve",
);

const child = spawn(
  serveBin,
  ["-s", "dist", "-l", `tcp://0.0.0.0:${port}`],
  { stdio: "inherit", shell: true },
);

function shutdown() {
  child.kill("SIGTERM");
}

process.on("SIGTERM", shutdown);
process.on("SIGINT", shutdown);
child.on("exit", (code) => process.exit(code ?? 1));
