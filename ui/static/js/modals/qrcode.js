// qrcode.js

import Manager from "./manager.js";
import repo from "../repo.js";
import routes from "../services/routes.js";

const QRCode = {
  /**
   * Opens a QR code modal.
   * @param {number} id The bookmark id
   */
  open(id) {
    if (!id) {
      console.error("QRModal: bookmark ID is null");
      return;
    }

    const modal = document.getElementById("modal-qrcode");
    Manager.register(modal);

    const qrImage = document.getElementById("qr-image");
    const qrURL = routes.front.viewQrCode(repo.getCurrent(), id);

    this.load({
      qrImage: qrImage,
      qrURL: qrURL,
      onSuccess: () => {
        modal.controller.open();
      },
      onError: () => {
        modal.controller.open();
      },
    });
  },

  /**
   * Loads a QR code image and handles success/error events.
   * @param {object} params - Function parameters
   * @param {HTMLImageElement} params.qrImage - Image element to display QR code
   * @param {string} params.qrURL - URL to load QR code from
   * @param {Function} [params.onSuccess] - Callback on successful image load
   * @param {Function} [params.onError] - Callback on image load error
   */
  load({ qrImage, qrURL, onSuccess, onError }) {
    qrImage.src = "";
    qrImage.alt = "Generating QR Code...";

    qrImage.onload = () => {
      onSuccess?.();
    };

    qrImage.onerror = () => {
      console.error("Failed to load QR-Code Image from:", qrURL);
      qrImage.alt = "Failed to generate QR-Code";
      onError?.();
    };

    qrImage.src = qrURL;
    qrImage.alt = `QR Code for ${qrURL}`;
  },
};

export default QRCode;
