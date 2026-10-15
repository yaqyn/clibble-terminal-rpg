"use strict";
const byId = (id) => document.getElementById(id);
const ui = {
  search: byId("search"),
  clear: byId("clear"),
  records: byId("records"),
  more: byId("more"),
  detail: byId("detail"),
  count: byId("count"),
  status: byId("status"),
  refresh: byId("refresh"),
  empty: byId("empty"),
};
const state = {
  pokemon: [],
  selected: null,
  loaded: false,
  busy: false,
  limit: 100,
};
const node = (tag, className, text) => {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (text !== undefined) element.textContent = text;
  return element;
};
const types = (pokemon) =>
  (pokemon.types || []).map((entry) => entry.type.name);
function visiblewildling() {
  const query = ui.search.value.trim().toLowerCase();
  return state.pokemon.filter(
    (pokemon) =>
      pokemon.name.includes(query) ||
      types(pokemon).some((type) => type.includes(query)),
  );
}
function select(pokemon) {
  state.selected = pokemon.name;
  for (const button of ui.records.children)
    button.setAttribute(
      "aria-pressed",
      String(button.dataset.name === pokemon.name),
    );
  renderDetail(pokemon);
}
function renderDetail(pokemon) {
  const header = node("div", "record-header");
  const introduction = node("div");
  introduction.append(
    node("p", "record-label", "Caught and saved"),
    node("h2", "", pokemon.name),
  );
  introduction.lastChild.id = "record-title";
  const badges = node("div", "types");
  types(pokemon).forEach((type) => badges.append(node("span", "type", type)));
  introduction.append(badges);
  const initial = node("div", "record-initial", pokemon.name[0].toUpperCase());
  initial.setAttribute("aria-hidden", "true");
  header.append(introduction, initial);
  const measurements = node("dl", "measurements");
  for (const [label, value] of [
    ["Height", `${(pokemon.height / 10).toFixed(1)} m`],
    ["Weight", `${(pokemon.weight / 10).toFixed(1)} kg`],
  ]) {
    const pair = node("div");
    pair.append(node("dt", "", label), node("dd", "", value));
    measurements.append(pair);
  }
  const stats = node("div", "stats");
  stats.append(node("h3", "stats-title", "Traits"));
  for (const entry of pokemon.stats || []) {
    const row = node("div", "stat");
    const label = entry.stat.name.replaceAll("-", " ");
    const progress = node("progress");
    progress.max = 255;
    progress.value = entry.base_stat;
    progress.setAttribute("aria-label", `${label}: ${entry.base_stat}`);
    row.append(
      node("span", "stat-label", label),
      node("span", "stat-value", entry.base_stat),
      progress,
    );
    stats.append(row);
  }
  ui.detail.replaceChildren(
    header,
    measurements,
    stats,
    node(
      "p",
      "record-note",
      "Stored locally. You can inspect this record without a network connection.",
    ),
  );
}
function renderList() {
  const matches = visiblewildling();
  const visible = matches.slice(0, state.limit);
  ui.more.hidden = visible.length >= matches.length;
  ui.more.disabled = ui.more.hidden;
  ui.more.textContent = `Show ${Math.min(100, matches.length - visible.length)} more`;
  ui.clear.hidden = ui.search.value.length === 0;
  ui.clear.disabled = ui.clear.hidden;
  ui.records.replaceChildren();
  ui.count.textContent = ui.search.value
    ? `${visible.length} of ${state.pokemon.length}`
    : `${state.pokemon.length} caught`;
  ui.empty.hidden = visible.length > 0;
  if (!visible.length) {
    const isEmpty = state.pokemon.length === 0;
    byId("empty-title").textContent = isEmpty
      ? "A new trail starts here."
      : "No matching records.";
    byId("empty-copy").textContent = isEmpty
      ? "Catch your first wildling in the terminal, then refresh this page."
      : "Try another name or habitat, or clear the search.";
    ui.empty.querySelector("code").hidden = !isEmpty;
    const placeholder = node("div", "record-placeholder");
    const heading = node("h2", "", "A closer look.");
    heading.id = "record-title";
    placeholder.append(
      heading,
      node("p", "", "Choose a wildling to open its field record."),
    );
    ui.detail.replaceChildren(placeholder);
    state.selected = null;
    return;
  }
  if (!visible.some((pokemon) => pokemon.name === state.selected))
    state.selected = visible[0].name;
  for (const pokemon of visible) {
    const button = node("button", "creature");
    button.type = "button";
    button.dataset.name = pokemon.name;
    button.setAttribute(
      "aria-pressed",
      String(pokemon.name === state.selected),
    );
    const initial = node("span", "initial", pokemon.name[0].toUpperCase());
    initial.setAttribute("aria-hidden", "true");
    const description = node("span");
    description.append(
      node("span", "creature-name", pokemon.name),
      node("span", "creature-types", types(pokemon).join(" / ")),
    );
    button.append(initial, description);
    button.addEventListener("click", () => select(pokemon));
    ui.records.append(button);
  }
  renderDetail(visible.find((pokemon) => pokemon.name === state.selected));
}
async function refresh() {
  if (state.busy) return;
  state.busy = true;
  ui.refresh.disabled = true;
  ui.refresh.textContent = "Refreshing…";
  ui.status.classList.remove("error");
  ui.status.textContent = "Reading your saved collection…";
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);
  try {
    const response = await fetch("/api/collection", {
      signal: controller.signal,
      cache: "no-store",
    });
    if (!response.ok)
      throw new Error(
        "The collection could not be read. Check the save file in the CLI, then retry.",
      );
    const data = await response.json();
    if (!Array.isArray(data.pokemon))
      throw new Error("The collection response was not valid.");
    state.pokemon = data.pokemon;
    state.loaded = true;
    ui.search.disabled = state.pokemon.length === 0;
    renderList();
    ui.status.textContent =
      "Up to date. Catch in the CLI, then refresh to see new records.";
  } catch (error) {
    ui.status.classList.add("error");
    ui.status.textContent =
      error.name === "AbortError"
        ? "Refresh timed out. Your previous records are still available; try again."
        : error.message;
    if (!state.loaded) {
      ui.count.textContent = "Unavailable";
    }
  } finally {
    clearTimeout(timeout);
    state.busy = false;
    ui.refresh.disabled = false;
    ui.refresh.textContent = "Refresh collection";
  }
}
ui.search.addEventListener("input", () => {
  state.limit = 100;
  renderList();
});
ui.more.addEventListener("click", () => {
  state.limit += 100;
  renderList();
});
ui.clear.addEventListener("click", () => {
  ui.search.value = "";
  renderList();
  ui.search.focus();
});
ui.refresh.addEventListener("click", refresh);
refresh();
