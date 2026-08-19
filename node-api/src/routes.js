const { Router } = require("express");
const { computeStats } = require("./stats");
const { parseMatrices } = require("./validation");

const router = Router();

router.post("/stats", (req, res, next) => {
  try {
    const matrices = parseMatrices(req.body);
    res.json(computeStats(matrices));
  } catch (err) {
    next(err);
  }
});

module.exports = router;
