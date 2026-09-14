(function () {
  "use strict";

  var root = document.documentElement;
  var rootPrefix = root.getAttribute("data-root") || "";

  function byID(id) {
    return document.getElementById(id);
  }

  function storageGet(key) {
    try { return localStorage.getItem(key); } catch (e) { return null; }
  }

  function storageSet(key, value) {
    try { localStorage.setItem(key, value); } catch (e) { /* ignore */ }
  }

  function storageRemove(key) {
    try { localStorage.removeItem(key); } catch (e) { /* ignore */ }
  }

  /* Theme */

  var themeSelect = byID("theme-select");
  var storedTheme = storageGet("godoc2-theme");
  if (themeSelect) {
    themeSelect.value = storedTheme === "light" || storedTheme === "dark" ? storedTheme : "system";
    themeSelect.addEventListener("change", function () {
      var value = themeSelect.value;
      if (value === "system") {
        root.removeAttribute("data-theme");
        storageRemove("godoc2-theme");
      } else {
        root.setAttribute("data-theme", value);
        storageSet("godoc2-theme", value);
      }
    });
  }

  /* Popovers */

  var settingsMenu = byID("settings-menu");
  var helpPanel = byID("help-panel");

  function closePopovers() {
    if (settingsMenu) { settingsMenu.hidden = true; }
    if (helpPanel) { helpPanel.hidden = true; }
  }

  function toggle(panel, other) {
    if (!panel) { return; }
    var open = panel.hidden;
    closePopovers();
    if (other) { other.hidden = true; }
    panel.hidden = !open;
  }

  var settingsButton = byID("settings-button");
  if (settingsButton) {
    settingsButton.addEventListener("click", function () { toggle(settingsMenu, helpPanel); });
  }
  var helpButton = byID("help-button");
  if (helpButton) {
    helpButton.addEventListener("click", function () { toggle(helpPanel, settingsMenu); });
  }

  document.addEventListener("click", function (event) {
    var target = event.target;
    if (target.closest && (target.closest(".popover") || target.closest(".topbar-buttons"))) {
      return;
    }
    closePopovers();
  });

  /* Sidebar resize */

  var sidebar = byID("sidebar");
  var resizer = byID("sidebar-resizer");
  if (sidebar && resizer) {
    var dragging = false;
    resizer.addEventListener("mousedown", function (event) {
      dragging = true;
      resizer.classList.add("active");
      document.body.style.userSelect = "none";
      event.preventDefault();
    });
    document.addEventListener("mousemove", function (event) {
      if (!dragging) { return; }
      var width = Math.min(Math.max(event.clientX, 140), 600);
      root.style.setProperty("--sidebar-width", width + "px");
    });
    document.addEventListener("mouseup", function () {
      if (!dragging) { return; }
      dragging = false;
      resizer.classList.remove("active");
      document.body.style.userSelect = "";
      var width = parseInt(getComputedStyle(root).getPropertyValue("--sidebar-width"), 10);
      if (width) { storageSet("godoc2-sidebar-width", String(width)); }
    });
  }

  /* Read more */

  document.querySelectorAll(".read-more").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();
      event.stopPropagation();
      var doc = button.parentElement;
      var summary = doc.querySelector(".summary");
      var full = doc.querySelector(".full");
      if (summary) { summary.hidden = true; }
      if (full) { full.hidden = false; }
      button.hidden = true;
    });
  });

  /* Keep clicks on links inside summaries from toggling the details. */

  document.querySelectorAll("details.item > summary a").forEach(function (link) {
    link.addEventListener("click", function (event) { event.stopPropagation(); });
  });

  function setAllDetails(open) {
    document.querySelectorAll("details.item, details.example, details.group").forEach(function (d) { d.open = open; });
  }

  /* Search */

  var input = byID("search-input");
  var results = byID("search-results");
  var content = byID("content");
  var index = window.godoc2SearchIndex || [];

  function escapeHTML(text) {
    return String(text).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" }[c];
    });
  }

  function score(entry, query) {
    var name = entry.name.toLowerCase();
    var last = name.slice(name.lastIndexOf(".") + 1);
    if (last === query) { return 0; }
    if (last.indexOf(query) === 0) { return 1; }
    if (name.indexOf(query) === 0) { return 2; }
    if (name.indexOf(query) >= 0) { return 3; }
    if (entry.package.toLowerCase().indexOf(query) >= 0) { return 4; }
    return -1;
  }

  function search(query) {
    query = query.trim().toLowerCase();
    if (!results || !content) { return; }
    if (query === "") {
      results.hidden = true;
      results.innerHTML = "";
      content.hidden = false;
      return;
    }
    var matches = [];
    for (var i = 0; i < index.length; i++) {
      var s = score(index[i], query);
      if (s >= 0) { matches.push({ entry: index[i], score: s }); }
    }
    matches.sort(function (a, b) {
      if (a.score !== b.score) { return a.score - b.score; }
      return a.entry.name.localeCompare(b.entry.name);
    });
    matches = matches.slice(0, 100);
    var html = "<h2>Results for “" + escapeHTML(query) + "”</h2>";
    if (matches.length === 0) {
      html += "<p class=\"message\">No results.</p>";
    }
    for (var j = 0; j < matches.length; j++) {
      var e = matches[j].entry;
      html += "<div class=\"search-result\"><span class=\"kind\">" + escapeHTML(e.kind) + "</span>" +
        "<span><a href=\"" + escapeHTML(rootPrefix + e.path) + "\">" + escapeHTML(e.name) + "</a>" +
        "<span class=\"pkg\">" + escapeHTML(e.package) + "</span></span></div>";
    }
    results.innerHTML = html;
    results.hidden = false;
    content.hidden = true;
  }

  if (input) {
    input.addEventListener("input", function () { search(input.value); });
    if (input.value) { search(input.value); }
  }

  /* Keyboard shortcuts */

  document.addEventListener("keydown", function (event) {
    var active = document.activeElement;
    var typing = active && (active.tagName === "INPUT" || active.tagName === "SELECT" || active.tagName === "TEXTAREA");
    if (event.key === "Escape") {
      closePopovers();
      if (input && typing) {
        input.value = "";
        search("");
        input.blur();
      }
      return;
    }
    if (typing || event.ctrlKey || event.metaKey || event.altKey) { return; }
    switch (event.key) {
      case "s":
      case "S":
      case "/":
        if (input) { input.focus(); event.preventDefault(); }
        break;
      case "?":
        toggle(helpPanel, settingsMenu);
        event.preventDefault();
        break;
      case "+":
      case "=":
        setAllDetails(true);
        break;
      case "-":
        setAllDetails(false);
        break;
      default:
        break;
    }
  });
})();
