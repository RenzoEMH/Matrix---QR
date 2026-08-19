const express = require("express");
const routes = require("./routes");
const { ValidationError } = require("./validation");

const app = express();

app.use(express.json({ limit: "1mb" }));

app.get("/health", (_req, res) => {
  res.json({ status: "ok", service: "node-api" });
});

app.use("/api/v1", routes);

app.use((_req, res) => {
  res.status(404).json({ error: "not found" });
});

app.use((err, _req, res, _next) => {
  if (err instanceof ValidationError) {
    return res.status(400).json({ error: err.message });
  }

  if (err instanceof SyntaxError && err.status === 400 && "body" in err) {
    return res.status(400).json({ error: "invalid json body" });
  }

  console.error(err);
  return res.status(500).json({ error: "internal server error" });
});

const port = Number(process.env.PORT || process.env.NODE_PORT) || 3000;

if (require.main === module) {
  app.listen(port, () => {
    console.log(`node-api listening on :${port}`);
  });
}

module.exports = app;
