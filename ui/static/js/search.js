// search.js

import config from "./config.js";
import { tagOps } from "./tags.js";

const keyboard = config.keyboard;
const keybinds = keyboard.keybinds;

/**
 * Handles the visibility of the keybind tip for the search input field.
 * @returns {void} This function does not return a value.
 */
export function keybindTipHandler() {
  // FIX: drop this...
  const searchInput = document.querySelector('.search-bar input[type="text"]');
  const keybindTip = document.querySelector(".keybind-tip");

  if (searchInput && keybindTip) {
    // Initial check on page load: hide tip if input has a value
    if (searchInput.value.length > 0) {
      keybindTip.style.opacity = "0";
    }

    // Add event listener to the search input to hide/show the tip
    searchInput.addEventListener("input", () => {
      if (searchInput.value.length > 0) {
        keybindTip.style.opacity = "0";
      } else {
        keybindTip.style.opacity = "1";
      }
    });

    searchInput.addEventListener("focus", () => {
      keybindTip.style.opacity = "0";
    });

    searchInput.addEventListener("blur", () => {
      if (searchInput.value.length === 0) {
        keybindTip.style.opacity = "1";
      }
    });

    document.addEventListener("keydown", (e) => {
      const isModalOpen = document.querySelector(".modal.show");
      if (isModalOpen) return;

      if (!keyboard.vimMode) {
        if ((e.ctrlKey || e.metaKey) && e.key === keybinds.search.focus.key) {
          e.preventDefault();
          searchInput.focus();
          searchInput.select();
        }
      }

      // VimMode
      if (e.key === keybinds.search.search.key) {
        e.preventDefault();
        searchInput.focus();
        searchInput.select();
      }

      if (e.key === keybinds.utility.escape.key) {
        if (document.activeElement === searchInput) {
          searchInput.blur();
        }
      }
    });
  }
}

/**
 * Class representing a search autocomplete feature.
 *
 * This class manages the search input, displays tag suggestions,
 * and handles user interactions for searching and selecting tags.
 */
export class InputCmp {
  constructor() {
    this.container = document.getElementById("search-container");
    this.input = this.container.querySelector('input[name="q"]');
    this.dropdown = this.container.querySelector("#tag-cmp-search-bar");
    this.form = this.container.querySelector(".search-bar");
    this.clearBtn = this.container.querySelector(".clear-search-btn");

    this.selectedIndex = -1;
    this.isTagMode = false;
    this.currentSuggestions = [];
    this.availableTags = [];

    this.bindEvents();

    tagOps.fetch().then((tags) => {
      if (tags) {
        this.availableTags = tags;
        console.log("Tags found:", tags.length);

        if (this.input.value.startsWith("#") && document.activeElement === this.input) {
          const tagQuery = this.input.value.substring(1).toLowerCase();
          this.showTagSuggestions(tagQuery);
        }
      } else {
        this.dropdown.innerHTML = '<div class="no-results">Error loading tags. Please try again.</div>';
        this.showDropdown();
      }
    });
  }

  bindEvents() {
    this.input.addEventListener("input", (e) => {
      this.handleInput(e);
      this.toggleClearButton();
    });
    this.input.addEventListener("keydown", (e) => this.handleKeydown(e));
    this.input.addEventListener("focus", () => this.handleFocus());

    this.clearBtn.addEventListener("click", () => {
      this.clearSearch();
    });

    // Handle clicks outside
    document.addEventListener("click", (e) => {
      if (!this.form.contains(e.target)) {
        this.hideDropdown();
      }
    });

    // Initial check for clear button
    this.toggleClearButton();
  }

  handleInput(e) {
    const value = e.target.value.trim();
    this.isTagMode = value.startsWith("#");

    if (this.isTagMode) {
      const tagQuery = value.substring(1).toLowerCase();

      // If it's just "#", suppress dropdown and show tip instead
      if (tagQuery.length === 0) {
        this.hideDropdown();
        const tipEl = document.querySelector(".keybind-tip");
        if (tipEl) tipEl.style.opacity = "1"; // show tip
        return;
      }

      // Hide tip once we actually type after #
      const tipEl = document.querySelector(".keybind-tip");
      if (tipEl) tipEl.style.opacity = "0";

      this.showTagSuggestions(tagQuery);
    } else {
      this.hideDropdown();
      const tipEl = document.querySelector(".keybind-tip");
      if (tipEl) tipEl.style.opacity = "0";
    }
  }

