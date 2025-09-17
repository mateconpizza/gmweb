/**
 * @module utils
 *
 * This module provides a collection of utility functions.
 */

import cleanURL from "./cleaner.js";
import clipboard from "./clipboard.js";

const utils = {
  /**
   * Creates a debounced function that delays its execution.
   * @param {Function} func The function to debounce.
   * @param {number} delay The number of milliseconds to delay.
   * @returns {Function} A new debounced function.
   */
  debounce(func, delay) {
    let timeout;
    return function (...args) {
      const context = this;
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(context, args), delay);
    };
  },

  /**
   * Adjusts the height of all textareas to fit their content.
   * The height is automatically resized when the page loads
   * and whenever the user types in the textarea.
   */
  resizeAllTextArea() {
    return;
    const textareas = document.querySelectorAll("textarea");
    textareas.forEach((textarea) => {
      /** @private */
      function adjustTextareaHeight() {
        textarea.style.height = "auto";
        textarea.style.height = textarea.scrollHeight + "px";
      }

      adjustTextareaHeight();
      textarea.addEventListener("input", adjustTextareaHeight);
    });
  },

  /**
   * Attaches an auto-resize handler to a textarea element.
   * Ensures the listener is only attached once per element.
   * @param {HTMLTextAreaElement} textarea -
   */
  resizeTextArea(textarea) {
    if (!textarea || !(textarea instanceof HTMLTextAreaElement)) return;

    // Skip if listener already attached
    if (textarea._hasResizeListener) return;
    /** @private */
    function adjustTextareaHeight() {
      textarea.style.height = "auto";
      textarea.style.height = textarea.scrollHeight + "px";
    }

    // Initial resize
    adjustTextareaHeight();

    // Attach once
    textarea.addEventListener("input", adjustTextareaHeight);
    textarea._hasResizeListener = true;
  },

  /**
   * Removes a suffix from a string if it exists.
   * @param {string} inputStr The string to remove the suffix from.
   * @param {string} suffix The suffix to find and remove.
   * @returns {string} The resulting string without the suffix.
   */
  stripSuffix(inputStr, suffix) {
    const regex = new RegExp(suffix + "$");
    return inputStr.replace(regex, "");
  },

  /**
   * Formats a Unix timestamp into a string like "Mar. 2, 2025, 10:08 p.m.".
   * @param {string|number} timestamp The Unix timestamp in seconds.
   * @returns {string} The formatted date and time string.
   */
  formatTimestamp(timestamp) {
    const year = parseInt(String(timestamp).substring(0, 4));
    const month = parseInt(String(timestamp).substring(4, 6)) - 1;
    const day = parseInt(String(timestamp).substring(6, 8));
    const hour = parseInt(String(timestamp).substring(8, 10));
    const minute = parseInt(String(timestamp).substring(10, 12));
    const second = parseInt(String(timestamp).substring(12, 14));

    const date = new Date(year, month, day, hour, minute, second);
    const monthNames = ["Jan.", "Feb.", "Mar.", "Apr.", "May", "Jun.", "Jul.", "Aug.", "Sep.", "Oct.", "Nov.", "Dec."];
    const formattedMonth = monthNames[date.getMonth()];
    const formattedDay = date.getDate();
    const formattedYear = date.getFullYear();

    let hours = date.getHours();
    const minutes = date.getMinutes();
    const ampm = hours >= 12 ? "p.m." : "a.m.";
    hours = hours % 12;
    hours = hours ? hours : 12;
    const formattedTime = `${hours}:${String(minutes).padStart(2, "0")} ${ampm}`;

    return `${formattedMonth} ${formattedDay}, ${formattedYear}, ${formattedTime}`;
  },

  /**
   * Hides the form message.
   * @private
   * @param {HTMLElement} mesgDiv The HTML element.
   * @returns {void}
   */
  hideFormMessage(mesgDiv) {
    if (!mesgDiv) {
      console.error("mesgDiv:", mesgDiv);
      return;
    }
    mesgDiv.textContent = "";
    mesgDiv.style.display = "none";
    mesgDiv.classList.add("hidden");
  },

  /**
   * Shows the form message.
   * @private
   * @param {HTMLElement} mesgDiv The HTML element.
   * @param {string} message The error message to display.
   * @returns {void}
   */
  showFormMessage(mesgDiv, message) {
    if (!mesgDiv) {
      console.error("mesgDiv:", mesgDiv, "message:", message);
      return;
    }
    mesgDiv.textContent = message;
    mesgDiv.style.display = "block";
    mesgDiv.classList.remove("hidden");
  },

  /**
   * Creates a spinner controller for a button.
   * @param {HTMLButtonElement} ele - The button element to control.
   * @param {boolean} restoreDim - Whether to restore the button's dimensions after loading.
   * @returns {{ start: () => void, stop: () => void }} Methods.
   */
  createBtnSpinner(ele, restoreDim = true) {
    if (!ele.dataset.originalContent) {
      ele.dataset.originalContent = ele.innerHTML;
    }

    const start = () => {
      if (restoreDim) {
        ele.style.width = ele.offsetWidth + "px";
        ele.style.height = ele.offsetHeight + "px";
      }
      ele.disabled = true;
      ele.classList.add("is-loading");
      ele.innerHTML = '<span class="loading"></span>';
    };

    const stop = () => {
      ele.disabled = false;
      ele.classList.remove("is-loading");
      ele.style.width = "";
      ele.style.height = "";
      if (ele.dataset.originalContent) {
        ele.innerHTML = ele.dataset.originalContent;
      }
    };
    return { start, stop };
  },

  /**
   * Creates a form messenger for displaying success and error messages.
   * @param {HTMLElement} successDiv - The element to display success messages.
   * @param {HTMLElement} errorDiv - The element to display error messages.
   * @returns {{ error: () => void, success: () => void, hide: () => void }} Methods.
   */
  createFormMessenger(successDiv, errorDiv) {
    const error = (message) => {
      utils.hideFormMessage(successDiv);
      utils.showFormMessage(errorDiv, message);
    };

    const success = (message, fn) => {
      utils.hideFormMessage(errorDiv);
      utils.showFormMessage(successDiv, message);
      setTimeout(() => {
        utils.hideFormMessage(successDiv);
        if (fn) fn();
      }, 2000);
    };

    const hide = () => {
      utils.hideFormMessage(errorDiv);
      utils.hideFormMessage(successDiv);
    };
    return { error, success, hide };
  },

  /**
   * Checks if a given string is a valid URL.
   * @param {string} urlString The URL to validate.
   * @returns {boolean} True if the URL is valid, false otherwise.
   */
  isValidUrl(urlString) {
    try {
      new URL(urlString);
      return true;
    } catch {
      console.warn("Invalid URL");
      return false;
    }
  },

  cleanURL: cleanURL,
  clipboard: clipboard,
};

export default utils;
