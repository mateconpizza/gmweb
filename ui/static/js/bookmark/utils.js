// params.js

import config from "../config.js";
import routes from "../services/routes.js";
import { tagOps } from "../tags.js";
import utils from "../utils/utils.js";

/**
 * Creates and manages a list of URL parameters, cleaning unnecessary ones and updating the DOM accordion accordingly.
 * It also provides interactive checkboxes to include or exclude parameters and a toggle-all functionality.
 * @param {HTMLInputElement} urlInput - The input element containing the URL to process.
 * @param {HTMLElement} accordion - The accordion DOM element to update with URL parameter information.
 */
async function createUrlParamsList(urlInput, accordion) {
  const initialURL = new URL(urlInput.value);
  const originalParams = new Map();
  initialURL.searchParams.forEach((value, key) => {
    originalParams.set(key, value);
  });

  const { deletedParams } = utils.cleanURL(urlInput.value);
  const spanUselessParams = accordion.querySelector("#url-useless-params");
  const toggleAllBtn = accordion.querySelector("#toggle-all-params");
  const accordionTitle = accordion.querySelector(".accordion-header").querySelector(".accordion-title");

  spanUselessParams.innerHTML = "";
  spanUselessParams.classList.add("hidden");

  if (deletedParams.length > 0) {
    accordion.classList.remove("hidden");
    spanUselessParams.classList.remove("hidden");
    accordionTitle.textContent = `URL Params (${deletedParams.length})`;
    console.log(`URL Params (${deletedParams.length})`);

    deletedParams.forEach((param) => {
      const containerSpan = document.createElement("span");
      containerSpan.className = "url-param";

      const tagContentSpan = document.createElement("span");
      tagContentSpan.className = "tag-text";
      tagContentSpan.textContent = param;
      containerSpan.appendChild(tagContentSpan);

      const checkbox = document.createElement("input");
      checkbox.type = "checkbox";
      checkbox.id = `param-checkbox-${param}`;
      checkbox.className = "minimal-checkbox";
      checkbox.title = "Keep/remove URL param";
      checkbox.checked = true;

      // Individual checkbox toggle
      checkbox.addEventListener("change", () => {
        try {
          const currentURL = new URL(urlInput.value);
          const searchParams = currentURL.searchParams;

          if (!checkbox.checked) {
            searchParams.delete(param);
          } else {
            const originalValue = originalParams.get(param);
            if (originalValue !== undefined) {
              searchParams.set(param, originalValue);
            }
          }

          currentURL.search = searchParams.toString();
          urlInput.value = currentURL.toString();

          // Update URL Params in accordionTitle
          const checkedCount = deletedParams.filter((p) => {
            const cb = document.getElementById(`param-checkbox-${p}`);
            return cb && cb.checked;
          }).length;
          accordionTitle.textContent = `URL Params (${checkedCount})`;
        } catch (error) {
          console.error("Error updating URL:", error);
        }
      });

      containerSpan.appendChild(checkbox);
      spanUselessParams.appendChild(containerSpan);
    });

    // Toggle All button logic
    if (toggleAllBtn) {
      toggleAllBtn.onclick = () => {
        const checkboxes = spanUselessParams.querySelectorAll("input[type='checkbox']");
        const allChecked = Array.from(checkboxes).every((cb) => cb.checked);

        checkboxes.forEach((cb) => {
          cb.checked = !allChecked;
          cb.dispatchEvent(new Event("change")); // Trigger same logic as manual click
        });

        toggleAllBtn.textContent = allChecked ? "Check All" : "Uncheck All";
      };
    }
  }
}

/**
 * Scrapes metadata (title, description, tags, favicon) from a given URL and populates form fields.
 * @async
 * @param {string} url The URL to scrape.
 * @param {HTMLInputElement} titleInput The input element for the title.
 * @param {HTMLTextAreaElement} descInput The input element for the description.
 * @param {HTMLInputElement} tagsInput The input element for the tags.
 * @param {HTMLInputElement} faviconInput The input element for the favicon URL.
 * @param {HTMLImageElement} faviconPreview The <img> element that previews the favicon.
 * @returns {Promise<void>}
 */
