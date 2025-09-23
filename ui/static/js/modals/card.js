// card.js

import BookmarkMgr from "../bookmark/bookmark.js";
import config from "../config.js";
import utils from "../utils/utils.js";
import BookmarkDetail from "./detail.js";
import Manager from "./manager.js";
import QRCode from "./qrcode.js";

const closeAllMenus = () => {
  const openMenus = document.querySelectorAll(".dropdown-menu.visible");
  openMenus.forEach((menu) => {
    menu.classList.remove("visible");
  });
};

const ACTIONS = {
  QRCODE: "qrcode",
  COPY: "copy",
  EDIT: "edition",
  DELETE: "delete",
  DETAIL: "detail",
};

const BookmarkCard = {
  init() {
    document.addEventListener("click", this.handleClick.bind(this));
  },

  // --- Event Delegation ---
  async handleClick(e) {
    const { target } = e;

    // Opens and adds a visit to the bookmark in the repository.
    if (target.matches("#url-visit, .bookmark-detail-url")) return this.markAsVisited(target);
    // Handle `Favorite` button
    if (target.closest(".bookmark-detail-fav-btn")) return this.markAsFavorite(target);
    // Handle `Expand` tags buttons
    if (target.classList.contains("tag-toggle")) return this.toggleCardTags(target);
    // Handle `3dots` bookmark card menu click
    if (target.matches(".dropdown-card-opt")) return this.handleMenuOpt(target);
    // Handle `3dots` menu
    if (target.closest(".btn-dots-menu-container")) return this.toggleMenu(target);
    // Handle `Icons` click in Compact mod
    if (target.closest(".bookmark-card-coso")) return this.handleCompactCardBtns(target);
    // Handle `3dots` click outside menu
    if (!target.closest(".dropdown-menu") || !target.closest(".btn-dots-menu-container")) return closeAllMenus();
  },

  /**
   * Handles a click on the 3dots menu.
   * @param {HTMLElement} target The element that was clicked, expected to be a `.dropdown-option`.
   */
  toggleMenu(target) {
    const dotsMenu = target.closest(".btn-dots-menu-container");
    const dropdown = dotsMenu.querySelector(".dropdown-menu");
    if (!dropdown) {
      console.error("DotsMenu: dropdown not found.");
    }

    if (!dropdown.classList.contains("visible")) {
      closeAllMenus();
    }

    dropdown.classList.toggle("visible");
  },

  /**
   * Handles a click on a dropdown menu option, delegating to the appropriate action handler.
   * @param {HTMLElement} target The element that was clicked, expected to be a `.dropdown-option`.
   */
  handleMenuOpt(target) {
    // FIX: merge with `handleCompactCardBtns`
    const action = target.getAttribute("data-action");
    const container = target.closest(".btn-dots-menu-container");
    const record = {
      id: container.dataset.id,
      url: container.dataset.bookmarkUrl,
    };

    target.closest(".dropdown-menu").classList.remove("visible");

    switch (action) {
      case ACTIONS.QRCODE: {
        console.log("Showing QR code for bookmark ID:", record.id);
        QRCode.open(record.id);
        break;
      }
      case ACTIONS.COPY: {
        console.log("Copying URL to clipboard:", record.url);
        utils.clipboard.copy(record.url);
        break;
      }
      case ACTIONS.EDIT: {
        this.openEditionModal(record.id);
        break;
      }
      case ACTIONS.DELETE: {
        // FIX: Deletion logic is not yet implemented.
        console.log("Deleting bookmark with ID:", record.id);
        console.error("Delete action is not yet implemented.");
        this.confirmDelete(record.id);
        break;
      }
      default: {
        console.log("Unknown dropdown action:", action);
      }
    }
  },

  markAsVisited(target) {
    const id = target.dataset.id;
    if (!id) {
      console.error("markAsVisited: bookmark id is null");
      return;
    }

    return BookmarkMgr.markAsVisit(id);
  },

  markAsFavorite(target) {
    const favBtn = target.closest(".bookmark-detail-fav-btn");
    const id = favBtn.dataset.bookmarkId;
    favBtn.classList.toggle("favorited");
    return BookmarkMgr.markAsFavorite(id);
  },

  toggleCardTags(target) {
    const container = target.closest(".bookmark-tags");
    const hiddenTags = container.querySelectorAll(".tag-hidden");
    hiddenTags.forEach((tag) => tag.classList.toggle("hidden"));
    if (!config.keyboard.vimMode) {
      const isOpen = target.textContent === "+";
      target.textContent = isOpen ? " - " : "+";
    }

    return;
  },

  handleCompactCardBtns(target) {
    // FIX: merge with `handleMenuOpt`
    const btn = target.closest(".bookmark-card-btn");
    if (!btn) return;

    const card = btn.closest(".bookmark-card");
    if (!card) return;

    const { action } = btn.dataset;
    const { id, url } = card.dataset;

    switch (action) {
      case ACTIONS.COPY:
        return utils.clipboard.copy(url);
      case ACTIONS.QRCODE:
        return QRCode.open(id);
      case ACTIONS.EDIT:
        return this.openEditionModal(id);
      case ACTIONS.DELETE:
        return this.confirmDelete(btn, id);
      case ACTIONS.DETAIL:
        return BookmarkDetail.open(id);
      default:
        console.warn(`Unknown compact card action: ${action}`);
    }
  },

  openEditionModal(id) {
    const modal = document.getElementById(`modal-edit-${id}`);
    const controller = Manager.register(modal);
    BookmarkMgr.Edit.setup(modal);
    controller.open();
  },

  confirmDelete(btn, id) {
    if (btn.dataset.state === "confirm") {
      console.log("Deleting!!!", { id });
    } else {
      if (!btn.dataset.originalContent) {
        btn.dataset.originalContent = btn.innerHTML;
      }

      btn.classList.add("confirm-state");
      btn.dataset.state = "confirm";
      setTimeout(() => {
        if (btn.dataset.state === "confirm") {
          btn.classList.remove("confirm-state");
          btn.dataset.state = "";

          if (btn.dataset.originalContent) {
            btn.innerHTML = btn.dataset.originalContent;
          }
        }
      }, 3000);
    }
  },
};

export default BookmarkCard;
