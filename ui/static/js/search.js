// search.js

import config from "./config.js";
import repo from "./repo.js";
import routes from "./services/routes.js";
import { tagOps } from "./tags.js";

const keyboard = config.keyboard;
const KEYBINDS = keyboard.keybinds;

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
        if ((e.ctrlKey || e.metaKey) && e.key === KEYBINDS.search.focus.key) {
          e.preventDefault();
          searchInput.focus();
          searchInput.select();
        }
      }

      // VimMode
      if (e.key === KEYBINDS.search.search.key) {
        e.preventDefault();
        searchInput.focus();
        searchInput.select();
      }

      if (e.key === KEYBINDS.utility.escape.key) {
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
    this.input = document.querySelector('input[name="q"]');
    this.dropdown = document.getElementById("tag-cmp-search-bar");
    this.form = document.querySelector(".search-bar");
    this.clearBtn = document.querySelector(".clear-search-btn");

    this.selectedIndex = -1;
    this.isTagMode = false;
    this.isBookmarkMode = false;
    this.currentSuggestions = [];
    this.bookmarks = [];
    this.availableTags = [];

    this.bindEvents();
    this.loadBookmarks();

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

  async loadBookmarks() {
    try {
      const response = await fetch(routes.api.listBookmarks(repo.getCurrent()), {
        method: "GET",
        headers: {
          Accept: "application/json",
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      // The data from the API is already the bookmarks array.
      const bookmarksData = await response.json();
      this.bookmarks = bookmarksData.map((bookmark) => {
        return {
          url: bookmark.url,
          title: bookmark.title,
          desc: bookmark.desc,
          tags: bookmark.tags,
          id: bookmark.id,
        };
      });

      console.log("Bookmarks loaded:", this.bookmarks.length);
    } catch (error) {
      console.error("Failed to load bookmarks:", error);
      this.bookmarks = [];
    }
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
    this.isBookmarkMode = !this.isTagMode && value.length > 0;

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
    } else if (this.isBookmarkMode) {
      // Hide tip for bookmark mode
      const tipEl = document.querySelector(".keybind-tip");
      if (tipEl) tipEl.style.opacity = "0";

      this.showBookmarkSuggestions(value.toLowerCase());
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

    const isCtrlOrMeta = e.ctrlKey || e.metaKey;

    switch (e.key) {
      case "ArrowDown":
        this.moveSelection(1);
        break;

      case "ArrowUp":
        this.moveSelection(-1);
        break;

      case KEYBINDS.dropdown.downCtrl.key:
        if (isCtrlOrMeta) {
          e.preventDefault();
          this.moveSelection(1);
        }
        break;

      case KEYBINDS.dropdown.upCtrl.key:
        if (isCtrlOrMeta) {
          e.preventDefault();
          this.moveSelection(-1);
        }
        break;

      case KEYBINDS.actions.enter.key:
      case KEYBINDS.dropdown.accept.key:
        if (this.selectedIndex >= 0 || (isCtrlOrMeta && e.key === "y")) {
          e.preventDefault();
          this.selectCurrent();
        } else if (e.key === "Enter") {
          this.form.submit();
        }
        break;

      case KEYBINDS.dropdown.tab.key:
        e.preventDefault();
        if (e.shiftKey) {
          this.moveSelection(-1); // Shift+Tab moves up
        } else {
          this.moveSelection(1); // Tab moves down
        }
        break;

      case KEYBINDS.utility.escape.key:
        this.hideDropdown();
        break;
    }
  }

  handleFocus() {
    const value = this.input.value.trim();
    if (this.isTagMode) {
      const tagQuery = value.substring(1).toLowerCase();
      this.showTagSuggestions(tagQuery);
    } else if (value.length > 0) {
      this.showBookmarkSuggestions(value.toLowerCase());
    }
  }

  showTagSuggestions(query) {
    const filtered = this.availableTags.filter((tag) => tag.name.toLowerCase().includes(query)).slice(0, 10);
    this.currentSuggestions = filtered;
    this.selectedIndex = -1;
    this.renderTagSuggestions(filtered);
    this.showDropdown();
  }

  showBookmarkSuggestions(query) {
    const q = query.toLowerCase();
    const filtered = this.bookmarks
      .filter((bookmark) => {
        const title = bookmark.title?.toLowerCase() || "";
        const desc = bookmark.desc?.toLowerCase() || "";
        const url = bookmark.url?.toLowerCase() || "";
        const tags = bookmark.tags?.toLowerCase() || "";

        return title.includes(q) || desc.includes(q) || url.includes(q) || tags.includes(q);
      })
      .slice(0, 8);

    this.currentSuggestions = filtered;
    this.selectedIndex = -1;
    this.renderBookmarkSuggestions(filtered);
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

  renderBookmarkSuggestions(suggestions) {
    if (suggestions.length === 0) {
      this.dropdown.innerHTML = '<div class="no-results">No bookmarks found</div>';
      return;
    }

    this.dropdown.innerHTML = suggestions
      .map((bookmark, index) => {
        const displayUrl = this.truncateUrl(bookmark.url);
        const displayTitle = this.truncateText(bookmark.title, 50);

        return `
          <div class="bookmark-autocmp-item" data-index="${index}" data-url="${bookmark.url}">
            <div class="bookmark-cmp-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
              </svg>
            </div>
            <div class="bookmark-cmp-content">
              <div class="bookmark-cmp-title">${displayTitle}</div>
              <div class="bookmark-cmp-url">${displayUrl}</div>
            </div>
          </div>
        `;
      })
      .join("");

    // Add click handlers
    this.dropdown.querySelectorAll(".bookmark-autocmp-item").forEach((item) => {
      item.addEventListener("click", () => {
        const url = item.dataset.url;
        this.selectBookmark(url);
      });
    });
  }

  truncateUrl(url) {
    try {
      const urlObj = new URL(url);
      const domain = urlObj.hostname.replace("www.", "");
      const path = urlObj.pathname;

      if (path.length > 25) {
        return `${domain}${path.substring(0, 22)}...`;
      }
      return `${domain}${path}`;
    } catch {
      return url.length > 40 ? url.substring(0, 37) + "..." : url;
    }
  }

  truncateText(text, maxLength) {
    return text.length > maxLength ? text.substring(0, maxLength - 3) + "..." : text;
  }

  moveSelection(direction) {
    const items = this.dropdown.querySelectorAll(".tag-autocmp-item, .bookmark-autocmp-item");
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
      } else if (this.isBookmarkMode) {
        const bookmark = this.currentSuggestions[this.selectedIndex];
        this.selectBookmark(bookmark.url);
      }
    }
  }

  selectTag(tagName) {
    this.input.value = `${tagName}`;
    this.input.name = "tag";
    this.hideDropdown();
    this.form.submit();
  }

  selectBookmark(url) {
    window.open(url, "_blank");
    this.hideDropdown();
    this.input.value = '';
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
