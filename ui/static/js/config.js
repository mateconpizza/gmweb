// config.js

const config = {
  dbName: "main",
  static: {
    base: "/static",
    favicon: "/static/img/favicon-star.png",
  },
  security: {
    csrfToken: () => document.querySelector("input[name='csrf_token']")?.value,
  },
  dev: {
    enabled: () => window.DEV_MODE || false,
  },
  paths: {
    api: "/api",
    web: "/web",
  },
  keyboard: {
    vimMode: false,
    keybinds: {
      navigation: {
        up: { key: "k", description: "Navigate up", done: true },
        down: { key: "j", description: "Navigate down", done: true },
        top: { key: "gg", description: "Go to top", done: true },
        bottom: { key: "G", description: "Go to bottom", done: true },
        middle: { key: "M", description: "Go to the middle", done: true },
        pageNext: { key: "n", description: "Next page", done: true },
        pagePrev: { key: "p", description: "Prev page", done: true },
      },
      actions: {
        enter: { shortcut: "CR", key: "Enter", description: "Open highlighted", done: true },
        openTab: { key: "O", description: "Open in new tab", done: true },
        edit: { key: "e", description: "Edit bookmark", done: true },
        favorite: { key: "f", description: "Mark as favorite", done: true },
        newBookmark: { key: "a", description: "New bookmark", done: true },
        delete: { key: "D", description: "Delete bookmark", done: true },
        qrcode: { key: "c", description: "Show QR code", done: true },
        paste: { key: "P", description: "Add from clipboard", done: true },
        yank: { key: "Y", description: "Copy link to clipboard", done: true },
        selection: { key: "V", description: "Multi-selection", done: true },
      },
      search: {
        search: { key: "/", description: "Search bookmark", done: true },
        focus: { shortcut: "Ctrl-k", key: "k", description: "Focus search bar", done: true },
        tags: { key: "t", description: "Search by #tag", done: false },
      },
      dropdown: {
        upCtrl: { shortcut: "Ctrl-u", key: "u", description: "Navigate down", done: true },
        downCtrl: { shortcut: "Ctrl-d", key: "d", description: "Navigate up", done: true },
        accept: { shortcut: "Ctrl-y", key: "y", description: "Accept item", done: true },
        tab: { key: "Tab", description: "Next", done: true },
        tabShift: { key: "S-Tab", description: "Prev", done: true },
      },
      utility: {
        help: { key: "?", description: "Show help", done: true },
        settings: { key: "S", description: "Show settings", done: true },
        themeToggle: { key: "T", description: "Toggle dark/light mode", done: true },
        reload: { key: "R", description: "Reload page", done: true },
        escape: { key: "Escape", description: "Close modal", done: true },
        close: { key: "q", description: "Close modal", done: true },
        sort: { key: "s", description: "sort menu" },
      },
    },
  },
};

export default config;