  handleKeydown(e) {
    if (!this.dropdown.style.display || this.dropdown.style.display === "none") {
      if (e.key === "Enter") {
        // e.preventDefault();
        this.form.submit();
      }
      return;
    }

    switch (e.key) {
      case "ArrowDown":
      case keybinds.dropdown.downCtrl.key:
        e.preventDefault();
        if (e.ctrlKey && e.key === keybinds.dropdown.downCtrl.key) {
          this.moveSelection(1);
        } else if (e.key === "ArrowDown") {
          this.moveSelection(1);
        }
        break;

      case "ArrowUp":
      case keybinds.dropdown.upCtrl.key:
        e.preventDefault();
        if (e.ctrlKey && e.key === keybinds.dropdown.upCtrl.key) {
          this.moveSelection(-1);
        } else if (e.key === "ArrowUp") {
          this.moveSelection(-1);
        }
        break;

      case "Enter":
      case keybinds.dropdown.accept.key:
        if (this.selectedIndex >= 0 || (e.ctrlKey && e.key === keybinds.dropdown.accept.key)) {
          this.selectCurrent();
        } else {
          this.form.submit();
        }
        break;

      case "Tab": // Add Tab key support
        e.preventDefault();
        if (e.shiftKey) {
          this.moveSelection(-1); // Shift+Tab moves up
        } else {
          this.moveSelection(1); // Tab moves down
        }
        break;

      case "Escape":
        this.hideDropdown();
        break;
    }
  }

  handleFocus() {
    const value = this.input.value.trim();
    if (this.isTagMode) {
      const tagQuery = value.substring(1).toLowerCase();
      this.showTagSuggestions(tagQuery);
    }
  }

  showTagSuggestions(query) {
    const filtered = this.availableTags.filter((tag) => tag.name.toLowerCase().includes(query)).slice(0, 10);
    this.currentSuggestions = filtered;
    this.selectedIndex = -1;
    this.renderTagSuggestions(filtered);
    this.showDropdown();
  }

  renderTagSuggestions(suggestions) {
    if (suggestions.length === 0) {
      this.dropdown.innerHTML = '<div class="no-results">No tags found</div>';
      return;
    }

    this.dropdown.innerHTML = suggestions
      .map(
        (tag, index) => `
                    <div class="tag-autocmp-item" data-index="${index}" data-tag="${tag.name}">
                        <svg class="tag-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
                            <line x1="7" y1="7" x2="7.01" y2="7"/>
                        </svg>
                        <span class="tag-text">#${tag.name}</span>
                        <span class="tag-count">${tag.count}</span>
                    </div>
                `,
      )
      .join("");

    // Add click handlers
    this.dropdown.querySelectorAll(".tag-autocmp-item").forEach((item) => {
      item.addEventListener("click", () => {
        const tagName = item.dataset.tag;
        this.selectTag(tagName);
      });
    });
  }

  moveSelection(direction) {
    const items = this.dropdown.querySelectorAll(".tag-autocmp-item");
    if (items.length === 0) return;

    // Remove current highlight
    items.forEach((item) => item.classList.remove("highlighted"));

    // Update selection index
    this.selectedIndex += direction;

    if (this.selectedIndex < 0) {
      this.selectedIndex = items.length - 1;
    } else if (this.selectedIndex >= items.length) {
      this.selectedIndex = 0;
    }

    // Add highlight to new selection
    items[this.selectedIndex].classList.add("highlighted");

    // Scroll into view if needed
    items[this.selectedIndex].scrollIntoView({ block: "nearest" });
  }

  selectCurrent() {
    if (this.selectedIndex >= 0 && this.currentSuggestions[this.selectedIndex]) {
      if (this.isTagMode) {
        const tag = this.currentSuggestions[this.selectedIndex];
        this.selectTag(tag.name);
      }
    }
  }

  selectTag(tagName) {
    this.input.value = `${tagName}`;
    this.input.name = "tag";
    this.hideDropdown();
    this.form.submit();
  }

  showDropdown() {
    this.dropdown.style.display = "block";
  }

  hideDropdown() {
    this.dropdown.style.display = "none";
    this.selectedIndex = -1;
  }

  clearSearch() {
    this.input.value = "";
    this.hideDropdown();
    this.toggleClearButton();
    this.input.focus();
  }

  toggleClearButton() {
    if (this.input.value.length > 0) {
      this.clearBtn.classList.add("show-clear-button");
    } else {
      this.clearBtn.classList.remove("show-clear-button");
    }
  }

  matchesKeybind(e, bind) {
    if (bind.ctrl && !e.ctrlKey) return false;
    if (bind.shift && !e.shiftKey) return false;
    if (bind.alt && !e.altKey) return false;
    if (bind.meta && !e.metaKey) return false;
    return e.key === bind.key;
  }
}
