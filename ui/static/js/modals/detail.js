// detail.js

import BookmarkMgr from "../bookmark/bookmark.js";
import config from "../config.js";
import repo from "../repo.js";
import routes from "../services/routes.js";
import utils from "../utils/utils.js";
import Manager from "./manager.js";
import QRCode from "./qrcode.js";

/**
 * Sets up event handlers for modal detail interactions.
 */
const BookmarkDetail = {
  init() {
    document.addEventListener("click", this.handleClick.bind(this));
  },

  // --- Event Delegation ---
  async handleClick(e) {
    const { target } = e;
    // Handle 'Detail modal'
    if (target.matches(".bookmark-card-link, .bookmark-desc")) return this.open(target);
    // Handle accordion toggle
    if (target.closest(".accordion-header")) return this.handleAccordion(target);
    // Handle buttons
    if (target.closest("a[data-id]")) return this.buttonsHandler(target);
    // Handle 'Refresh' button (status)
    if (target.closest("#btn-status-refresh")) return await this.bookmarkStatus(target);
    // Handle fetch `Internet Archive` button
    if (target.closest("#btn-refresh-archive")) return await this.fetchArchive(target);
    // Handle `Save` notes button
    if (target.closest("#add-note-btn")) return this.addNote(target);
    // Handle `Un|check all` URL params
    if (target.closest("#toggle-all-params")) return this.toggleAllParams(target);
  },

  // --- Handlers ---
  open(target) {
    const bookmarkLink = target.closest(".bookmark-card-link");
    if (!bookmarkLink) {
      console.error("DetailEvent: Bookmark link not found");
      return;
    }
    const id = bookmarkLink.getAttribute("data-id");

    // Main bookmark modal
    const modal = document.getElementById(`modal-detail-${id}`);
    const controller = Manager.register(modal);

    // Accordion handler
    modal.querySelectorAll(".accordion.open").forEach((a) => {
      a.classList.remove("open");
    });

    // Tags handler
    const tagLinks = modal.querySelectorAll(".tag-in-modal");
    tagLinks.forEach((link) => {
      if (link.dataset.bound === "true") return;
      link.dataset.bound = "true";

      link.addEventListener("click", function (event) {
        event.preventDefault();

        const tag = this.getAttribute("data-tag");
        const encodedTag = encodeURIComponent(tag);

        const currentUrl = new URL(window.location.href);
        const params = currentUrl.searchParams;
        params.set("tag", encodedTag);
        window.location.href = currentUrl.toString();
      });
    });

    controller.open();
  },

  editBookmark(id) {
    const modal = document.getElementById(`modal-edit-${id}`);
    BookmarkMgr.Edit.setup(modal);
    const controller = Manager.register(modal);
    controller.open();
  },

  async fetchArchive(target) {
    const archiveBtn = target.closest("#btn-refresh-archive");
    const parent = archiveBtn.parentNode;
    const spanEle = parent.querySelector("#span-fetch-snapshot");
    const url = archiveBtn.dataset.url;
    return await BookmarkMgr.fetchArchive(url, spanEle, archiveBtn);
  },

  async bookmarkStatus(target) {
    const modal = target.closest(".modal");
    return await BookmarkMgr.checkStatus(modal);
  },

  buttonsHandler(target) {
    const clickedButton = target.closest("a[data-id]");
    const id = clickedButton.dataset.id;
    const buttonId = clickedButton.id;

    switch (buttonId) {
      case "btn-edit":
        return this.editBookmark(id);
      case "btn-delete":
        return BookmarkMgr.handleDeleteClick(clickedButton, id);
    }
  },

  addNote(target) {
    const button = target.closest("#add-note-btn");
    const id = button?.dataset.id;
    const modal = button?.closest(".modal");
    return BookmarkMgr.updateNotes(id, modal);
  },

  handleAccordion(target) {
    const accordionHeader = target.closest(".accordion-header");
    const accordion = accordionHeader.closest(".accordion");
    const toggleBtn = accordion.querySelector(".accordion-toggle");
    const accordionContent = accordion.querySelector(".accordion-content");

    const textarea = accordion.querySelector("textarea");
    if (textarea) {
      utils.resizeTextArea(textarea);
    }

    // Close other open accordions
    const container = accordion.closest(".modal, .accordion-container") || document;
    container.querySelectorAll(".accordion.open").forEach((a) => {
      if (a !== accordion) {
        a.classList.remove("open");
        const btn = a.querySelector(".accordion-toggle");
        if (btn) {
          btn.textContent = "+";
          btn.setAttribute("aria-expanded", false);
        }
      }
    });

    // QRCode accordion special handling
    if (accordion.querySelector("#accordion-qr-image")) {
      return this.openQRAccordion(accordion);
    }

    // Toggle accordion
    accordion.classList.toggle("open");
    const isOpen = accordion.classList.contains("open");

    if (!config.keyboard.vimMode) {
      toggleBtn.textContent = isOpen ? "−" : "+";
    }

    toggleBtn.setAttribute("aria-expanded", isOpen);
    setTimeout(() => {
      accordionContent.scrollIntoView({
        behavior: "smooth",
        block: "end",
        inline: "nearest",
      });
    }, 150);
  },

  openQRAccordion(accordion) {
    const isOpen = accordion.classList.contains("open");
    const toggleBtn = accordion.querySelector(".accordion-toggle");

    if (isOpen) {
      accordion.classList.remove("open");
      if (!config.keyboard.vimMode) {
        toggleBtn.textContent = "+";
      }
      toggleBtn.setAttribute("aria-expanded", "false");
      return;
    }

    const id = accordion.dataset.id;
    const qrImage = accordion.querySelector("#accordion-qr-image");
    const qrURL = routes.front.viewQrCode(repo.getCurrent(), id);

    QRCode.load({
      qrImage: qrImage,
      qrURL: qrURL,
      onSuccess: () => {
        accordion.classList.add("open");
        setTimeout(() => {
          qrImage.scrollIntoView({ behavior: "smooth", block: "end", inline: "nearest" });
        }, 300);
      },
    });

    if (!config.keyboard.vimMode) {
      toggleBtn.textContent = "−";
    }
    toggleBtn.setAttribute("aria-expanded", "true");
  },

  toggleAllParams(target) {
    const allParams = target.closest(".accordion-content").querySelector("#url-useless-params");
    console.log("toggle-all-params", allParams);
  },
};

export default BookmarkDetail;
