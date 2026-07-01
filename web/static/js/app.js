(function () {
  "use strict";

  var THEME_KEY = "worldcup-theme";
  var root = document.documentElement;

  function applyTheme(theme) {
    root.setAttribute("data-theme", theme);
  }

  function initTheme() {
    var stored = localStorage.getItem(THEME_KEY);
    var preferred = stored || (window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark");
    applyTheme(preferred);
  }

  function toggleTheme() {
    var current = root.getAttribute("data-theme") === "light" ? "light" : "dark";
    var next = current === "light" ? "dark" : "light";
    applyTheme(next);
    localStorage.setItem(THEME_KEY, next);
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
          if (!res.ok) throw new Error("refresh failed: " + res.status);
          window.location.reload();
        })
        .catch(function (err) {
          console.error(err);
          btn.disabled = false;
          btn.textContent = originalText;
          alert("Failed to update data. Please try again.");
        });
    });
  }

  initTheme();
  document.addEventListener("DOMContentLoaded", function () {
    var toggle = document.getElementById("theme-toggle");
    if (toggle) toggle.addEventListener("click", toggleTheme);
    initRefresh();
  });
})();
