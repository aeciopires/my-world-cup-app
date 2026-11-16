(function () {
  "use strict";

  var THEME_KEY = "worldcup-theme";
  var root = document.documentElement;

  function applyTheme(theme) {
    root.setAttribute("data-theme", theme);
  }

  // localStorage can throw (private browsing, blocked site data); a failure
  // here must not abort the rest of the script, or no button would work.
  function readStoredTheme() {
    try {
      return localStorage.getItem(THEME_KEY);
    } catch (e) {
      return null;
    }
  }

  function storeTheme(theme) {
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch (e) {
      /* preference just won't persist */
    }
  }

  function initTheme() {
    var stored = readStoredTheme();
    var prefersLight = window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches;
    applyTheme(stored || (prefersLight ? "light" : "dark"));
  }

  function toggleTheme() {
    var current = root.getAttribute("data-theme") === "light" ? "light" : "dark";
    var next = current === "light" ? "dark" : "light";
    applyTheme(next);
    storeTheme(next);
  }

  function initRefresh() {
    var btn = document.getElementById("refresh-btn");
    if (!btn) return;
    btn.addEventListener("click", function () {
      btn.disabled = true;
      var originalText = btn.textContent;
      btn.textContent = "Updating…";
      fetch("/refresh", { method: "POST" })
        .then(function (res) {
          if (res.status === 429) throw new Error("in-progress");
          if (!res.ok) throw new Error("refresh failed: " + res.status);
          window.location.reload();
        })
        .catch(function (err) {
          console.error(err);
          btn.disabled = false;
          btn.textContent = originalText;
          alert(err.message === "in-progress"
            ? "An update is already in progress. Please try again in a few seconds."
            : "Failed to update data. Please try again.");
        });
    });
  }

  // On the Groups page, each group's match results sit behind its own
  // <details class="group-matches">; this lets one button open or close all
  // of them at once instead of clicking every group individually.
  function initGroupMatchesToggle() {
    var btn = document.getElementById("toggle-all-matches");
    if (!btn) return;
    var detailsList = document.querySelectorAll(".group-matches");
    if (!detailsList.length) return;

    function setExpanded(expand) {
      detailsList.forEach(function (details) {
        details.open = expand;
      });
      btn.setAttribute("aria-expanded", String(expand));
      btn.textContent = expand ? "Collapse all matches" : "Expand all matches";
    }

    btn.addEventListener("click", function () {
      setExpanded(btn.getAttribute("aria-expanded") !== "true");
    });
  }

  initTheme();
  document.addEventListener("DOMContentLoaded", function () {
    var toggle = document.getElementById("theme-toggle");
    if (toggle) toggle.addEventListener("click", toggleTheme);
    initRefresh();
    initGroupMatchesToggle();
  });
})();
