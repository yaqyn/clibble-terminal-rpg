import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { JSDOM } from "jsdom";
const html = await readFile(
  new URL("../internal/webui/assets/index.html", import.meta.url),
  "utf8",
);
const script = await readFile(
  new URL("../internal/webui/assets/app.js", import.meta.url),
  "utf8",
);
const creature = (name, type) => ({
  name,
  height: 3,
  weight: 18,
  types: [{ type: { name: type } }],
  stats: [{ base_stat: 40, stat: { name: "hp" } }],
});
async function setup(pokemon = []) {
  const dom = new JSDOM(html, {
    url: "http://127.0.0.1:8787/",
    runScripts: "outside-only",
  });
  let payload = pokemon,
    failed = false,
    calls = 0;
  dom.window.fetch = async () => {
    calls++;
    return { ok: !failed, json: async () => ({ pokemon: payload }) };
  };
  dom.window.eval(script);
  await tick();
  return {
    dom,
    document: dom.window.document,
    calls: () => calls,
    fail: () => {
      failed = true;
    },
    replace: (value) => {
      payload = value;
    },
  };
}
const tick = () => new Promise((resolve) => setTimeout(resolve, 20));
test("search by type, choose a record, clear, and recover from no results", async () => {
  const app = await setup([
    creature("pidgey", "flying"),
    creature("pikachu", "electric"),
  ]);
  const doc = app.document;
  assert.match(doc.querySelector("#record-title").textContent, /pidgey/);
  const buttons = doc.querySelectorAll(".creature");
  buttons[1].click();
  assert.equal(doc.querySelector("#record-title").textContent, "pikachu");
  assert.equal(buttons[1].getAttribute("aria-pressed"), "true");
  const search = doc.querySelector("#search");
  search.value = "flying";
  search.dispatchEvent(new app.dom.window.Event("input"));
  assert.equal(doc.querySelectorAll(".creature").length, 1);
  assert.equal(doc.querySelector("#count").textContent, "1 of 2");
  search.value = "missing";
  search.dispatchEvent(new app.dom.window.Event("input"));
  assert.equal(
    doc.querySelector("#empty-title").textContent,
    "No matching records.",
  );
  doc.querySelector("#clear").click();
  assert.equal(search.value, "");
  assert.equal(doc.activeElement, search);
  assert.equal(doc.querySelectorAll(".creature").length, 2);
  app.dom.window.close();
});
test("empty state provides a real next action", async () => {
  const app = await setup();
  assert.equal(app.document.querySelector("#empty").hidden, false);
  assert.match(
    app.document.querySelector("#empty code").textContent,
    /dexrove catch pidgey/,
  );
  assert.equal(app.document.querySelector("#search").disabled, true);
  app.dom.window.close();
});
test("refresh failures preserve readable records and enable retry", async () => {
  const app = await setup([creature("pidgey", "flying")]);
  app.fail();
  app.document.querySelector("#refresh").click();
  await tick();
  assert.match(
    app.document.querySelector("#status").textContent,
    /could not be read/,
  );
  assert.equal(
    app.document.querySelector("#record-title").textContent,
    "pidgey",
  );
  assert.equal(app.document.querySelector("#refresh").disabled, false);
  assert.equal(app.calls(), 2);
  app.dom.window.close();
});
test("refresh reconciles selection and renders text without HTML execution", async () => {
  const app = await setup([creature("pidgey", "flying")]);
  app.replace([creature("pikachu", "electric")]);
  app.document.querySelector("#refresh").click();
  await tick();
  assert.equal(
    app.document.querySelector("#record-title").textContent,
    "pikachu",
  );
  assert.equal(
    app.document.querySelector(".measurements dd").textContent,
    "0.3 m",
  );
  app.dom.window.close();
});

test("large collections use explicit load more", async () => {
  const app = await setup(
    Array.from({ length: 201 }, (_, index) =>
      creature(`pidgey-${index}`, "flying"),
    ),
  );
  assert.equal(app.document.querySelectorAll(".creature").length, 100);
  app.document.querySelector("#more").click();
  assert.equal(app.document.querySelectorAll(".creature").length, 200);
  app.document.querySelector("#more").click();
  assert.equal(app.document.querySelectorAll(".creature").length, 201);
  assert.equal(app.document.querySelector("#more").hidden, true);
  app.dom.window.close();
});
