// new.js

import config from "../config.js";
import Modal from "../modals/modals.js";
import repo from "../repo.js";
import routes from "../services/routes.js";
import { TagAutocomplete, tagOps } from "../tags.js";
import utils from "../utils/utils.js";
import bUtils from "./utils.js";

const New = {
  /**
   * Opens the new bookmark modal and initializes the form.
   * @async
   * @param {string} [url] - The URL to pre-fill in the new bookmark form (optional).
   * @returns {Promise<void>} A promise that resolves when the bookmark form is initialized.
   */
  async open(url) {
    const modal = document.getElementById("modal-new-bookmark");
    if (!modal) {
      console.error("New Bookmark modal not found");
      return;
    }

    if (url) {
      modal.querySelector("#input-bookmark-url").value = url;
    }
    const controller = Modal.Manager.register(modal);

    if (!url) {
      this.reset(modal);
    }

    // sideMenu
    const slideMenu = document.getElementById("slide-menu");
    if (slideMenu.classList.contains("active")) {
      document.getElementById("btn-hamburger").classList.toggle("active");
      document.getElementById("menu-overlay").classList.toggle("active");
      slideMenu.classList.toggle("active");
    }

    this.setup(modal);
    controller.open();
  },

  /**
   * Initializes the new bookmark form within a modal.
   * @async
   * @param {HTMLElement} modal - The modal element containing the new bookmark form.
   * @returns {void}
   */
  async setup(modal) {
    const form = modal.querySelector("#form-new-bookmark");
    const dropdown = modal.querySelector("#dropdown-tags-cmp");
    const successMessageDiv = modal.querySelector("#form-success-message");
    const errorMessageDiv = modal.querySelector("#form-error-message");
    const messenger = utils.createFormMessenger(successMessageDiv, errorMessageDiv);

    if (!config.security.csrfToken() && !config.dev.enabled()) {
      messenger.error("Error: No CSRF token found");
      return;
    }

    // inputs
    const urlInput = modal.querySelector("#input-bookmark-url");
    const tagsInput = modal.querySelector("#input-bookmark-tags");
    const titleInput = modal.querySelector("#input-bookmark-title");
    const descInput = modal.querySelector("#input-bookmark-desc");
    const faviconInput = modal.querySelector("#input-bookmark-favicon");
    const faviconPreview = modal.querySelector("#img-bookmark-favicon");

    tagsInput.addEventListener("blur", () => {
      const raw = tagsInput.value || "";
      tagsInput.value = tagOps.format(raw);
    });

    // accordion
    const accordionParams = modal.querySelector("#accordion-url-params");

    // fetch website data
    const fetchTitleDebounced = utils.debounce(bUtils.scrapeURLData, 500);
    const initialUrl = urlInput.value.trim();
    if (initialUrl) {
      fetchTitleDebounced(initialUrl, titleInput, descInput, tagsInput, faviconInput, faviconPreview);
      await bUtils.createUrlParamsList(urlInput, accordionParams);
    }
    urlInput.addEventListener("input", async (e) => {
      const url = e.target.value.trim();
      fetchTitleDebounced(url, titleInput, descInput, tagsInput, faviconInput, faviconPreview);
      messenger.hide();
      utils.resizeAllTextArea();
      await bUtils.createUrlParamsList(urlInput, accordionParams);
    });

    new TagAutocomplete(tagsInput, dropdown, form, (selectedTag, lastQuery) => {
      const currentTags = tagsInput.value
        .split(/[\s,]+/)
        .map((tag) => tag.trim().replace(/^#/, ""))
        .filter(Boolean);

      const lastTagIndex = currentTags.findIndex((tag) => tag.toLowerCase() === lastQuery.toLowerCase());

      if (lastTagIndex !== -1) {
        currentTags[lastTagIndex] = selectedTag;
      } else if (!currentTags.includes(selectedTag)) {
        currentTags.push(selectedTag);
      }

      tagsInput.value = currentTags.map((tag) => `${tag}`).join(", ") + ", ";
      tagsInput.focus();
    });

    form.addEventListener("submit", async (e) => {
      e.preventDefault();
      e.stopImmediatePropagation();

      // buttons
      const btnCancel = modal.querySelector("#btn-cancel");
      const btnSubmit = modal.querySelector("#new-submit-btn");
      const btnContainer = modal.querySelector("#form-actions");
      // misc
      const spinner = utils.createBtnSpinner(btnSubmit);

      btnCancel.classList.add("hidden");
      spinner.start();

      const data = {
        url: urlInput.value.trim(),
        title: titleInput.value.trim(),
        desc: descInput.value.trim(),
        favicon_url: faviconInput.value.trim(),
        tags: tagsInput.value.trim()
          ? tagsInput.value
              .split(/[\s,]+/)
              .map((tag) => tag.trim().replace(/^#/, ""))
              .filter(Boolean)
          : [],
      };

      try {
        const res = await fetch(routes.api.createBookmark(repo.getCurrent()), {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": config.security.csrfToken(),
          },
          body: JSON.stringify(data),
        });

        setTimeout(async () => {
          if (res.ok) {
            const data = await res.json();
            const mesg = data.message || "Bookmark saved successfully!";
            messenger.success(mesg);
            btnContainer.classList.add("hidden");

            setTimeout(() => {
              const dbCurrent = repo.getCurrent();
              window.location.href = routes.front.viewAllBookmarks(dbCurrent);
            }, 2000);
          } else {
            const errorMessage = await res.json();
            console.error("Error saving bookmark:", res.status, res.statusText, errorMessage.error);
            console.log(errorMessage);
            messenger.error(`Error: ${errorMessage.error || res.statusText}`);
            btnCancel.classList.remove("hidden");
            spinner.stop();
          }
        }, 2000);
      } catch (error) {
        console.log("Fetch error:", error);
        messenger.error("A network error occurred. Please try again.");
        btnCancel.classList.remove("hidden");
        spinner.stop();
      }
    });
  },

  /**
   * Resets the new bookmark form within a given modal.
   * @param {HTMLElement} modal - The modal element containing the new bookmark form.
   * @returns {void} This function does not return a value.
   */
  reset(modal) {
    const form = modal.querySelector("#form-new-bookmark");
    form.reset();

    // Clear custom messages
    const successMesg = modal.querySelector("#form-success-message");
    const errorMesg = modal.querySelector("#form-error-message");
    const clearMesg = (ele) => {
      ele.textContent = "";
      ele.style.display = "none";
      ele.classList.add("hidden");
    };
    clearMesg(successMesg);
    clearMesg(errorMesg);

    // Reset any hidden inputs (like favicon)
    const faviconInput = modal.querySelector("#input-bookmark-favicon");
    faviconInput.value = "";
    const faviconPreview = modal.querySelector("#img-bookmark-favicon");
    faviconPreview.src = config.static.favicon;

    // Show cancel button again, reset spinners
    const btnCancel = modal.querySelector("#btn-cancel");
    const btnSubmit = modal.querySelector("#new-submit-btn");
    btnCancel.classList.remove("hidden");
    if (btnSubmit.spinner) btnSubmit.spinner.reset?.();

    // If your dropdown has leftover tags/autocomplete state, reset it too
    const dropdown = modal.querySelector("#dropdown-tags-cmp");
    dropdown.innerHTML = "";

    // Reset URL Params accordion
    const accordion = modal.querySelector(".accordion");
    console.log({ accordion });
    const spanUselessParams = accordion.querySelector("#url-useless-params");
    const accordionTitle = accordion.querySelector(".accordion-header").querySelector(".accordion-title");
    const toggleAllBtn = accordion.querySelector("#toggle-all-params");

    if (spanUselessParams) {
      spanUselessParams.innerHTML = ""; // Clear the displayed parameters
      spanUselessParams.classList.add("hidden"); // Hide the container
    }
    if (accordionTitle) {
      accordionTitle.textContent = "URL Params"; // Reset title text
    }
    if (toggleAllBtn) {
      toggleAllBtn.textContent = "Check All"; // Reset button text
    }
    // You might also want to hide the entire accordion if no params are shown
    accordion.classList.add("hidden");
  },
};

export default New;
