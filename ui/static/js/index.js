// index.js

import App from "./app.js";
import BookmarkMgr from "./bookmark/bookmark.js";

/**
 * Sets up event delegation on the main document.
 */
const IndexEvents = {
  init() {
    document.addEventListener("click", this.handleClick.bind(this));
  },

  // --- Event Delegation ---
  async handleClick(e) {
    const { target } = e;

    // Handle `New bookmark` button
    if (target.closest("#btn-new-bookmark, #btn-new-nav")) return BookmarkMgr.New.open();
    // Handle `Sort` bookmarks menu
    if (target.closest("#btn-sort-bookmark")) return this.showSortMenu(target);
    // Handle `Sort` click outside
    if (!target.closest("#btn-sort-bookmark")) return this.hideSortMenu();
  },

  showSortMenu(target) {
    const container = target.closest(".sort-menu-container");
    const dropdown = container.querySelector("#sort-dropdown");
    dropdown.classList.toggle("visible");
    return;
  },
  hideSortMenu() {
    const dropdown = document.getElementById("sort-dropdown");
    if (!dropdown) return;
    if (dropdown.classList.contains("visible")) {
      dropdown.classList.toggle("visible");
    }
  },
};

document.addEventListener("DOMContentLoaded", () => {
  App.init();
  App.setupModals();
  IndexEvents.init();
});
