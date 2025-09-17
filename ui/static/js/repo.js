// repo.js

import config from "./config.js";
import Cookie from "./cookie.js";
import Modal from "./modals/modals.js";
import api from "./services/api.js";
import routes from "./services/routes.js";
import utils from "./utils/utils.js";

const KEYBINDS = config.keyboard.keybinds;

function getCurrent() {
  const match = window.location.pathname.match(/\/web\/([^/]+)/);
  return match ? match[1] : config.dbName;
}

/**
 * Collection of Utils for create, remove, and render repositories.
 * @namespace repoUtils
 */
const repoUtils = {
  load(dbName) {
    console.log(`Loading database: ${dbName}`);
    setTimeout(() => {
      this._hideDatabaseList();
    }, 1000);

    window.location.href = routes.front.viewAllBookmarks(dbName);
  },

  /**
   * Creates a new database by making a POST request to the server.
   * @private
   * @async
   * @param {string} dbName The name of the database to be created.
   * @param {Function} renderListFn The function to call to re-render the list of repositories after creation.
   * @returns {Promise<void>} A promise that resolves when the database creation process is complete.
   */
  async _createDatabase(dbName, renderListFn) {
    if (!dbName) {
      alert("Database name cannot be empty!");
      return;
    }
    const csrfToken = config.security.csrfToken();

    if (!csrfToken && !window.DEV_MODE) {
      console.error("Error: CSRF token is missing. The request was not sent.");
      alert("An internal error occurred. Please refresh the page and try again.");
      return;
    }

    console.log({ dbName, csrfToken });

    try {
      const response = await fetch(routes.api.createDb(dbName), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken,
        },
      });

      console.log("New repository:", response);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: "Unknown error" }));
        throw new Error(`Failed to create database: ${response.status} - ${errorData.message || response.statusText}`);
      }

      const data = await response.json();
      console.log({ data });
      const databases = await api.listDatabases();

      renderListFn(databases);

      // Hide the input field after successful creation and list refresh
      const newDbContainer = document.getElementById("new-repo-container");
      const inputGroup = newDbContainer?.querySelector(".new-repo-input-group");
      if (inputGroup) {
        inputGroup.classList.remove("active");
        inputGroup.querySelector(".new-repo-input").value = "";
      }

      const newDbButton = document.getElementById("btn-new-repo");
      if (newDbButton) {
        newDbButton.style.display = "flex";
      }
    } catch (error) {
      console.error("Error creating or refreshing database list:", error);
      alert(`Failed to create database: ${error.message}`);
    }
  },

  /**
   * Changes the repository name in the current URL's path.
   * @private
   * @param {string} repoName The new repository name to set in the URL.
   */
  _changeRepo(repoName) {
    // WIP: in iframe::not implemented yet
    const currentUrl = new URL(window.location.href);

    // pathname
    const pathSegments = currentUrl.pathname.split("/");

    // pathSegments[0] -> ""
    // pathSegments[1] -> "web"
    // pathSegments[2] -> "main" <- current repo|database
    // pathSegments[3] -> "bookmarks"
    // ...

    if (pathSegments.length > 2) {
      pathSegments[2] = repoName;
    }

    // join segs
    currentUrl.pathname = pathSegments.join("/");

    // Use the History API to update the URL without reloading
    // History.replaceState() is ideal for this case since we don't want the user to be able to
    // return to the previous URL with the "Back" button.
    // https://developer.mozilla.org/en-US/docs/Web/API/History/replaceState
    history.replaceState({}, "", currentUrl.toString());
    window.location.reload();
  },

  /**
   * Hides the database list window by removing the 'show' class from the modal overlay.
   * @private
   */
  _hideDatabaseList() {
    const modalOverlay = document.getElementById("modalOverlay");
    modalOverlay.classList.remove("show");
  },

  /**
   * Sorts repositories: current repository first, then others sorted by bookmark count (descending).
   * @private
   * @param {Array<object>} repos Array of repository objects (e.g., { name: 'repo1', bookmarks: 5 }).
   * @param {string} currentRepo The name of the currently active repository.
   * @returns {Array<object>} The sorted array of repository objects.
   */
  _sortRepos(repos, currentRepo) {
    let currentRepoItem = null;
    const otherRepos = [];

    for (const r of repos) {
      if (r.name === currentRepo) {
        currentRepoItem = r;
      } else {
        otherRepos.push(r);
      }
    }

    otherRepos.sort((a, b) => b.bookmarks - a.bookmarks);

    const finalSorted = [];
    if (currentRepoItem) finalSorted.push(currentRepoItem);
    finalSorted.push(...otherRepos);

    return finalSorted;
  },

  /**
   * Creates a `<span>` element for a repository name.
   * If the repository is the current one, it appends a checkmark icon.
   * @private
   * @param {object} db Repository object (e.g., { name: 'repo1', bookmarks: 5 }).
   * @param {string} currentRepo The name of the current repository.
   * @returns {HTMLSpanElement} The created span element.
   */
  _createRepoNameSpan(db, currentRepo) {
    const spanName = Object.assign(document.createElement("span"), {
      className: db.name === currentRepo ? "repo-current" : "repo-name",
      textContent: db.name,
    });

    if (db.name === currentRepo) {
      const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
      svg.setAttribute("width", "18");
      svg.setAttribute("height", "18");
      svg.setAttribute("viewBox", "0 0 24 24");
      svg.setAttribute("fill", "none");
      svg.setAttribute("stroke", "currentColor");
      svg.setAttribute("stroke-width", "2");
      svg.setAttribute("stroke-linecap", "round");
      svg.setAttribute("stroke-linejoin", "round");
      svg.classList.add("lucide", "lucide-check", "repo-current-icon");

      const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
      path.setAttribute("d", "M20 6 9 17l-5-5");
      svg.appendChild(path);

      spanName.append(svg);
    }

    return spanName;
  },

  /**
   * Creates the delete button for a repository list item.
   * @private
   * @param {object} db Repository object.
   * @param {Function} renderListFn The function to call to re-render the list of repositories after creation.
   * @returns {HTMLButtonElement} The created button element.
   */
  _createDeleteButton(db, renderListFn) {
    const deleteButton = document.createElement("button");
    deleteButton.className = "btn-repo-remove";
    deleteButton.innerHTML = `
      <svg
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M3 6H21"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <path
          d="M19 6V20C19 21.1046 18.1046 22 17 22H7C5.89543 22 5 21.1046 5 20V6M8 6V4C8 3.44772 8.44772 3 9 3H15C15.5523 3 16 3.44772 16 4V6"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <path
          d="M10 11V17"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <path
          d="M14 11V17"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    `;
    deleteButton.dataset.dbName = db.name;
    deleteButton.addEventListener("click", (e) => {
      e.stopPropagation();
      this._deleteDatabase(utils.stripSuffix(db.name, ".db"), deleteButton, renderListFn);
    });
    return deleteButton;
  },

  /**
   * Creates the checkbox for marking a repository as default.
   * @private
   * @param {object} db Repository object.
   * @param {HTMLElement} repoListEle The parent element containing all repository list items.
   * @returns {HTMLInputElement} The created checkbox element.
   */
  _createDefaultCheckbox(db, repoListEle) {
    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    checkbox.className = "minimal-checkbox";
    checkbox.title = "Set as default repository";
    checkbox.checked = db.name === config.dbName;

    checkbox.addEventListener("click", (e) => {
      e.stopPropagation();

      repoListEle.querySelectorAll(".minimal-checkbox").forEach((cb) => {
        cb.checked = false;
      });

      checkbox.checked = true;
      config.dbName = db.name;
      Cookie.set(Cookie.jar.defaultRepoName, db.name);
    });

    return checkbox;
  },

  /**
   * Builds a complete `<li>` element for a single repository item.
   * @private
   * @param {object} db Repository object.
   * @param {string} currentRepo Name of the current repository.
   * @param {HTMLElement} repoListEle The parent element for the list.
   * @param {Function} renderListFn The function to call to re-render the list of repositories after creation.
   * @returns {HTMLLIElement} The constructed list item element.
   */
  _buildRepoItem(db, currentRepo, repoListEle, renderListFn) {
    const listItem = document.createElement("li");
    listItem.className = "repo-item";

    const contentWrapper = document.createElement("div");
    contentWrapper.className = "repo-item-content-wrapper";
    contentWrapper.onclick = (e) => {
      if (!e.target.closest(".btn-repo-remove") && !e.target.closest(".minimal-checkbox")) {
        this.load(utils.stripSuffix(db.name, ".db"));
      }
    };

    const spanName = this._createRepoNameSpan(db, currentRepo);
    const spanCount = Object.assign(document.createElement("span"), {
      className: "repo-count",
      textContent: db.bookmarks,
    });
    const deleteButton = this._createDeleteButton(db, renderListFn);
    const checkbox = this._createDefaultCheckbox(db, repoListEle);

    contentWrapper.append(checkbox, spanName, spanCount);
    listItem.append(contentWrapper, deleteButton);

    return listItem;
  },

  /**
   * Public method to delete a database. Handles the confirmation logic.
   * @private
   * @param {string} dbName The name of the database to delete.
   * @param {HTMLButtonElement} delBtn The delete button element.
   * @param {Function} renderListFn The function to call to re-render the list of repositories after creation.
   */
  async _deleteDatabase(dbName, delBtn, renderListFn) {
    const isConfirmState = delBtn.classList.contains("confirm-state");

    if (!isConfirmState) {
      delBtn.classList.add("confirm-state");

      delBtn._confirmTimeoutId = setTimeout(() => {
        if (delBtn.classList.contains("confirm-state")) {
          delBtn.classList.remove("confirm-state");
        }
      }, 3000);
    } else {
      clearTimeout(delBtn._confirmTimeoutId);
      delBtn.classList.remove("confirm-state");

      console.log(`Deleting database: ${dbName}`);

      try {
        const res = await fetch(routes.api.deleteDb(dbName), {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
        });

        if (!res.ok) {
          const errorData = await res.json().catch(() => ({ error: "Unknown error" }));
          throw new Error(`Failed to delete database: ${res.status} - ${errorData.error || res.statusText}`);
        }

        console.log(`Database "${dbName}" deleted successfully.`);

        const updatedDatabases = await api.listDatabases();
        renderListFn(updatedDatabases);

        const currentDB = getCurrent();

        if (dbName === currentDB) {
          if (updatedDatabases.length > 0) {
            this.load(updatedDatabases[0].name);
          } else {
            window.location.href = window.location.origin;
          }
        }
      } catch (error) {
        console.error("Error deleting database:", error);
        alert(`Error deleting database: ${error.message}`);
        delBtn.classList.remove("confirm-state");
      }
    }
  },
};

