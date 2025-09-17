/**
 * API module for managing bookmarks and database interactions.
 *
 * This module provides a set of endpoints for performing CRUD operations on bookmarks.
 * Each function handles CSRF token validation.
 * @module api
 * @see module:types
 */

import config from "../config.js";
import repo from "../repo.js";
import routes from "./routes.js";

/**
 * @typedef {import('../types.js').Bookmark} Bookmark
 * @typedef {import('../types.js').BookmarkJSON} BookmarkJSON
 */


/**
 * Collection of API methods for managing bookmarks and databases.
 *
 * Each function wraps a `fetch` call to the backend,
 * handling CSRF validation and error reporting.
 * @namespace api
 */
const api = {
  /**
   * Deletes a bookmark record from the database by its ID.
   * @async
   * @param {string} dbName The name of the database where the record is stored.
   * @param {string} id The unique identifier of the record to be deleted.
   * @returns {Promise<Response|false|undefined>} -
   */
  async deleteRecord(dbName, id) {
    if (!config.security.csrfToken() && !window.DEV_MODE) {
      console.error("Error: CSRF token is missing. The request was not sent.");
      alert("An internal error occurred. Please refresh the page and try again.");
      return false;
    }

    try {
      return await fetch(routes.api.deleteBookmark(dbName, id), {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": config.security.csrfToken(),
        },
      });
    } catch (error) {
      console.error("Network or parsing error fetching URL data:", error);
    }
  },

  /**
   * Updates a bookmark record in the database.
   * @async
   * @param {string} dbName The name of the database.
   * @param {Bookmark} bookmark The bookmark object with updated data.
   * @returns {Promise<boolean>} -
   */
  async updateRecord(dbName, bookmark) {
    if (!config.security.csrfToken() && !window.DEV_MODE) {
      console.error("Error: CSRF token is missing. The request was not sent.");
      alert("An internal error occurred. Please refresh the page and try again.");
      return false;
    }

    try {
      const res = await fetch(routes.api.updateBookmark(dbName, bookmark.id), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": config.security.csrfToken(),
        },
        body: JSON.stringify(bookmark),
      });

      if (!res.ok) {
        const errorMessage = await res.json();
        console.error("Error saving bookmark:", res.status, res.statusText, errorMessage.error);
        return false;
      }

      const data = await res.json();
      console.log("Bookmark updated:", data);
      return true;
    } catch (err) {
      console.error("Error updating bookmark:", err);
      return false;
    }
  },

  /**
   * Fetches a single record by its ID from the specified database.
   * @async
   * @param {string} dbName The name of the database.
   * @param {string} id The unique ID of the record.
   * @returns {Promise<Bookmark|undefined>} -
   */
  async getRecord(dbName, id) {
    try {
      const response = await fetch(routes.api.getBookmarkById(dbName, id), {
        method: "GET",
        headers: { "Content-Type": "application/json" },
      });

      if (response.ok) {
        return await response.json();
      } else {
        const errorText = await response.json();
        console.error("Failed to fetch data:", response.status, errorText.Error);
      }
    } catch (error) {
      console.error("Network or parsing error fetching URL data:", error);
    }
  },

  /**
   * Records a visit to a bookmark by its ID.
   * @async
   * @param {string} dbName The name of the database.
   * @param {string} id The unique ID of the bookmark.
   * @returns {Promise<void|false>} -
   */
  async recordVisit(dbName, id) {
    if (!config.security.csrfToken() && !window.DEV_MODE) {
      console.error("Error: CSRF token is missing. The request was not sent.");
      return false;
    }

    try {
      const res = await fetch(routes.api.recordVisit(dbName, id), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": config.security.csrfToken(),
        },
      });

      if (res.ok) {
        const data = await res.json();
        console.log(data.message);
      } else {
        console.log("Failed to add a visit");
      }
    } catch (error) {
      console.error("Network or parsing error fetching URL data:", error);
    }
  },

  /**
   * Sends a URL to the Internet Archive to be saved.
   * @async
   * @param {string} url The URL to be archived.
   * @returns {Promise<object|string|false|undefined>} -
   */
  async archiveURL(url) {
    if (!config.security.csrfToken() && !window.DEV_MODE) {
      console.error("Error: CSRF token is missing. The request was not sent.");
      alert("An internal error occurred. Please refresh the page and try again.");
      return false;
    }

    try {
      const res = await fetch(`${routes.api.archiveUrl}?url=${encodeURIComponent(url)}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": config.security.csrfToken(),
        },
      });

      return await res.json();
    } catch (error) {
      console.error("Network or parsing error fetching URL data:", error);
    }
  },

  /**
   * Marks a specific bookmark as a favorite.
   * @async
   * @param {string} dbName The database name.
   * @param {string} id The unique ID of the bookmark.
   * @returns {Promise<Response|false|undefined>} -
   */
  async favoriteRecord(dbName, id) {
    if (!config.security.csrfToken() && !window.DEV_MODE) {
      console.error("Error: CSRF token is missing. The request was not sent.");
      alert("An internal error occurred. Please refresh the page and try again.");
      return false;
    }

    try {
      return await fetch(routes.api.toggleFavorite(dbName, id), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": config.security.csrfToken(),
        },
      });
    } catch (error) {
      console.error("Network or parsing error fetching URL data:", error);
    }
  },

  /**
   * Checks the status of a bookmark by its ID.
   * @async
   * @param {string} dbName The database name.
   * @param {string} bookmarkId The unique ID of the bookmark.
   * @returns {Promise<object|false|undefined>} -
   */
  async checkRecordStatus(dbName, bookmarkId) {
    if (!config.security.csrfToken() && !window.DEV_MODE) {
      console.error("Error: CSRF token is missing. The request was not sent.");
      alert("An internal error occurred. Please refresh the page and try again.");
      return false;
    }

    try {
      const response = await fetch(routes.api.updateStatus(dbName, bookmarkId), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": config.security.csrfToken(),
        },
      });

      if (response.ok) {
        return await response.json();
      } else {
        console.error(`Failed to fetch data: ${response.statusText}:: ${response.status}`);
      }
    } catch (error) {
      console.error("Network or parsing error fetching URL data:", error);
    }
  },

  /**
   * Fetches a list of all databases and their bookmark counts.
   * @async
   * @returns {Promise<Array<{name: string, bookmarks: number}>|undefined>} -
   */
  async listDatabases() {
    const databases = [];

    try {
      const res = await fetch(routes.api.getAllDbInfo);

      if (res.ok) {
        const data = await res.json();
        data.forEach((db) => {
          databases.push({ name: db.name, bookmarks: db.bookmarks });
        });
        return databases;
      } else {
        const errorMessage = await res.json();
        console.error("Error fetching databases:", res.status, res.statusText, errorMessage.error);
      }
    } catch (error) {
      console.error(`Failed to fetch databases: ${error.message}`);
    }
  },

  async updateNotes(id, notes) {
    try {
      const dbName = repo.getCurrent();
      const res = await fetch(routes.api.updateNotes(dbName, id), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": config.security.csrfToken(),
        },
        body: JSON.stringify({ notes: notes }),
      });

      return res
    } catch (error) {
      console.error(`Failed to update notes: ${error.message}`);
    }
  },
};

export default api;
