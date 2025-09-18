// app.js

import config from "./config.js";
import Modal from "./modals/modals.js";
import VimNavigator from "./navigation/global.js";
import repo from "./repo.js";
import { InputCmp, keybindTipHandler } from "./search.js";
import theme from "./theme.js";

const App = {
  /**
   * Bootstraps core features and config.
   */
  init() {
    if (config.dev.enabled()) {
      console.log("Running in development mode");
    }

    // navigation
    Modal.SettingsApp.restorePref();
    if (config.keyboard.vimMode) {
      new VimNavigator();
    }

    // search input bar autocompletion
    new InputCmp();
    keybindTipHandler();

    // repository
    repo.setupNewRepoBtn();

    // colorscheme
    theme.switcher();
  },

  /**
   * Initializes modal-related features.
   */
  setupModals() {
    const modals = [
      Modal.AboutApp,
      Modal.BookmarkCard,
      Modal.BookmarkDetail,
      Modal.ImportManager,
      Modal.Repository,
      Modal.SettingsApp,
      Modal.SideMenu,
      Modal.Nav,
    ];

    // VIM Mode
    if (config.keyboard.vimMode) {
      modals.push(Modal.HelpApp);
    }

    modals.forEach((m) => m.init());
  },
};

export default App;