const repo = {
  /**
   * Extracts the current database name from the URL path.
   *
   * This function parses the `window.location.pathname` to find a database name
   * located directly after '/web/'.
   * For example, if the URL is 'https://example.com/web/my-database/table', it will return 'my-database'.
   * If no database name is found in the path, it returns a predefined default database name.
   * @returns {string} The name of the current database or a default name if not found.
   */
  getCurrent: getCurrent,

  /**
   * Renders a list of repositories in the modal.
   * It sorts the repositories and then creates and appends a list item for each one.
   * @param {Array<object>} repos An array of repository objects.
   * @returns {Promise<void>}
   */
  async renderList(repos) {
    const modal = document.getElementById("modal-repo-list");
    const controller = Modal.Manager.register(modal);

    const currentRepo = repo.getCurrent();
    const repoListEle = document.getElementById("repo-unsorted-list");
    repoListEle.innerHTML = "";

    const sortedRepos = repoUtils._sortRepos(repos, currentRepo);

    sortedRepos.forEach((db) => {
      const listItem = repoUtils._buildRepoItem(db, currentRepo, repoListEle, repo.renderList);
      repoListEle.appendChild(listItem);
    });

    controller.open();
  },

  /**
   * Handles the UI for creating a new repository.
   * It manages the state between a "New Repository" button and an input field with a confirm button.
   * It also handles the creation of the database by calling the `createDatabase` function.
   * @returns {void}
   */
  setupNewRepoBtn() {
    const newDbContainer = document.getElementById("new-repo-container");
    const newDbButton = document.getElementById("btn-new-repo");

    // Exit if essential elements are not found
    if (!newDbContainer || !newDbButton) {
      console.warn("New database elements not found. Initialization skipped.");
      return;
    }

    /**
     * Function to transform the button into an input field
     */
    function showNewDbInput() {
      // Hide the initial button
      newDbButton.style.display = "none";

      // Create the input group if it doesn't exist
      let inputGroup = newDbContainer.querySelector(".new-repo-input-group");
      if (!inputGroup) {
        inputGroup = document.createElement("div");
        inputGroup.className = "new-repo-input-group";
        inputGroup.innerHTML = `
                <input type="text" class="new-repo-input" placeholder="Enter name">
                <button class="new-repo-confirm-btn">+ Add</button>
            `;
        newDbContainer.appendChild(inputGroup);
      }

      // Show the input group and focus on the input
      inputGroup.classList.add("active");
      const newDbInput = inputGroup.querySelector(".new-repo-input");
      const confirmButton = inputGroup.querySelector(".new-repo-confirm-btn");

      newDbInput.focus();

      // --- Event Listeners for the Input State ---

      // Handle confirmation click
      confirmButton.onclick = () => {
        repoUtils._createDatabase(newDbInput.value.trim(), repo.renderList);
        hideNewDbInput();
      };

      // Handle Enter key in the input field
      newDbInput.onkeydown = (e) => {
        if (e.key === KEYBINDS.actions.enter.key) {
          repoUtils._createDatabase(newDbInput.value.trim(), repo.renderList);
          hideNewDbInput();
        } else if (e.key === KEYBINDS.utility.escape.key) {
          hideNewDbInput();
        }
      };

      document.addEventListener("click", handleOutsideClick);
    }

    /**
     * Handles a click event outside the new repository input container.
     * If the click target is not inside the container and not the new repository button itself,
     * it hides the input field and removes this event listener.
     * @param {MouseEvent} event The click event object.
     */
    function handleOutsideClick(event) {
      if (!newDbContainer.contains(event.target) && event.target !== newDbButton) {
        hideNewDbInput();
        document.removeEventListener("click", handleOutsideClick);
      }
    }

    /**
     * Hides the input field for creating a new database and shows the original button again.
     * It also clears the input field's value and removes the event listener for clicks outside the container.
     */
    function hideNewDbInput() {
      const inputGroup = newDbContainer.querySelector(".new-repo-input-group");
      if (inputGroup) {
        inputGroup.remove();
      }

      // Show the button again
      newDbButton.style.display = "flex";
      document.removeEventListener("click", handleOutsideClick);
    }

    newDbButton.addEventListener("click", showNewDbInput);
  },
};

export default repo;
