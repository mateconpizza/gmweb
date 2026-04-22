// modal.js

import config from "../config.js";
import HelpApp from "../modals/help.js";
import Manager from "../modals/manager.js";

const KEYBINDS = config.keyboard.keybinds;

export default class ModalNavigator {
  /**
   * @param {HTMLElement} modalElement The modal element to navigate.
   * @param {object} customKeybinds Custom keybinds configuration for this modal.
   */
  constructor(modalElement, customKeybinds = {}) {
    this.modal = modalElement;
    this.scrollableContent = this.modal.querySelector(".modal-base");
    this.customKeybinds = customKeybinds;
    this.helpModal = null;
    this.initializeKeybinds();
    this.bindEvents();
    this.updateKeybindHints(); // Update both input and accordion keybinds
  }

  initializeKeybinds() {
    // Default keybinds - you can customize these
    this.defaultKeybinds = {
      navigation: {
        down: KEYBINDS.navigation.down,
        up: KEYBINDS.navigation.up,
      },
      utility: {
        help: KEYBINDS.utility.help,
      },
    };

    // Merge default and custom keybinds
    this.keybinds = this.mergeKeybinds(this.defaultKeybinds, this.customKeybinds);
  }

  mergeKeybinds(defaultKeybinds, customKeybinds) {
    if (customKeybinds.length == 0) return;

    const merged = JSON.parse(JSON.stringify(defaultKeybinds));

    for (const category in customKeybinds) {
      if (merged[category]) {
        Object.assign(merged[category], customKeybinds[category]);
      } else {
        merged[category] = customKeybinds[category];
      }
    }

    return merged;
  }

  bindEvents() {
    this.keydownHandler = this.handleKeydown.bind(this);
    document.addEventListener("keydown", this.keydownHandler);
  }

  unbindEvents() {
    document.removeEventListener("keydown", this.keydownHandler);
  }

  handleKeydown(event) {
    // Only listen if the modal is currently open and active
    const modalOpen = this.modal?.classList.contains("show");
    const helpOpen = this.helpModal?.classList.contains("show");

    if (!modalOpen && !helpOpen) {
      this.unbindEvents();
      return;
    }

    // Don't trigger keybinds if user is typing in an input/textarea
    if (this.isTypingInInput(event.target)) {
      return;
    }

    const key = event.key;
    console.log({ key });

    if (key === this.keybinds.utility.help?.key) {
      this.handleUtilityKeybinds(event, key);
      return;
    }

    if (!this.scrollableContent) {
      console.warn("ModalNavigator: Scrollable content element not found.");
      return;
    }

    // -- Navigation
    // Handle 'down'
    if (key === this.keybinds.navigation.down.key) return this.goDown(event);
    // Handle 'up'
    if (key === this.keybinds.navigation.up.key) return this.goUp(event);

    // Handle default utility keybinds
    this.handleUtilityKeybinds(event, key);

    // Handle custom keybinds
    this.handleCustomKeybinds(event, key);
  }

  isTypingInInput(target) {
    const inputTypes = ["input", "textarea", "select"];
    return inputTypes.includes(target.tagName.toLowerCase()) || target.contentEditable === "true";
  }

  handleUtilityKeybinds(event, key) {
    if (key === this.keybinds.utility.help?.key) {
      event.preventDefault();

      // Initialize help modal if not already done
      if (!this.helpModal) {
        this.helpModal = document.getElementById("modal-help-keybinds");
        if (!this.helpModal) {
          console.warn("ModalNavigator: Help modal element not found");
          return;
        }

        // Render the keybinds help content
        HelpApp.renderKeybindsHelp(this.helpModal, this.keybinds, {
          ignore: [KEYBINDS.utility.escape.key, KEYBINDS.utility.help.key],
        });
      }

      // Use the manager to toggle the help modal
      const controller = Manager.register(this.helpModal);
      controller.toggle();
    }
  }

