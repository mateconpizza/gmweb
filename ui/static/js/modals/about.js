/**
 * Module for handling the About functionality. Listens for clicks on the About link
 * and opens a modal dialog when triggered.
 */

import Manager from "./manager.js";

/**
 * AboutApp handles the initialization and click events for the about section of the application.
 * @namespace AboutApp
 */
const AboutApp = {
  /** @type {HTMLElement|null} */
  modal: null,

  init() {
    this.modal = document.getElementById("modal-about-app");
    document.addEventListener("click", this.handleClick.bind(this));
  },

  // --- Event Delegation ---
  handleClick(e) {
    const { target } = e;

    // Handle `About` button in side menu.
    const aboutBtn = target.closest("#btn-about-app");
    if (aboutBtn) {
      Manager.register(this.modal).open();
    }
  },
};

export default AboutApp;
