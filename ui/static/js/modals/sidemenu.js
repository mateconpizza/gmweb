// sidemenu.js

// TODO:
// - [ ] Redo closeSideMenu, add addEventListener for 'escape'

/**
 * SideMenu handles the initialization and event delegation for the side menu.
 * @namespace SideMenu
 */
const SideMenu = {
  /** @type {HTMLElement|null} */
  menu: null,

  /** @type {HTMLElement|null} */
  overlay: null,

  init() {
    document.addEventListener("click", this.handleClick.bind(this));
    this.menu = document.getElementById("slide-menu");
    this.overlay = document.getElementById("menu-overlay");
  },

  // --- Event Delegation ---
  handleClick(e) {
    const { target } = e;

    // Handle `Hamburger` button (toggle menu)
    if (target.closest("#btn-hamburger, #btn-hamburger-mobile")) this.openSideMenu();
    // Handle `click` outside menu
    if (target.closest("#menu-overlay") || target.closest(".menu-item")) this.closeSideMenu();
    // Handle `new bookmark` button
    if (target.closest("#btn-new-bookmark")) this.closeSideMenu();
  },

  openSideMenu() {
    this.menu.classList.add("show");
    this.overlay.classList.add("show");
  },

  closeSideMenu() {
    this.menu.classList.remove("show");
    this.overlay.classList.remove("show");
  },
};

export default SideMenu;