  handleCustomKeybinds(event, key) {
    if (!this.keybinds) return;

    for (const category in this.keybinds) {
      for (const [fieldName, config] of Object.entries(this.keybinds[category])) {
        // focus input/textarea
        if (key === config.key && typeof config.focus == "boolean") {
          event.preventDefault();
          this.focusField(config.selector || `#${config.id}` || `[name="${fieldName}"]`);
          break;
        }
        //
        console.log({ key, config }, "key === config.key", key === config.key);
        if (key === config.key) {
          event.preventDefault();
          if (config.action && typeof config.action === "function") {
            config.action();
          } else if (config.selector) {
            // this.triggerClick(config.selector);
            const toggle = this.modal.querySelector(config.selector);
            if (toggle) {
              toggle.click();

              // Only focus if accordion just opened
              if (config.focus && toggle.getAttribute("aria-expanded") === "true") {
                const focusSelector = typeof config.focus === "string" ? config.focus : config.focus.selector;
                this.focusField(focusSelector);
              }
            }
          }
          break;
        }
      }
    }
  }

  handleCustomKeybindsOld(event, key) {
    // Handle focus keybinds
    if (this.keybinds.focus) {
      for (const [fieldName, config] of Object.entries(this.keybinds.focus)) {
        if (key === config.key) {
          event.preventDefault();
          this.focusField(config.selector || `#${config.id}` || `[name="${fieldName}"]`);
          break;
        }
      }
    }

    // Handle custom action keybinds
    if (this.keybinds.custom) {
      for (const [actionName, config] of Object.entries(this.keybinds.custom)) {
        if (key === config.key) {
          console.log({ actionName, config });
          event.preventDefault();
          if (config.action && typeof config.action === "function") {
            config.action();
          } else if (config.selector) {
            // this.triggerClick(config.selector);
            const toggle = this.modal.querySelector(config.selector);
            if (toggle) {
              toggle.click();

              // Only focus if accordion just opened
              if (config.focus && toggle.getAttribute("aria-expanded") === "true") {
                const focusSelector = typeof config.focus === "string" ? config.focus : config.focus.selector;
                this.focusField(focusSelector);
              }
            }
          }
          break;
        }
      }
    }
  }

  focusField(selector) {
    const field = this.modal.querySelector(selector);
    if (field) {
      field.focus();
      // If it's a text input/textarea, move cursor to end
      if (field.setSelectionRange && field.value) {
        field.setSelectionRange(field.value.length, field.value.length);
      }
    } else {
      console.warn(`ModalNavigator: Field with selector "${selector}" not found.`);
    }
  }

  triggerClick(selector) {
    const btn = this.modal.querySelector(selector);
    if (btn) {
      btn.click();
    } else {
      console.warn(`ModalNavigator: Button with selector "${selector}" not found.`);
    }
  }

  updateKeybindHints() {
    // Update input field keybinds
    this.updateInputKeybinds();

    // Update accordion toggle keybinds
    this.updateAccordionKeybinds();
  }

  updateInputKeybinds() {
    if (this.keybinds.focus) {
      for (const [fieldName, config] of Object.entries(this.keybinds.focus)) {
        const fieldSelector = config.selector || `[name="${fieldName}"]`;
        const field = this.modal.querySelector(fieldSelector);

        if (field) {
          // Look for the keybind element in the parent form-group
          const formGroup = field.closest(".form-group");
          const keybindElement = formGroup?.querySelector(".input-float-keybind kbd");

          if (keybindElement) {
            keybindElement.textContent = config.key;
            keybindElement.parentElement.classList.remove("hidden");
          } else {
            console.warn(`ModalNavigator: Keybind element not found for field "${fieldName}"`);
          }
        } else {
          console.warn(`ModalNavigator: Field with selector "${fieldSelector}" not found.`);
        }
      }
    }
  }

  updateAccordionKeybinds() {
    if (!this.keybinds) return;

    for (const category in this.keybinds) {
      for (const [, config] of Object.entries(this.keybinds[category])) {
        if (
          config.selector &&
          (config.selector.includes(".accordion-toggle") || config.selector.includes(".tag-toggle"))
        ) {
          const toggleButton = this.modal.querySelector(config.selector);
          if (toggleButton) {
            toggleButton.innerHTML = `<kbd>${config.key}</kbd>`;
            toggleButton.setAttribute("title", `Press '${config.key}' to toggle`);
          } else {
            console.warn(`ModalNavigator: Toggle not found for selector "${config.selector}"`);
          }
        }
      }
    }
  }

  // Navigation
  goDown(e) {
    e.preventDefault();
    this.scrollableContent.scrollBy({ top: 60, behavior: "smooth" });
    return;
  }
  goUp(e) {
    e.preventDefault();
    this.scrollableContent.scrollBy({ top: -60, behavior: "smooth" });
    return;
  }
}
