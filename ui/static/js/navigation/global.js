// navigation.js

import BookmarkMgr from "../bookmark/bookmark.js";
import config from "../config.js";
import Modal from "../modals/modals.js";
import utils from "../utils/utils.js";

const handlers = {
  newPasteBookmark: async () => {
    const url = await utils.clipboard.read();
    if (!utils.isValidUrl(url)) return;
    return BookmarkMgr.New.open(url);
  },

  newBookmark: () => {
    return BookmarkMgr.New.open();
  },
};

/**
 * Vim-style navigator for managing bookmarks.
 */
export default class VimNavigator {
  constructor() {
    this.bookmarks = document.querySelectorAll(".bookmark-card");
    this.selectedIndex = -1;
    this.isInsertMode = false;
    this.lastKeyPressTime = 0;
    this.lastPressedKey = "";
    this.multiSelectMode = false;
    this.keybinds = config.keyboard.keybinds;
    this.selectedBookmarks = new Set();
    this.bindEvents();
  }

  bindEvents() {
    document.addEventListener("keydown", this.handleKeydown.bind(this));
    document.querySelectorAll("input, textarea").forEach((el) => {
      el.addEventListener("focus", () => (this.isInsertMode = true));
      el.addEventListener("blur", () => (this.isInsertMode = false));
    });
  }

  async handleKeydown(event) {
    if (this.isInsertMode) return;
    if (this.isTypingInInput(event.target)) return;

    const key = event.key;
    const modal = Modal.Manager.getCurrent();
    if (modal) {
      const scrollableContent = modal.querySelector(".modal-base");
      // Down
      if (key === this.keybinds.navigation.down.key) {
        event.preventDefault();
        scrollableContent.scrollBy({ top: 60, behavior: "smooth" });
        return;
      }
      // Up
      if (key === this.keybinds.navigation.up.key) {
        event.preventDefault();
        scrollableContent.scrollBy({ top: -60, behavior: "smooth" });
        return;
      }

      return;
    }

    const currentTime = Date.now();
    console.log("VimNavigator", { key });

    // -- Navigation --
    // Handle 'gg' for top page
    if (key === "g" && this.lastPressedKey === "g" && currentTime - this.lastKeyPressTime < 300)
      return this.goToTop(event);
    // Handle 'M' for middle page
    if (event.shiftKey && key === this.keybinds.navigation.middle.key) return this.goToMiddle(event);
    // Handle 'G' for bottom page
    if (event.shiftKey && key === this.keybinds.navigation.bottom.key) return this.goToBottom(event);
    // Handle 'n' next page
    if (key === this.keybinds.navigation.pageNext.key) return this.goToNextPage();
    // Handle 'p' prev page
    if (key === this.keybinds.navigation.pagePrev.key) return this.goToPrevPage();

    // -- Actions --
    // Handle 'Enter' key for open Detail modal
    if (key === this.keybinds.actions.enter.key) return this.openDetailModal(event);
    // Handle 'O' for Open in new tab
    if (event.shiftKey && key === this.keybinds.actions.openTab.key) return this.openInTab();
    // Handle 'P' new bookmark from clipboard
    if (event.shiftKey && key === this.keybinds.actions.paste.key) return handlers.newPasteBookmark();
    // Handle 'a' new bookmark
    if (key === this.keybinds.actions.newBookmark.key) return handlers.newBookmark();
    // Handle "e" for edit bookmark modal
    if (key === this.keybinds.actions.edit.key) return this.openEditModal(event);
    // Handle 'f' for favorite
    if (key === this.keybinds.actions.favorite.key) return this.markAsFavorite();
    // Yank URL
    if (event.shiftKey && key === this.keybinds.actions.yank.key) return this.yankURL();
    // handle 'D' delete bookmark
    if (event.shiftKey && key === this.keybinds.actions.delete.key) await this.deleteBookmark(event);
    // Toggle multiselection mode with 'V' key
    if (key.toLowerCase() === "v" || key === this.keybinds.actions.selection?.key) {
      event.preventDefault();
      this.toggleMultiSelectMode();
      return;
    }
    // In multiselect mode, space or enter adds/removes current bookmark to selection
    if (this.multiSelectMode && (key === " " || key === this.keybinds.actions.enter.key)) {
      event.preventDefault();
      if (this.selectedIndex !== -1) {
        this.toggleBookmarkSelection(this.selectedIndex);
      }
      return;
    }

    // -- Utility --
    // Handle key for QRCode Modal
    if (key === this.keybinds.actions.qrcode.key) return this.openQRCodeModal();
    // Handle '?' for help
    if (key === this.keybinds.utility.help.key) return this.showHelp(event);
    // Handle 'S' for settings
    if (key === this.keybinds.utility.settings.key) return Modal.SettingsApp.open();
    // handle 'R' to reload page
    if (event.shiftKey && key === this.keybinds.utility.reload.key) return window.location.reload();
    // Handle toggle theme mode
    if (key === this.keybinds.utility.themeToggle.key) return this.toggleThemeMode();
    // Handle 's' for sort menu
    if (key === this.keybinds.utility.sort.key) return this.toggleSortBookmarksMenu();
    // Handle 'Escape'
    if (key === this.keybinds.utility.escape.key) {
      event.preventDefault();
      // Multi-selection
      if (this.multiSelectMode) {
        this.toggleMultiSelectMode();
      }
      // Remove `Deleting` state
      this.bookmarks.forEach((b) => b.classList.remove("confirm-state"));
      return;
    }

    if (key !== this.keybinds.navigation.down.key && key !== this.keybinds.navigation.up.key) {
      this.lastPressedKey = key;
      this.lastKeyPressTime = currentTime;
      return;
    }

    event.preventDefault();
    const direction = key === this.keybinds.navigation.down.key ? 1 : -1;
    let nextIndex = this.selectedIndex + direction;

    if (nextIndex < 0) {
      nextIndex = 0;
    } else if (nextIndex >= this.bookmarks.length) {
      nextIndex = this.bookmarks.length - 1;
    }

    this.updateSelection(nextIndex);
  }

