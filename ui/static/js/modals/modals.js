// modals.js

import AboutApp from "./about.js";
import BookmarkCard from "./card.js";
import BookmarkDetail from "./detail.js";
import HelpApp from "./help.js";
import ImportManager from "./import.js";
import Manager from "./manager.js";
import Nav from "./nav.js";
import QRCode from "./qrcode.js";
import Repository from "./repo.js";
import SettingsApp from "./settings.js";
import SideMenu from "./sidemenu.js";

/**
 * UI utilities for showing modals, side menus, and dropdowns.
 */
const Modal = {
  /**
   * Opens a modal specified by ID.
   * @param {string} id - modal id
   */
  open(id) {
    const modal = document.getElementById(id);
    if (!modal) {
      console.error(`ModalShow selector: '${id}' not found`);
      return;
    }

    Manager.register(modal).open();
  },

  AboutApp,
  BookmarkCard,
  BookmarkDetail,
  HelpApp,
  ImportManager,
  Manager,
  Nav,
  QRCode,
  Repository,
  SettingsApp,
  SideMenu,
};

export default Modal;
