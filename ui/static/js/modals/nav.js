// nav.js

import Manager from "./manager.js";

const Nav = {
  /** @type {HTMLElement|null} */
  navBar: null,

  /** @type {HTMLElement|null} */
  tagsModal: null,

  /** @type {{ open: () => void, close: () => void, toggle: () => void, isOpen: () => boolean }|null} */
  tagsController: null,

  init() {
    document.addEventListener("click", this.handleClick.bind(this));
    this.navBar = document.getElementById("floating-nav");
    this.tagsModal = document.getElementById("modal-tags-mobile");
  },

  handleClick(e) {
    const { target } = e;

    // Handle `Tags` sidemenu and floating nav button
    if (target.closest("#tags-menu-item, #tags-floating-nav")) this.toggleTags();
    // Handle `Search` button in floating nav
    if (target.closest("#btn-search-nav")) return this.toggleSearchInput();
  },

  toggleTags() {
    if (!this.tagsController) {
      this.tagsController = Manager.register(this.tagsModal);
    }

    return this.tagsController.toggle();
  },

  toggleSearchInput() {
    const header = document.querySelector("header");
    if (!header) return;

    header.classList.toggle("visible");

    if (header.classList.contains("visible")) {
      header.querySelector('.search-bar input[type="text"]').focus();
    }

    return;
  },

  hideHeader() {
    const header = document.querySelector("header, .header, #header");
    if (!header) return;
    return header.classList.remove("visible");
  },
};

export default Nav;