  isTypingInInput(target) {
    const inputTypes = ["input", "textarea", "select"];
    return inputTypes.includes(target.tagName.toLowerCase()) || target.contentEditable === "true";
  }

  updateSelection(newIndex) {
    this.unhighlightAll();

    this.selectedIndex = newIndex;
    if (this.selectedIndex !== -1) {
      const selectedBookmark = this.bookmarks[this.selectedIndex];
      selectedBookmark.classList.add("highlighted");

      // Use 'center' to keep the highlighted bookmark in the middle
      selectedBookmark.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });
    }
  }

  toggleMultiSelectMode() {
    this.multiSelectMode = !this.multiSelectMode;

    if (this.multiSelectMode) {
      console.log("Entered multiselection mode");
      // Add visual indicator for multiselect mode
      document.body.classList.add("multiselect-mode");

      // If there's a currently highlighted bookmark, add it to selection
      if (this.selectedIndex !== -1) {
        this.toggleBookmarkSelection(this.selectedIndex);
      }
    } else {
      console.log("Exited multiselection mode");
      // Clear all selections and exit multiselect mode
      this.clearAllSelections();
      document.body.classList.remove("multiselect-mode");
    }
  }

  toggleBookmarkSelection(index) {
    const bookmark = this.bookmarks[index];
    if (!bookmark) return;

    if (this.selectedBookmarks.has(index)) {
      // Remove from selection
      this.selectedBookmarks.delete(index);
      bookmark.classList.remove("selected");
      bookmark.dataset.state = "";
    } else {
      // Add to selection
      this.selectedBookmarks.add(index);
      bookmark.classList.add("selected");
      bookmark.dataset.state = "selected";
    }

    console.log(`Selected bookmarks: ${Array.from(this.selectedBookmarks).join(", ")}`);
  }

  clearAllSelections() {
    this.selectedBookmarks.forEach((index) => {
      const bookmark = this.bookmarks[index];
      if (bookmark) {
        bookmark.classList.remove("selected");
        bookmark.classList.remove("confirm-state");
        bookmark.dataset.state = "";
      }
    });
    this.selectedBookmarks.clear();
  }

  async handleMultipleDelete() {
    const selectedIndexes = Array.from(this.selectedBookmarks);

    // Check if any selected bookmarks are in confirm state
    const confirmedBookmarks = selectedIndexes.filter((index) => {
      const bookmark = this.bookmarks[index];
      return bookmark && bookmark.dataset.state === "confirm";
    });

    if (confirmedBookmarks.length > 0) {
      // Delete all confirmed bookmarks
      for (const index of confirmedBookmarks) {
        const bookmark = this.bookmarks[index];
        const id = bookmark.getAttribute("data-id");

        bookmark.classList.remove("confirm-state");
        bookmark.classList.add("deleting");

        try {
          await BookmarkMgr.deleteCard(id);
        } catch (e) {
          console.error(e);
        }
      }

      // Remove deleted bookmarks from selection
      confirmedBookmarks.forEach((index) => this.selectedBookmarks.delete(index));
    } else {
      // Put all selected bookmarks into confirm state
      selectedIndexes.forEach((index) => {
        const bookmark = this.bookmarks[index];
        if (bookmark) {
          bookmark.classList.add("confirm-state");
          bookmark.dataset.state = "confirm";

          // Auto-remove confirm state after 3 seconds
          setTimeout(() => {
            if (bookmark.dataset.state === "confirm") {
              bookmark.classList.remove("confirm-state");
              bookmark.dataset.state = "selected";
            }
          }, 3000);
        }
      });
    }
  }

  async handleSingleDelete(index) {
    const bookmarkCard = this.bookmarks[index];
    if (!bookmarkCard) return;

    const id = bookmarkCard.getAttribute("data-id");

    if (bookmarkCard.dataset.state === "confirm") {
      bookmarkCard.classList.remove("confirm-delete");
      bookmarkCard.classList.add("deleting");

      try {
        await BookmarkMgr.deleteCard(id);
      } catch (e) {
        console.error(e);
      }
    } else {
      bookmarkCard.classList.add("confirm-state");
      bookmarkCard.dataset.state = "confirm";

      setTimeout(() => {
        bookmarkCard.classList.remove("confirm-state");
        bookmarkCard.dataset.state = "";
      }, 3000);
    }
  }

  unhighlightAll() {
    this.bookmarks.forEach((b) => b.classList.remove("highlighted"));
    // Note: Don't reset selectedIndex here as we still need it for navigation
  }

  triggerClick(modal, selector) {
    const btn = modal.querySelector(selector);
    if (btn) {
      btn.click();
    } else {
      console.warn(`GlobalNav: Button with selector "${selector}" not found.`);
    }
  }

  // -- Handlers --
  showHelp(e) {
    console.log("Help ??");
    e.preventDefault();
    Modal.HelpApp.toggle();
  }

  // -- Actions
  openDetailModal(e) {
    if (this.selectedIndex !== -1) {
      const link = this.bookmarks[this.selectedIndex].querySelector(".bookmark-card-link");
      if (link) {
        link.click();
        e.preventDefault();
      }
    }
    return;
  }
  openEditModal(e) {
    if (this.selectedIndex !== -1) {
      const link = this.bookmarks[this.selectedIndex].querySelector(".bookmark-card-link");
      if (link) {
        const id = link.dataset.id;
        const mainModal = document.getElementById(`modal-detail-${id}`);
        e.preventDefault();
        this.triggerClick(mainModal, "#btn-edit");
      }
    }
    return;
  }
  openQRCodeModal() {
    if (this.selectedIndex !== -1) {
      const id = this.bookmarks[this.selectedIndex].getAttribute("data-id");
      if (!id) {
        console.error("QRCode event: bookmark id not found");
        return;
      }
      Modal.QRCode.open(id);
    }
    return;
  }
  openInTab() {
    const link = this.bookmarks[this.selectedIndex].querySelector("#url-visit");
    if (!link) {
      console.error("OpenInTab: bookmark link not found");
      return;
    }
    link.click();
  }

  yankURL() {
    if (this.selectedIndex !== -1) {
      const link = this.bookmarks[this.selectedIndex].querySelector("#url-visit");
      if (!link) {
        console.error("Yank event: bookmark link not found");
        return;
      }
      return utils.clipboard.copy(link.href);
    }
    return;
  }

  markAsFavorite() {
    // Multi-selection
    if (this.multiSelectMode && this.selectedBookmarks.size > 0) {
      const selectedIndexes = Array.from(this.selectedBookmarks);
      for (const index of selectedIndexes) {
        const bookmark = this.bookmarks[index];
        const favBtn = bookmark.querySelector(".bookmark-detail-fav-btn");
        if (favBtn) favBtn.click();
      }
      return;
    }

    if (this.selectedIndex !== -1) {
      const favBtn = this.bookmarks[this.selectedIndex].querySelector(".bookmark-detail-fav-btn");
      if (!favBtn) return;
      favBtn.click();
      return;
    }
    return;
  }

  async deleteBookmark(e) {
    e.preventDefault();
    if (this.multiSelectMode && this.selectedBookmarks.size > 0) {
      // Handle deletion for multiple selected bookmarks
      await this.handleMultipleDelete();
    } else if (this.selectedIndex !== -1) {
      // Handle single bookmark deletion (existing logic)
      await this.handleSingleDelete(this.selectedIndex);
    }
    return;
  }

  toggleThemeMode() {
    return document.getElementById("dark-mode-toggle")?.click();
  }

  toggleSortBookmarksMenu() {
    const dropdown = document.getElementById("sort-dropdown");
    dropdown.classList.toggle("visible");
  }

  // -- Navigation --
  goToTop(e) {
    e.preventDefault();
    if (this.bookmarks.length > 0) {
      this.updateSelection(0);
    }
    this.lastPressedKey = "";
    return;
  }
  goToMiddle(e) {
    e.preventDefault();
    if (this.bookmarks.length > 0) {
      const middleIndex = Math.floor(this.bookmarks.length / 2);
      this.updateSelection(middleIndex);
    }
    return;
  }
  goToBottom(e) {
    e.preventDefault();
    if (this.bookmarks.length > 0) {
      const lastIndex = this.bookmarks.length - 1;
      this.updateSelection(lastIndex);
    }
    return;
  }
  goToNextPage() {
    return document.getElementById("btn-next-page")?.click();
  }
  goToPrevPage() {
    return document.getElementById("btn-prev-page")?.click();
  }
}
