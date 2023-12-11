"use strict";

function getDocuments() {
  var content = document.querySelector("#content");

  var data = [];
  var currentId = null;
  var current = { body: [], examples: [] };

  function push() {
    if (!currentId) return;
    if (current.body.length > 0 || current.examples.length > 0) {
      data.push({ id: currentId, ...current });
      current = { body: [], examples: [] };
    }
  }

  for (var i = 0; i < content.children.length; i++) {
    var child = content.children.item(i);

    // content of subrules should be collapsed to their parent rule
    if (child.classList.contains("anchor-subrule")) continue;

    if (child.id) {
      push();
      currentId = child.id;
      continue;
    }

    if (child.classList.contains("rules-example"))
      current.examples.push(child.textContent);
    else current.body.push(child.textContent);
  }
  push();

  return data;
}

var state = {
  query: "",
  results: [],
  selected: 0,

  abortController: new AbortController(),
  timer: setTimeout(function () {}, 0),
  setQuery: function (v) {
    this.query = v;

    clearTimeout(this.timer);
    this.timer = setTimeout(() => {
      this.abortController.abort();
      this.abortController = new AbortController();
      search(v, this.abortController.signal);
    }, 250);
  },
};

function search(v, signal) {
  console.log("searching", v);

  var results = window.index.search(v, { limit: 10 });
  if (!signal.aborted) {
    state.selected = 0;
    state.results = results;
    m.redraw();
  } else {
    console.warn("search aborted");
  }
}

function SearchResult(result, idx, selected, gotoSearchResult) {
  return m(
    "div.search-result",
    {
      key: result.refIndex,
      class: selected ? "search-result-selected" : null,
      onclick: function () {
        gotoSearchResult(idx);
      },
    },
    result.item.body[0]
  );
}

function SearchResults(state, gotoSearchResult) {
  return state.results.map((v, idx) =>
    SearchResult(v, idx, idx === state.selected, gotoSearchResult)
  );
}

function main() {
  var data = getDocuments();
  window.index = new Fuse(data, {
    includeScore: true,
    includeMatches: true,
    minMatchCharLength: 2,
    keys: [{ name: "body", weight: 2 }, "examples"],
  });

  var search = document.querySelector("#search");
  var modal = document.querySelector("#search-modal");

  function gotoSearchResult(idx) {
    var i = idx;
    if (idx === null) i = state.selected;

    console.log(idx, i, state.results);

    closeTocIfSmallScreen();
    var id = state.results[i].item.id;
    window.location.hash = id;
    search.blur();
    modal.classList.add("hide-search-modal");
  }

  search.addEventListener("keydown", function (ev) {
    var hide = false;
    var block = false;

    if (ev.key === "Escape") {
      hide = true;
      block = true;
    }

    if (ev.key === "ArrowDown" || ev.key === "Tab") {
      if (state.selected === null) {
        state.selected = 0;
      } else {
        state.selected += 1;
        if (state.selected >= state.results.length) {
          if (ev.key === "Tab") {
            state.selected = 0;
          } else {
            state.selected = state.results.length - 1;
          }
        }
      }

      block = true;
    }

    if (ev.key === "ArrowUp") {
      if (state.selected === null) {
        state.selected = 0;
      } else {
        state.selected -= 1;
        if (state.selected < 0) {
          state.selected = state.results.length - 1;
        }
      }

      block = true;
    }

    if (ev.key === "Enter" && state.selected !== null) {
      gotoSearchResult(null);
      hide = true;
      block = true;
    }

    if (hide) {
      modal.classList.add("hide-search-modal");
      search.blur();
    }

    if (block) {
      ev.preventDefault();
      m.redraw();
    }
  });

  search.addEventListener("focus", function () {
    modal.classList.remove("hide-search-modal");
  });

  search.addEventListener("blur", function () {
    state.selected = null;
    m.redraw();
  });

  search.addEventListener("input", function (ev) {
    modal.classList.remove("hide");
    state.setQuery(ev.target.value);
    m.redraw();
  });

  m.mount(modal, {
    view: function () {
      return SearchResults(state, gotoSearchResult);
    },
  });

  search.attributes.removeNamedItem("disabled");
  search.placeholder = "search rules";
}

window.addEventListener("load", main);
