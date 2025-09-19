// repo.js

import api from "../services/api.js";
import Manager from "./manager.js";
import repo from "../repo.js";
import routes from "../services/routes.js";

const Repository = {
  init() {
    document.addEventListener("click", this.handleClick.bind(this));
  },

  // --- Event Delegation ---
  async handleClick(e) {
    const { target } = e;

    // Handle repo info
    if (target.matches("#current-repo-info")) return await this.open();
    // Handle 'Change Repo' button or 'Repositories' button in Side menu
    if (target.closest("#btn-list-repos")) return await this.openRepoList();
  },

  /**
   * Displays information about the current database in a modal.
   * @async
   */
  async open() {
    const modal = document.getElementById("modal-repo-information");
    const controller = Manager.register(modal);

    try {
      const url = routes.api.getDbInfo(repo.getCurrent());
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const dbInfo = await response.json();

      modal.querySelector("#repo-info-name").innerText = dbInfo.name;
      modal.querySelector("#repo-info-count").innerText = dbInfo.bookmarks;
      modal.querySelector("#repo-info-fav-count").innerText = dbInfo.favorites;
      modal.querySelector("#repo-info-tag-count").innerText = dbInfo.tags;
    } catch (error) {
      console.error("Error fetching DB info:", error);
      alert("Failed to load database information. Please try again.");
      controller.close();
    }

    controller.open();
  },

  async openRepoList() {
    const repos = await api.listDatabases();
    return repo.renderList(repos);
  },
};

export default Repository;
