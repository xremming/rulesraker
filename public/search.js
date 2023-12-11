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

function main() {
  var data = getDocuments();
  window.index = new Fuse(data, {
    includeScore: true,
    includeMatches: true,
    minMatchCharLength: 2,
    keys: [{ name: "body", weight: 2 }, "examples"],
  });
}

window.addEventListener("load", main);
