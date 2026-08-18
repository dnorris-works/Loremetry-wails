import express from "express";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(fileURLToPath(import.meta.url));
const dist = join(root, "dist");
const port = Number(process.env.PORT || 4173);

const app = express();

app.use(express.static(dist));
app.get(/^(?!.*\.).*$/, (_req, res) => {
  res.sendFile(join(dist, "index.html"));
});

app.listen(port, () => {
  console.log(`loremetry website listening on ${port}`);
});
