"use strict";

const supportedMethods = ["get", "post", "put", "patch", "delete", "head", "options", "trace"];
const state = { operations: [] };

function text(value, fallback) {
  return typeof value === "string" && value.trim() ? value.trim() : fallback;
}

function collect(document) {
  const operations = [];
  for (const [path, item] of Object.entries(document.paths || {})) {
    for (const method of supportedMethods) {
      const operation = item && item[method];
      if (!operation || typeof operation !== "object") continue;
      operations.push({
        method: method.toUpperCase(),
        path,
        summary: text(operation.summary, "No summary provided"),
        operationId: text(operation.operationId, "Unidentified operation")
      });
    }
  }
  return operations.sort((left, right) => left.path.localeCompare(right.path) || left.method.localeCompare(right.method));
}

function render(query) {
  const normalized = query.trim().toLowerCase();
  const target = document.getElementById("operations");
  const template = document.getElementById("operation-template");
  target.replaceChildren();
  const visible = state.operations.filter((entry) =>
    !normalized || `${entry.method} ${entry.path} ${entry.summary} ${entry.operationId}`.toLowerCase().includes(normalized));
  for (const entry of visible) {
    const fragment = template.content.cloneNode(true);
    fragment.querySelector("article").dataset.method = entry.method;
    fragment.querySelector(".method").textContent = entry.method;
    fragment.querySelector(".path").textContent = entry.path;
    fragment.querySelector(".summary").textContent = entry.summary;
    fragment.querySelector(".operation-id").textContent = entry.operationId;
    target.append(fragment);
  }
  document.getElementById("status").textContent = `${visible.length} of ${state.operations.length} operations`;
}

async function load() {
  const response = await fetch("/swagger/openapi.json", { credentials: "same-origin", cache: "no-store" });
  if (!response.ok) throw new Error("OpenAPI contract is unavailable");
  const contract = await response.json();
  document.getElementById("title").textContent = text(contract.info && contract.info.title, "Harpy API");
  document.getElementById("version").textContent = `OpenAPI ${text(contract.openapi, "unknown")} · ${text(contract.info && contract.info.version, "unknown release")}`;
  state.operations = collect(contract);
  render("");
}

document.getElementById("search").addEventListener("input", (event) => render(event.target.value));
load().catch((error) => { document.getElementById("status").textContent = error.message; });
