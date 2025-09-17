import config from "./config.js";

(function () {
  /**
   * Sets up a fallback image for favicon elements that fail to load.
   * @returns {void}
   */
  function setupFaviconFallback() {
    const applyFallback = (img) => {
      if (!img.dataset.fallbackApplied) {
        img.dataset.fallbackApplied = "true";
        img.src = config.static.favicon;
      }
    };

    document.querySelectorAll(".label-url-favicon").forEach((img) => {
      if (img.complete && img.naturalWidth === 0) {
        applyFallback(img);
        return;
      }

      img.addEventListener("error", () => applyFallback(img), { once: true });
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", setupFaviconFallback);
  } else {
    setupFaviconFallback();
  }
})();
