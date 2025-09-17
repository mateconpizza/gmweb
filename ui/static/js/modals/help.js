// help.js

import config from "../config.js";
import Manager from "./manager.js";

const KEYBINDS = config.keyboard.keybinds;

/**
 * Renders a help modal displaying keybinds and their descriptions.
 * @param {HTMLElement} modal - The modal element to render the help content in.
 * @param {object} keybinds - An object containing keybind categories and their details.
 * @param {object} [options] - Optional settings for rendering.
 * @param {Array<string>} [options.ignore] - An array of keys to ignore when rendering.
 * @param {boolean} [options.footer] - Add footer to the modal
 */
function renderKeybindsHelp(modal, keybinds, options = {}) {
  if (!(modal instanceof HTMLElement)) return;

  const addFooter = options.footer || false;
  const ignoreKeys = options.ignore || [];
  const container = modal.querySelector(".modal-help-content");
  if (!container) return;

  // Clear previous content
  container.innerHTML = "";

  for (const category in keybinds) {
    const filtered = Object.values(keybinds[category]).filter(
      ({ key, done = true }) => done && !ignoreKeys.includes(key),
    );
    if (filtered.length === 0) continue;

    const categoryContainer = document.createElement("div");
    categoryContainer.className = "category-container";

    const categoryTitle = document.createElement("h4");
    categoryTitle.textContent = category.charAt(0).toUpperCase() + category.slice(1);
    categoryContainer.appendChild(categoryTitle);

    const keybindsGrid = document.createElement("div");
    keybindsGrid.className = "keybinds-grid";

    filtered.forEach(({ key, description, shortcut }) => {
      const item = document.createElement("div");
      item.className = "keybind-item";

      const kbd = document.createElement("kbd");
      kbd.className = "keybind-key";
      kbd.textContent = shortcut ?? key;

      const span = document.createElement("span");
      span.className = "keybind-description";
      span.textContent = description;

      item.appendChild(kbd);
      item.appendChild(span);
      keybindsGrid.appendChild(item);
    });

    categoryContainer.appendChild(keybindsGrid);
    container.appendChild(categoryContainer);
  }

  // Footer
  if (addFooter) {
    const footer = document.createElement("div");
    footer.className = "modal-help-footer";
    footer.innerHTML = `<p>Press <kbd>?</kbd> anytime to toggle this help</p>`;
    container.appendChild(footer);
  }
}

const HelpApp = {
  /** @type {HTMLElement|null} */
  modal: null,

  /** @type {{ open: () => void, close: () => void, toggle: () => void, isOpen: () => boolean }|null} */
  controller: null,

  init() {
    this.modal = document.getElementById("modal-help-app");
    this.renderKeybindsHelp(this.modal, KEYBINDS, {
      ignore: [KEYBINDS.utility.escape.key, KEYBINDS.navigation.middle.key, KEYBINDS.utility.help.key],
    });
  },

  toggle() {
    if (!this.controller) {
      this.controller = Manager.register(this.modal);
    }

    this.controller.toggle();
  },

  renderKeybindsHelp: renderKeybindsHelp,
};

export default HelpApp;
