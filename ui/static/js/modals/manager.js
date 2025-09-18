// manager.js
import config from "../config.js";

/**
 * @typedef {object} ModalController
 * @property {() => void} open - Opens the modal.
 * @property {() => void} close - Closes the modal.
 * @property {() => void} toggle - Toggles the modal's visibility.
 * @property {() => boolean} isOpen -
 */

// Global modal stack
/** @type {HTMLElement[]} */
const modalStack = [];

let escapeListenerAttached = false;

/**
 * Retrieves the top modal from the modal stack.
 * @function getTopModal
 * @returns {HTMLElement|null} The top modal element or null if no modals are present.
 */
function getTopModal() {
  return modalStack.at(-1) || null;
}

/**
 * Removes the specified modal from the modal stack and hides it.
 * @param {object} modal The modal element to remove from the stack and hide.
 */
function cleanupModalState(modal) {
  // Remove from stack if present
  const index = modalStack.indexOf(modal);
  if (index !== -1) {
    modalStack.splice(index, 1);
  }

  // Ensure modal is hidden
  modal.classList.remove("show");
}

/**
 * Sets up event handlers for closing a modal.
 * @function setupModalCloseHandlers
 * @param {HTMLElement} modal - The modal element to which the close handlers will be attached.
 * @param {Function} closeModalFn - The function to call when the modal should be closed.
 * @returns {void}
 */
function setupModalCloseHandlers(modal, closeModalFn) {
  if (modal.dataset.closers === "true") return;
  modal.dataset.closers = "true";

  // Backdrop click
  modal.addEventListener("click", (e) => {
    if (e.target === modal) closeModalFn();
  });

  // Escape & Close key – global once
  if (!escapeListenerAttached) {
    document.addEventListener("keydown", (event) => {
      const isEscape = event.key === config.keyboard.keybinds.utility.escape.key;
      const isClose = event.key === config.keyboard.keybinds.utility.close.key;

      if (isEscape || isClose) {
        const active = document.activeElement;

        if (active && (active.tagName === "INPUT" || active.tagName === "TEXTAREA")) {
          active.blur(); // lose focus instead of closing modal
          return;
        }

        const top = getTopModal();
        if (top) top.controller.close();
      }
    });
    escapeListenerAttached = true;
  }

  // Close button
  const closeBtn = modal.querySelector(".close-btn");
  if (closeBtn) {
    closeBtn.addEventListener("click", () => {
      closeModalFn();
    });
  }

  // Back button
  const backBtn = modal.querySelector(".back-btn");
  if (backBtn) {
    backBtn.addEventListener("click", () => {
      closeModalFn();
    });
  }

  // Cancel button
  modal.querySelectorAll("#btn-cancel").forEach((btn) => {
    btn.addEventListener("click", closeModalFn);
  });
}

/**
 * Modal management object providing modal control functionality
 */
const Manager = {
  /**
   * Manages the visibility of a modal dialog.
   * @param {HTMLElement} modal - The modal element to control.
   * @returns {ModalController} Controller object.
   */
  register(modal) {
    if (modal.controller) {
      cleanupModalState(modal);
      return modal.controller;
    }

    const controller = {
      open: () => {
        // Hide the current top modal if it exists
        const topModal = getTopModal();
        if (topModal && topModal !== modal) {
          topModal.classList.remove("show");
        }

        // Only add to stack if not already present
        if (!modalStack.includes(modal)) {
          modalStack.push(modal);
        }

        modal.classList.add("show");
        document.body.style.overflow = "hidden";
      },

      close: () => {
        if (!modal.classList.contains("show")) return;

        modal.classList.remove("show");

        // Remove the specific modal from the stack
        const modalIndex = modalStack.indexOf(modal);
        if (modalIndex !== -1) {
          modalStack.splice(modalIndex, 1);
        }

        // Show the new top modal if it exists
        const newTopModal = getTopModal();
        if (newTopModal) {
          newTopModal.classList.add("show");
        }

        // If no modals are left, restore body scroll
        if (modalStack.length === 0) {
          document.body.style.overflow = "";
        } else {
          document.body.style.overflow = "hidden";
        }
      },

      toggle: () => {
        if (modal.classList.contains("show")) {
          controller.close();
        } else {
          controller.open();
        }
      },

      isOpen: () => modal.classList.contains("show"),
    };

    modal.controller = controller;
    setupModalCloseHandlers(modal, controller.close);
    return controller;
  },

  /**
   * Returns the modal currently visible to the user (top of stack).
   * @returns {HTMLElement|null} -
   */
  getCurrent() {
    return getTopModal();
  },
};

export default Manager;
