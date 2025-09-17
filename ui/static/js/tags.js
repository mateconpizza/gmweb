// tags.js

import repo from "./repo.js";
import routes from "./services/routes.js";

/**
 * TagAutocomplete provides autocomplete suggestions for tags inside an input field.
 * @param {HTMLInputElement} inputEl - The text input element where the user types tags.
 * @param {HTMLElement} dropdownEl - The container element for rendering suggestions.
 * @param {HTMLFormElement} formEl - The form element that contains the input.
 * @param {(selectedTag: string, lastQuery: string) => void} onTagSelect - Callback invoked when a tag is selected.
 */
export class TagAutocomplete {
  constructor(inputEl, dropdownEl, formEl, onTagSelect) {
    this.input = inputEl;
    this.dropdown = dropdownEl;
    this.onTagSelect = onTagSelect;
    this.selectedIndex = -1;
    this.currentSuggestions = [];
    this.form = formEl;
    this.currentTagQuery = "";

    tagOps.fetch().then((tags) => {
      this.availableTags = tags || [];
    });

    this.input.addEventListener("input", () => this.handleInput());
    this.input.addEventListener("keydown", (event) => this.handleKeyPressed(event));
    this.input.addEventListener("focus", () => this.handleInput());

    document.addEventListener("click", (event) => {
      if (!this.dropdown.contains(event.target) && event.target !== this.input) {
        this.hideDropdown();
      }
    });
  }

  handleInput() {
    const value = this.input.value;
    const parts = value.split(/[\s,]+/).filter(Boolean);
    const lastPart = parts[parts.length - 1] || "";

    this.currentTagQuery = lastPart.toLowerCase();

    if (this.currentTagQuery.length === 0) {
      this.hideDropdown();
      return;
    }

    if (this.currentTagQuery.length < 1) {
      this.hideDropdown();
      return;
    }

    const filtered = this.availableTags
      .filter((tag) => tag.name.toLowerCase().includes(this.currentTagQuery))
      .slice(0, 10);

    this.currentSuggestions = filtered;
    this.selectedIndex = -1;
    this.renderSuggestions(filtered);
    this.showDropdown();
  }

  handleKeyPressed(event) {
    if (!this.dropdownIsVisible()) return;

    switch (event.key) {
      case "ArrowDown":
        event.preventDefault();
        this.moveSelection(1);
        break;
      case "ArrowUp":
        event.preventDefault();
        this.moveSelection(-1);
        break;
      case "Enter":
        event.preventDefault(); // <== SIEMPRE prevenir submit si el dropdown está abierto
        if (this.selectedIndex >= 0) {
          this.selectCurrent();
        }
        break;
      case "Tab":
        if (this.dropdownIsVisible()) {
          event.preventDefault();
          if (this.selectedIndex >= 0) {
            this.selectCurrent();
          } else {
            this.moveSelection(event.shiftKey ? -1 : 1);
          }
        }
        break;
      case "Escape":
        this.hideDropdown();
        break;
    }
  }

  dropdownIsVisible() {
    return this.dropdown.style.display && this.dropdown.style.display !== "none";
  }

  moveSelection(dir) {
    const items = this.dropdown.querySelectorAll(".tag-autocmp-item");
    if (!items.length) return;

    items.forEach((item) => item.classList.remove("highlighted"));

    this.selectedIndex = (this.selectedIndex + dir + items.length) % items.length;

    const selected = items[this.selectedIndex];
    selected.classList.add("highlighted");
    selected.scrollIntoView({ block: "nearest" });
  }

  renderSuggestions(tags) {
    // FIX: when added a tag, or when editing the bookmark and focus the tags input, it will show `No tags found...`
    if (!tags.length && this.currentTagQuery.length > 0) {
      this.dropdown.innerHTML = '<div class="no-results">No tags found, it will created.</div>';
      this.showDropdown();
      return;
    }

    if (!tags.length) {
      this.hideDropdown();
      return;
    }

    this.dropdown.innerHTML = tags
      .map(
        (tag, index) => `
                    <div class="tag-autocmp-item" data-index="${index}" data-tag="${tag.name}">
                        <svg class="tag-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
                            <line x1="7" y1="7" x2="7.01" y2="7"/>
                        </svg>
                        <span class="tag-text">${tag.name}</span>
                        <span class="tag-count">${tag.count}</span>
                    </div>
                `,
      )
      .join("");

    this.dropdown.querySelectorAll(".tag-autocmp-item").forEach((el) => {
      el.addEventListener("click", () => {
        const name = el.dataset.tag;
        this.selectTag(name);
      });
    });
  }

  selectCurrent() {
    if (this.selectedIndex >= 0 && this.currentSuggestions[this.selectedIndex]) {
      const tag = this.currentSuggestions[this.selectedIndex];
      this.selectTag(tag.name);
    }
  }

  selectTag(tagName) {
    this.onTagSelect(tagName, this.currentTagQuery);
    this.hideDropdown();
  }

  showDropdown() {
    this.dropdown.style.display = "block";
    this.dropdown.style.width = `${this.input.offsetWidth}px`;
    this.input.classList.add("dropdown-open");
  }

  hideDropdown() {
    this.dropdown.style.display = "none";
    this.selectedIndex = -1;
    this.input.classList.remove("dropdown-open");
  }
}

export const tagOps = {
  /**
   * Fetches and processes tags from the server for the current database.
   * @async
   * @returns {Promise<Array<object>>} A promise that resolves to a sorted array of tag objects,
   * each with 'name' (string) and 'count' (number) properties, or `undefined` if an error occurs.
   */
  fetch: async () => {
    try {
      const response = await fetch(routes.api.listTags(repo.getCurrent()));
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const tagsMap = await response.json(); // { "golang": 12, "python": 5 }
      const transformedTags = Object.entries(tagsMap).map(([name, count]) => ({
        name,
        count,
      }));

      transformedTags.sort((tagA, tagB) => tagA.name.localeCompare(tagB.name));
      return transformedTags;
    } catch (error) {
      console.error("Error fetching tags for autocomplete:", error);
    }
  },

  /**
   * Formats a string of tags into a unique, sorted, and comma-separated string.
   * @param {string} tags The raw string of tags to be formatted.
   * @returns {string} The formatted string of unique, sorted tags
   */
  format: (tags) => {
    const seen = new Set();
    return tags
      .split(/[\s,]+/)
      .map((tag) => tag.trim().replace(/^#/, ""))
      .filter((tag) => tag && !seen.has(tag) && seen.add(tag))
      .sort((tagA, tagB) => tagA.localeCompare(tagB))
      .map((tag) => `${tag}`)
      .join(", ");
  },
};
