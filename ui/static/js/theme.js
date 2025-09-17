// theme.js

import Cookie from "./cookie.js";

const theme = {
  /**
   * Toggles the website's theme between light and dark mode.
   * @returns {void}
   */
  switcher() {
    const settingsToggle = document.getElementById("settings-dark-mode");
    const toggleMode = () => {
      const currentTheme = document.documentElement.getAttribute("data-theme");
      const newTheme = currentTheme === "dark" ? "light" : "dark";

      document.documentElement.setAttribute("data-theme", newTheme);

      localStorage.setItem("theme", newTheme);

      Cookie.set(Cookie.jar.themeMode, newTheme);

      if (settingsToggle) {
        settingsToggle.checked = newTheme === "dark";
      }
    };

    if (settingsToggle) {
      const initialTheme = document.documentElement.getAttribute("data-theme");
      settingsToggle.checked = initialTheme === "dark";

      settingsToggle.addEventListener("change", () => {
        toggleMode();
      });
    }

    const toggles = document.querySelectorAll(".dark-mode-toggle");
    if (toggles.length === 0) {
      console.warn("No dark mode toggle buttons found.");
      return;
    }

    toggles.forEach((btn) => {
      btn.addEventListener("click", () => {
        toggleMode();
      });
    });
  },
};

export default theme;
