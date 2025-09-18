// index.js

import App from "./app.js";
import BookmarkMgr from "./bookmark/bookmark.js";

/**
 * Sets up event delegation on the main document.
 */
const GlobalEvents = {
  init() {
    document.addEventListener("click", this.handleClick.bind(this));
  },

  // --- Event Delegation ---
  async handleClick(e) {
    const { target } = e;

    // Handle `New bookmark` button
    if (target.closest("#btn-new-bookmark")) return BookmarkMgr.New.open();
    // Handle `Sort` bookmarks menu
    if (target.closest("#btn-sort-bookmark")) return this.showSortMenu(target);
  },

  showSortMenu(target) {
    const container = target.closest(".sort-menu-container");
    const dropdown = container.querySelector("#sort-dropdown");
    dropdown.classList.toggle("visible");
    console.log({ dropdown });
    return;
  },
};

document.addEventListener("DOMContentLoaded", () => {
  App.init();
  App.setupModals();
  GlobalEvents.init();
});
