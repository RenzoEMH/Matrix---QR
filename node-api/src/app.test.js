const request = require("supertest");
const app = require("./app");

describe("GET /health", () => {
  test("returns ok", async () => {
    const res = await request(app).get("/health");
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ status: "ok", service: "node-api" });
  });
});

describe("POST /api/v1/stats", () => {
  test("returns aggregated stats", async () => {
    const res = await request(app)
      .post("/api/v1/stats")
      .send({
        matrices: {
          q: [[1, 0], [0, 1]],
          r: [[2, 0], [0, 3]],
          rotated: [[4, 5], [6, 7]],
        },
      });

    expect(res.status).toBe(200);
    expect(res.body.max).toBe(7);
    expect(res.body.min).toBe(0);
    expect(res.body.sum).toBe(29);
    expect(res.body.average).toBeCloseTo(29 / 12);
    expect(res.body.diagonal).toEqual({ q: true, r: true, rotated: false });
  });

  test("rejects invalid json", async () => {
    const res = await request(app)
      .post("/api/v1/stats")
      .set("Content-Type", "application/json")
      .send("{");
    expect(res.status).toBe(400);
    expect(res.body.error).toBe("invalid json body");
  });

  test("rejects missing matrices", async () => {
    const res = await request(app).post("/api/v1/stats").send({});
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/matrices/);
  });

  test("rejects jagged matrix", async () => {
    const res = await request(app)
      .post("/api/v1/stats")
      .send({ matrices: { q: [[1, 2], [3]] } });
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/rectangular/);
  });

  test("rejects empty matrix", async () => {
    const res = await request(app)
      .post("/api/v1/stats")
      .send({ matrices: { q: [] } });
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/empty/);
  });

  test("rejects oversized matrix", async () => {
    const row = Array(51).fill(1);
    const matrix = Array.from({ length: 51 }, () => row);
    const res = await request(app)
      .post("/api/v1/stats")
      .send({ matrices: { q: matrix } });
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/50/);
  });

  test("unknown route is json 404", async () => {
    const res = await request(app).get("/nope");
    expect(res.status).toBe(404);
    expect(res.body).toEqual({ error: "not found" });
  });
});
