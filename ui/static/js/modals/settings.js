// settings.js

import config from "../config.js";
import Cookie from "../cookie.js";
import utils from "../utils/utils.js";
import Manager from "./manager.js";

/**
 * Settings manages the initialization and event handling for the settings modal.
 * @namespace SettingsApp
 */
const SettingsApp = {
  /** @type {HTMLElement|null} */
  modal: null,

  init() {
    document.addEventListener("click", this.handleClick.bind(this));
    if (!this.modal) {
      this.modal = document.getElementById("modal-settings");
    }
  },

  /**
   * Initialize state from cookies
   */
  restorePref() {
    this.modal = document.getElementById("modal-settings");
    this.setVimMode();
    this.setCompactMode();
    this.setThemeMode();
    this.setItemsPerPage();
  },

  // --- Event Delegation ---
  handleClick(e) {
    const { target } = e;

    // Handle open settings modal
    if (target.closest("#btn-settings")) return this.open();
    // Handle `Theme` selection
    // if (target.closest("#select-theme")) return this.selectTheme(target);
    // // Handle `CompactMode` toggle
    // if (target.closest("#checkbox-compact-mode")) return this.toggleCompactMode(target);
    // // Handle `VimMode` toggle
    // if (target.closest("#checkbox-vim-mode")) return this.toggleVimMode(target);
    // // Handle `ItemsPerPage` selection
    // if (target.closest("#items-per-page")) return this.setItemsPerPageCookie(target);
    // // Handle `DarkMode` toggle
    // if (target.closest("#settings-dark-mode")) return this.setThemeModeCookie(target);
    // Handle `Save` button in settings modal
    // if (target.closest("#btn-save")) return this.applySettings(target);
  },

  // --- Handlers ---
  /**
   * Opens the settings modal.
   * @function open
   * @returns {void}
   */
  open() {
    const controller = Manager.register(this.modal);
    controller.open();
  },

  selectTheme(target) {
    const themeSelect = target.closest("#select-theme");
    Cookie.set(Cookie.jar.theme, themeSelect.value);
    const mode = localStorage.getItem("theme");
    Cookie.set(Cookie.jar.themeMode, mode);
  },

  toggleCompactMode(target) {
    const compactModeToggle = target.closest("#checkbox-compact-mode");
    Cookie.set(Cookie.jar.compact, compactModeToggle.checked);
  },

  toggleVimMode(target) {
    const vimModeToggle = target.closest("#checkbox-vim-mode");
    Cookie.set(Cookie.jar.vim, vimModeToggle.checked);
  },

  setItemsPerPageCookie(target) {
    const itemsPerPageSelect = target.closest("#items-per-page");
    Cookie.set(Cookie.jar.itemsPage, itemsPerPageSelect.value);
  },

  setThemeModeCookie(target) {
    const settingsToggle = target.closest("#settings-dark-mode");
    Cookie.set(Cookie.jar.themeMode, settingsToggle.checked);
  },

  applySettings(target) {
    const saveBtn = target.closest("#btn-save");
    this.modal.querySelector("#btn-cancel")?.classList.add("hidden");
    const spinner = utils.createBtnSpinner(saveBtn);
    spinner.start();
    setTimeout(() => {
      // window.location.reload();
      spinner.stop();
    }, 500);
  },

  // --- State Initialization ---
  setVimMode() {
    const vimModeToggle = this.modal.querySelector("#checkbox-vim-mode");
    if (!vimModeToggle) return;

    const cookieValue = Cookie.get(Cookie.jar.vim);
    vimModeToggle.checked = cookieValue === "true";
    config.keyboard.vimMode = vimModeToggle.checked;
  },

  setCompactMode() {
    const compactModeToggle = this.modal.querySelector("#checkbox-compact-mode");
    if (!compactModeToggle) return;

    const cookieValue = Cookie.get(Cookie.jar.compact);
    compactModeToggle.checked = cookieValue === "true";
  },

  setThemeMode() {
    const settingsToggle = this.modal.querySelector("#settings-dark-mode");
    if (!settingsToggle) return;

    const cookieValue = Cookie.get(Cookie.jar.themeMode);
    settingsToggle.checked = cookieValue === "dark";
  },

  setItemsPerPage() {
    const itemsPerPageSelect = this.modal.querySelector("#items-per-page");
    if (!itemsPerPageSelect) return;

    const cookieValue = Cookie.get(Cookie.jar.itemsPage);
    if (cookieValue) {
      itemsPerPageSelect.value = cookieValue;
    }
  },
};

export default SettingsApp;
