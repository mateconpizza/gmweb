// clipboard.js

/**
 * @typedef {object} ClipboardApi
 * @property {() => Promise<string>} read - Reads text from the clipboard.
 * @property {(text: string) => void} copy - Copies the given text to the clipboard.
 */

/** @type {ClipboardApi} */
const clipboard = {
  /**
   * Reads text from the clipboard.
   * @throws {Error} If the Clipboard API is not supported.
   * @returns {Promise<string>} A promise that resolves to the text from the clipboard.
   */
  read: async () => {
    if (!navigator.clipboard) {
      throw new Error("Clipboard API not supported");
    }
    return await navigator.clipboard.readText();
  },

  /**
   * Copies the given text to the clipboard.
   * @param {string} text - The text to be copied.
   */
  copy: (text) => {
    if (!navigator.clipboard) {
      console.error("Clipboard API not supported by this browser.");
      return;
    }
    navigator.clipboard
      .writeText(text)
      .then(() => {
        console.log("Text successfully copied to clipboard!");
      })
      .catch((err) => {
        console.error("Failed to copy text: ", err);
      });
  },
};

export default clipboard;
