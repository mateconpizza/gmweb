// edit.js

import config from "../config.js";
import repo from "../repo.js";
import api from "../services/api.js";
import routes from "../services/routes.js";
import { TagAutocomplete, tagOps } from "../tags.js";
import utils from "../utils/utils.js";
import bUtils from "./utils.js";

const SCRAPE = {
  TITLE: "title",
  DESC: "desc",
};

const Edit = {
  /**
   * Sets up the edition modal for editing bookmarks, including form handling,
   * tag autocomplete, and success/error message display.
   * @param {HTMLElement} modal - The modal element containing the edit form.
   * @returns {void}
   */
  async setup(modal) {
    if (!modal) {
      console.error("Modal edition not found");
      return;
    }

    // elements
    const form = modal.querySelector("#form-bookmark-edit");
    const dropdown = modal.querySelector("#dropdown-tags-cmp");
    const successMessageDiv = modal.querySelector("#form-success-message");
    const errorMessageDiv = modal.querySelector("#form-error-message");
    const messenger = utils.createFormMessenger(successMessageDiv, errorMessageDiv);
    const csrfToken = config.security.csrfToken();

    // inputs
    const idInput = modal.querySelector("#edit-bookmark-id");
    const urlInput = modal.querySelector("#edit-bookmark-url");
    const tagsInput = modal.querySelector("#edit-bookmark-tags");
    const titleInput = modal.querySelector("#edit-bookmark-title");
    const descInput = modal.querySelector("#edit-bookmark-desc");

    utils.resizeTextArea(descInput);
    utils.resizeTextArea(titleInput);

    tagsInput.value = tagOps.format(tagsInput.value);
    tagsInput.addEventListener("blur", () => {
      const raw = tagsInput.value || "";
      tagsInput.value = tagOps.format(raw);
    });

    // buttons
    const btnTitleRefresh = modal.querySelector("#btn-refresh-title");
    const btnTitleSpinner = utils.createBtnSpinner(btnTitleRefresh);
    const btnDescRefresh = modal.querySelector("#btn-refresh-desc");
    const btnDescSpinner = utils.createBtnSpinner(btnDescRefresh);

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

    btnTitleRefresh.addEventListener("click", async () => {
      btnTitleSpinner.start();
      await bUtils.scrapeInput(urlInput.value, titleInput, SCRAPE.TITLE);
      btnTitleSpinner.stop();
    });

    btnDescRefresh.addEventListener("click", async () => {
      btnDescSpinner.start();
      await bUtils.scrapeInput(urlInput.value, descInput, SCRAPE.DESC);
      btnDescSpinner.stop();
    });

    const accordionParams = modal.querySelector("#accordion-url-params");
    await bUtils.createUrlParamsList(urlInput, accordionParams);

    const btnContainer = modal.querySelector("#btn-container-edit");
    const btnCancel = modal.querySelector("#btn-cancel");
    btnCancel.classList.remove("hidden");
    const saveBtn = modal.querySelector("#btn-edit-submit");

    saveBtn.classList.remove("hidden");
    btnContainer.classList.remove("hidden");

    saveBtn.addEventListener("click", async (e) => {
      e.preventDefault();
      e.stopImmediatePropagation();

      // disable inputs
      bUtils.disableInputs(urlInput, tagsInput, titleInput, descInput);

      const spinner = utils.createBtnSpinner(saveBtn);
      spinner.start();
      btnCancel.classList.add("hidden");

      tagsInput.value = tagOps.format(tagsInput.value);

      const dbName = repo.getCurrent();
      const id = parseInt(idInput.value, 10);
      const bookmark = await api.getRecord(dbName, id);

      bookmark.url = urlInput.value.trim();
      bookmark.title = titleInput.value.trim();
      bookmark.desc = descInput.value.trim();
      bookmark.tags = tagsInput.value.trim()
        ? tagsInput.value
            .split(/[\s,]+/)
            .map((tag) => tag.trim().replace(/^#/, ""))
            .filter(Boolean)
        : [];

      try {
        if (!csrfToken && !window.DEV_MODE) {
          messenger.error("Error: No CSRF token found");
          return;
        }

        console.log("csrfToken.value", csrfToken);

        const res = await fetch(routes.api.updateBookmark(dbName, bookmark.id), {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": config.security.csrfToken(),
          },
          body: JSON.stringify(bookmark),
        });

        const data = await res.json();
        setTimeout(() => {
          if (res.ok) {
            btnContainer.classList.add("hidden");
            messenger.success(data.message);
            setTimeout(() => {
              modal.classList.remove("show");
              location.reload();
            }, 1000);
          } else {
            console.error("Error saving bookmark:", res.status, res.statusText, data.error);
            messenger.error(`${data.error || res.statusText}`);
            spinner.stop();
            bUtils.enableInputs(urlInput, tagsInput, titleInput, descInput);
          }
        }, 1500);
      } catch (error) {
        console.error("Network or fetch error:", error);
        messenger.error("A network error occurred. Please try again.");
        spinner.stop();
        bUtils.enableInputs(urlInput, tagsInput, titleInput, descInput);
      }
    });
  },

  /**
   * Opens the edit modal and fills the fields for the given bookmarkId.
   * @param {string|number} id - The ID of the bookmark.
   */
  async openUnused(id) {
    // FIX: maybe drop creating a bookmark modal for each bookmark and use this?
    try {
      const bookmark = await api.getRecord(id);
      if (!bookmark) {
        console.error("Bookmark not found", id);
        return;
      }

      const modal = document.getElementById("editModal");
      if (!modal) {
        console.error("Edit modal not found");
        return;
      }
      const controller = utils.modalController(modal);

      modal.querySelector(".label-url-favicon").src = bookmark.faviconLocal || "";
      modal.querySelector(".bookmark-detail-title").textContent = bookmark.title || "";

      const urlEl = modal.querySelector(".bookmark-detail-url");
      urlEl.textContent = bookmark.url || "";
      urlEl.href = bookmark.url || "#";

      // Form fields
      modal.querySelector("#edit-bookmark-url").value = bookmark.url || "";
      modal.querySelector("#edit-bookmark-tags").value = bookmark.tags ? bookmark.tags.join(", ") : "";
      modal.querySelector("#edit-bookmark-title").value = bookmark.title || "";
      modal.querySelector("#edit-bookmark-desc").value = bookmark.description || "";

      controller.open();
    } catch (err) {
      console.error("Error opening edit bookmark modal", err);
    }
  },
};

export default Edit;
