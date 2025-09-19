// import.js

import Manager from "./manager.js";
import repo from "../repo.js";
import routes from "../services/routes.js";
import utils from "../utils/utils.js";

const endpoints = {
  "html-import": routes.api.importHtml,
  "git-json": routes.api.importRepoJson,
  "git-gpg": routes.api.importRepoGpg,
};

const ImportManager = {
  /** @type {HTMLElement|null} */
  modal: null,

  /** @type {{ open: () => void, close: () => void, toggle: () => void, isOpen: () => boolean }|null} */
  controller: null,

  /** @type {string} */
  selectedSource: null,

  init() {
    document.addEventListener("click", this.handleClick.bind(this));
    this.modal = document.getElementById("modal-import");
    if (!this.modal) {
      console.error("Modal import not found");
      return;
    }

    this.setupFileInput();
  },

  // --- Event Delegation ---
  async handleClick(e) {
    const { target } = e;

    // Open Import Modal
    if (target.closest("#btn-import")) return this.open();
    // Handle selecting an import option
    if (target.closest(".import-option")) return this.selectOption(target);
    // Handle import button inside modal
    if (target.closest("#btn-import-bookmarks")) await this.handleImport();
  },

  // --- Setup ---
  setupFileInput() {
    const fileInput = this.modal.querySelector("#file-input");
    if (!fileInput) {
      console.warn("ImportManager: file input not found");
      return;
    }

    fileInput.addEventListener("change", (e) => {
      const importBtn = this.modal.querySelector("#btn-import-bookmarks");
      importBtn.disabled = !e.target.files.length;

      if (e.target.files.length) {
        const label = this.modal.querySelector(".file-input-label");
        label.textContent = e.target.files[0].name;
      }
    });
  },

  // --- UI Actions ---
  open() {
    this.controller = Manager.register(this.modal);
    this.controller.open();
  },

  selectOption(target) {
    const importOpt = target.closest(".import-option");
    this.selectedSource = importOpt.dataset.source;

    // Cache DOM
    const fileSection = this.modal.querySelector("#file-input-section");
    const repoInputSection = this.modal.querySelector("#repo-input-section");
    const importBtn = this.modal.querySelector("#btn-import-bookmarks");
    const successMessageDiv = this.modal.querySelector("#form-success-message");
    const errorMessageDiv = this.modal.querySelector("#form-error-message");
    const messenger = utils.createFormMessenger(successMessageDiv, errorMessageDiv);

    // Reset
    document.querySelectorAll(".import-option").forEach((opt) => opt.classList.remove("selected"));
    messenger.hide();
    importOpt.classList.add("selected");
    fileSection.classList.remove("active");
    repoInputSection.classList.remove("active");
    importBtn.disabled = true;

    // Activate proper section
    switch (this.selectedSource) {
      case "html-import":
        fileSection.classList.add("active");
        break;
      case "git-json":
      case "git-gpg":
        repoInputSection.classList.add("active");
        importBtn.disabled = false;
        break;
      default:
        console.error("ImportManager: unknown source", this.selectedSource);
    }
  },

  async handleImport() {
    const successMessageDiv = this.modal.querySelector("#form-success-message");
    const errorMessageDiv = this.modal.querySelector("#form-error-message");
    const messenger = utils.createFormMessenger(successMessageDiv, errorMessageDiv);

    if (!this.selectedSource) {
      console.error("Select an import source");
      messenger.error("Select an import source");
      return;
    }

    const fileInput = this.modal.querySelector("#file-input");
    const repoInput = this.modal.querySelector("#repoInputId");
    const dbName = repo.getCurrent();

    const formData = new FormData();
    const endpoint = endpoints[this.selectedSource](dbName);

    if (this.selectedSource === "html-import") {
      if (!fileInput.files.length) {
        alert("Please select a HTML file to upload.");
        return;
      }
      formData.append("file", fileInput.files[0]);
    } else if (this.selectedSource === "git-json" || this.selectedSource === "git-gpg") {
      if (!repoInput.value.trim()) {
        alert("Please enter a repository URL.");
        return;
      }
      formData.append("repo", repoInput.value.trim());
    }

    const modalImportFooter = this.modal.querySelector("#modal-import-footer");
    modalImportFooter.classList.remove("hidden");
    const importBtn = this.modal.querySelector("#btn-import-bookmarks");
    const spinner = utils.createBtnSpinner(importBtn);
    spinner.start();

    try {
      const res = await fetch(endpoint, { method: "POST", body: formData });
      const data = await res.json();

      setTimeout(() => {
        spinner.stop();
        // modalImportFooter.classList.add("hidden");

        if (res.ok) {
          messenger.success(data.message, () => location.reload());
        } else {
          repoInput.style.display = "none";
          messenger.error(data.error);
        }
      }, 1200);
    } catch (err) {
      spinner.stop();
      console.error("ImportManager: import failed", err);
      alert("Import failed. Check console for details.");
    }
  },
};

export default ImportManager;
