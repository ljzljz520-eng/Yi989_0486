import { mkdir, writeFile } from "node:fs/promises";

await mkdir("dist", { recursive: true });
const page = `<!doctype html><html><head><meta charset="utf-8"><title>Coupon batches</title></head><body><main><h1>Coupon batches</h1><p>Storefront claim view</p></main></body></html>`;
await writeFile("dist/index.html", page);