async function scrapeURLData(url, titleInput, descInput, tagsInput, faviconInput, faviconPreview) {
  if (!url) return;

  // Show loading state
  titleInput.value = "Scraping title...";
  tagsInput.value = "Scraping tags or keywords...";
  descInput.value = "Scraping description...";
  titleInput.disabled = tagsInput.disabled = descInput.disabled = true;

  [titleInput, tagsInput, descInput].forEach((el) => {
    el.classList.remove("fade-in");
    el.classList.add("loading-state");
  });

  try {
    const response = await fetch(`${routes.api.scrapeUrl}?url=${encodeURIComponent(url)}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": config.security.csrfToken(),
      },
    });

    if (response.ok) {
      const data = await response.json();

      // Title
      if (data.title && data.title !== "untitled (unfiled)") {
        titleInput.value = data.title;
        titleInput.classList.add("fade-in");
        utils.resizeTextArea(titleInput);
        setTimeout(() => titleInput.classList.remove("fade-in"), 500);
      } else {
        titleInput.value = "";
      }

      // Description
      if (data.desc) {
        descInput.value = data.desc;
        descInput.classList.add("fade-in");
        utils.resizeTextArea(descInput);
        setTimeout(() => descInput.classList.remove("fade-in"), 500);
      } else {
        descInput.value = "";
      }

      // Tags
      if (data.tags && Array.isArray(data.tags)) {
        tagsInput.value = tagOps.format(data.tags.join(", "));
        tagsInput.classList.add("fade-in");
        setTimeout(() => tagsInput.classList.remove("fade-in"), 500);
      } else {
        tagsInput.value = "";
      }

      // Favicon
      if (data.favicon_url) {
        faviconInput.value = data.favicon_url;
        if (faviconPreview) {
          faviconPreview.src = data.favicon_url;
          faviconPreview.onerror = () => {
            faviconPreview.src = config.static.favicon; // fallback
          };
        }
      }
    } else {
      console.error("Failed to fetch data:", response.status, await response.text());
      titleInput.value = "";
      tagsInput.value = "";
    }
  } catch (error) {
    console.error("Network or parsing error fetching URL data:", error);
    titleInput.value = tagsInput.value = descInput.value = "";
  } finally {
    // Reset state
    titleInput.disabled = tagsInput.disabled = descInput.disabled = false;
    [titleInput, tagsInput, descInput].forEach((el) => el.classList.remove("loading-state"));
  }
}

/**
 * Scrapes a specific type of data (e.g., title, description) from a URL and updates a target input field.
 * @async
 * @param {string} url The URL to scrape.
 * @param {HTMLInputElement|HTMLTextAreaElement} targetInput The input element to update with the scraped data.
 * @param {string} type The type of data to scrape (e.g., 'title', 'description').
 * @returns {Promise<void>}
 */
async function scrapeInput(url, targetInput, type) {
  const initialValue = targetInput.value;
  targetInput.value = `Scraping ${type}...`;

  try {
    const res = await fetch(`${routes.api.scrapeUrl}?url=${encodeURIComponent(url)}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    });

    if (res.ok) {
      const data = await res.json();
      console.log(data);

      if (data[type]) {
        targetInput.value = data[type];
      } else {
        targetInput.value = initialValue;
      }

      targetInput.classList.add("fade-in");
      utils.resizeTextArea(targetInput);
      setTimeout(() => targetInput.classList.remove("fade-in"), 500);
    } else {
      targetInput.value = initialValue;
      const errorText = await res.json();
      console.error("Failed to fetch data:", res.error, errorText);
    }
  } catch (error) {
    targetInput.value = initialValue;
    console.error("Network or parsing error fetching URL data:", error);
    targetInput.value = "";
  }
}

const disableInputs = (...inputs) => {
  inputs.forEach((input) => {
    input.disabled = true;
  });
};

const enableInputs = (...inputs) => {
  inputs.forEach((input) => {
    input.disabled = false;
  });
};

const bUtils = {
  enableInputs: enableInputs,
  disableInputs: disableInputs,
  createUrlParamsList: createUrlParamsList,
  scrapeInput: scrapeInput,
  scrapeURLData: scrapeURLData,
};

export default bUtils;
