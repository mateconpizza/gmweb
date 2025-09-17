// bookmark.js

import api from "../services/api.js";
import config from "../config.js";
import repo from "../repo.js";
import routes from "../services/routes.js";
import utils from "../utils/utils.js";
import Edit from "./edit.js";
import New from "./new.js";

const BookmarkMgr = {
  Edit,
  New,

  /**
   * Marks a bookmark as favorite.
   * @async
   * @param {string} id - The ID of the bookmark to be marked as favorite.
   * @returns {Promise<void>} A promise that resolves when the operation is complete.
   */
  async markAsFavorite(id) {
    if (!id) {
      console.error("MarkAsFavorite: bookmarkId is null");
      return;
    }

    const dbName = repo.getCurrent();
    const res = await api.favoriteRecord(dbName, id);
    if (res.ok) {
      console.log(res);
    } else {
      const data = await res.json();
      console.log("Failed to toggle Favorite:", data.error);
    }
  },

  /**
   * Marks a bookmark as visited by its ID.
   * @async
   * @param {string} id - The ID of the bookmark to mark as visited.
   * @returns {Promise<void>} A promise that resolves when the visit is recorded.
   */
  async markAsVisit(id) {
    const currentDB = repo.getCurrent();
    await api.recordVisit(currentDB, id);
  },

  /**
   * Handles the two-step delete process for a bookmark.
   * @param {HTMLElement} deleteBtn - The delete button element.
   * @param {string} id - The ID of the bookmark to delete.
   * @returns {Promise<void>}
   */
  async handleDeleteClick(deleteBtn, id) {
    const modal = deleteBtn.closest(".modal");
    const btnContainer = modal.querySelector("#btn-container-modal");
    const spinner = utils.createBtnSpinner(deleteBtn);
    const csrfToken = config.security.csrfToken();

    // messages
    const successMessageDiv = modal.querySelector("#form-success-message");
    const errorMessageDiv = modal.querySelector("#form-error-message");
    const messenger = utils.createFormMessenger(successMessageDiv, errorMessageDiv);

    if (deleteBtn.dataset.state === "confirm") {
      // Second click: confirmed, proceed with deletion
      try {
        if (!csrfToken && !window.DEV_MODE) {
          messenger.error("Error: No CSRF token found");
          return;
        }

        spinner.start();
        modal.querySelector("#btn-edit").classList.add("hidden");

        const dbName = repo.getCurrent();
        const goHome = routes.front.viewAllBookmarks(dbName);

        const res = await api.deleteRecord(dbName, id);
        const data = await res.json();

        setTimeout(() => {
          btnContainer.classList.add("hidden");

          if (res.ok) {
            messenger.success(data.message);
            setTimeout(() => {
              modal.classList.remove("show");
              window.location.href = goHome;
            }, 1000);
          } else {
            console.error("Delete failed:", res.status, res.statusText, data.error);
            messenger.error(`Delete failed: ${data.error || res.statusText}`);
            spinner.stop();
          }
        }, 1500);
      } catch (error) {
        console.error("Network or fetch error:", error);
        messenger.error("A network error occurred. Please try again.");
        spinner.stop();
      }
    } else {
      // First click: prompt for confirmation
      const originalText = deleteBtn.textContent;
      deleteBtn.textContent = "Confirm";
      deleteBtn.classList.add("confirm-state");
      deleteBtn.dataset.state = "confirm";

      // Reset state if not confirmed within 3s
      setTimeout(() => {
        if (deleteBtn.dataset.state === "confirm") {
          deleteBtn.textContent = originalText;
          deleteBtn.classList.remove("confirm-state");
          deleteBtn.dataset.state = "";
          spinner.stop();
        }
      }, 3000);
    }
  },

  /**
   * Handles the deletion of a bookmark card.
   * @async
   * @param {string} id - The ID of the bookmark to be deleted.
   * @returns {Promise<void>} A promise that resolves when the deletion process is complete.
   */
  async deleteCard(id) {
    if (!id) {
      console.error("Bookmark not found with id=", id);
      return;
    }
    const res = await api.deleteRecord(repo.getCurrent(), id);
    const card = document.querySelector(`.bookmark-card[data-id="${id}"]`);
    if (card) {
      setTimeout(() => {
        card.classList.remove("deleting");
        card.classList.add("deleted");
      }, 1000);
    }

    setTimeout(async () => {
      if (res.ok) {
        if (card) card.remove();
        // Handle success
        const response = await res.json();
        console.log(response.message);
        window.location.reload();
      } else {
        // Handle error
        const errorMessage = await res.json();
        throw new Error(`Delete failed: ${errorMessage.error || res.statusText}`);
      }
    }, 2500);
  },

  /**
   * Updates the status of a bookmark card in the modal.
   * @async
   * @param {HTMLElement} modal - The modal element containing the bookmark status information.
   * @returns {Promise<void>} A promise that resolves when the status update process is complete.
   */
  async checkStatus(modal) {
    if (!modal) {
      console.error("Modal for checking status not found");
      return;
    }
    const id = modal.dataset.modalId;
    const statusSpan = modal.querySelector("#status-span");
    const statusRefreshBtn = modal.querySelector("#btn-status-refresh");
    const spinner = utils.createBtnSpinner(statusRefreshBtn, false);
    spinner.start();
    statusSpan.textContent = "Checking status...";
    statusSpan.classList.remove("success", "error");

    const dbName = repo.getCurrent();

    try {
      const bookmark = await api.checkRecordStatus(dbName, id);
      console.log(bookmark);
      this._updateRecordStatus(statusSpan, bookmark.status_code, bookmark.status_text);
    } catch (err) {
      console.error("Error al verificar el estado:", err);
      statusSpan.textContent = "Network Error";
      statusSpan.classList.add("error");
    } finally {
      statusRefreshBtn.disabled = false;
      spinner.stop();
    }
    return;
  },

  /**
   * Fetches a snapshot from the Internet Archive for a given URL.
   * @param {string} url - The URL to fetch the snapshot for.
   * @param {HTMLElement} statusSpan - The element to display the fetching status and result.
   * @param {HTMLButtonElement} archiveBtn - The button element to control the loading state.
   * @returns {Promise<void>} A promise that resolves when the fetching process is complete.
   */
  async fetchArchive(url, statusSpan, archiveBtn) {
    statusSpan.classList.remove("error");
    statusSpan.textContent = "Fetching snapshot...";
    const spinner = utils.createBtnSpinner(archiveBtn, false);
    spinner.start();

    try {
      const res = await api.archiveURL(url);

      if (res.ok) {
        const data = await res.json();
        const link = this._createArchiveLink(data.archive_url, data.archive_timestamp);
        statusSpan.innerHTML = "";
        statusSpan.appendChild(link);
      } else {
        statusSpan.textContent = "Not found";
        statusSpan.classList.add("snapshot-link", "error");
      }
    } catch (e) {
      console.error(e);
    } finally {
      spinner.stop();
    }
  },

  async updateNotes(id, modal) {
    const textarea = modal.querySelector("textarea");
    const notes = textarea.value;
    if (!notes) return;

    const addBtn = modal.querySelector("#add-note-btn");
    addBtn.classList.remove("disable");
    const spinner = utils.createBtnSpinner(addBtn);
    spinner.start();

    const successMessageDiv = modal.querySelector("#form-success-message");
    const errorMessageDiv = modal.querySelector("#form-error-message");
    const messenger = utils.createFormMessenger(successMessageDiv, errorMessageDiv);

    const res = await api.updateNotes(id, notes);

    if (res.ok) {
      const data = await res.json();
      setTimeout(() => {
        spinner.stop();
        messenger.success(data.message, () => messenger.hide());
      }, 2000);
    } else {
      const errorMessage = await res.json();
      setTimeout(() => {
        spinner.stop();
        messenger.error(errorMessage.error);
      }, 2000);
    }

    console.log("BookmarkID", id);
    console.log("BookmarkNotes", notes);
  },

  /**
   * Updates the HTML element with the correct status class and text
   * based on the HTTP status code.
   * @private
   * @param {HTMLElement} element The HTML element to update.
   * @param {number} statusCode The HTTP status code.
   * @param {string} statusText The HTTP status text (optional).
   */
  _updateRecordStatus(element, statusCode, statusText = "") {
    element.classList.remove("unknown", "success", "error");

    let statusClass;
    let statusContent;

    if (statusCode === 0) {
      statusClass = "unknown";
      statusContent = "unknown";
    } else if (statusCode >= 200 && statusCode < 300) {
      statusClass = "success";
      statusContent = `${statusCode} ${statusText}`;
    } else if (statusCode >= 400) {
      statusClass = "error";
      statusContent = `${statusCode} ${statusText}`;
    } else {
      // For 3xx redirects or other codes
      statusClass = "";
      statusContent = `${statusCode} ${statusText}`;
    }

    if (statusClass) {
      element.classList.add(statusClass);
    }

    element.textContent = statusContent.trim();
  },

  /**
   * Creates an HTML anchor element (<a>) for a snapshot link.
   * @private
   * @param {string} archiveURL The URL for the archive snapshot.
   * @param {string|number} archiveTimestamp The timestamp to format as the link's text.
   * @returns {HTMLAnchorElement} The created <a> element.
   */
  _createArchiveLink(archiveURL, archiveTimestamp) {
    const linkElement = document.createElement("a");

    // Set its attributes.
    linkElement.href = archiveURL;
    linkElement.classList.add("snapshot-link");
    linkElement.target = "_blank";
    linkElement.rel = "noopener noreferrer";

    // Set the text content of the link.
    linkElement.textContent = utils.formatTimestamp(archiveTimestamp);

    return linkElement;
  },
};

export default BookmarkMgr;
