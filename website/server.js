import express from "express";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(fileURLToPath(import.meta.url));
const dist = join(root, "dist");
const port = Number(process.env.PORT || 5000);
const host = "0.0.0.0";

const app = express();

app.get("/health", (_req, res) => {
  res.status(200).send("ok");
});

app.use(express.static(dist));
app.use((_req, res) => {
  res.sendFile(join(dist, "index.html"));
});

app.listen(port, host, () => {
  console.log(`loremetry website listening on ${host}:${port}`);
});
