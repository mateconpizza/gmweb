// cookie.js

const EXPIRY_ONE_YEAR = 60 * 60 * 24 * 365;

/**
 * @typedef {object} CookieJar
 * @property {string} vim - VimMode
 * @property {string} compact - Compact UI
 * @property {string} theme - Theme name
 * @property {string} themeMode - Theme mode
 * @property {string} itemsPage - Iterm per page
 * @property {string} defaultRepoName - Default repository
 */

/**
 * @typedef {object} CookieExpiry
 * @property {number} oneYear -
 */

/**
 * @typedef {object} CookieApi
 * @property {CookieJar} jar - Key names for the cookies.
 * @property {(name: string) => string | undefined} get - Gets a cookie value by name.
 * @property {(name: string, value: string, maxAge?: number) => void} set - Saves a cookie.
 * @property {(name: string) => void} delete - Deletes a cookie by name.
 * @property {() => object} getUserPreferences - Gets all user preferences from cookies.
 * @property {CookieExpiry} expiry - Expiry constants for cookies.
 */

/** @type {CookieApi} */
const Cookie = {
  jar: {
    vim: "vim_mode",
    compact: "compact_mode",
    theme: "user_theme",
    themeMode: "theme_mode",
    itemsPage: "items_per_page",
    defaultRepoName: "default_repo",
  },
  expiry: {
    oneYear: EXPIRY_ONE_YEAR,
  },

  get: (name) => {
    return document.cookie
      .split("; ")
      .find((row) => row.startsWith(`${name}=`))
      ?.split("=")[1];
  },
  set: (name, value, maxAge = Cookie.expiry.oneYear) => {
    document.cookie = `${name}=${value}; path=/; max-age=${maxAge}; SameSite=Lax`;
  },
  delete: (name) => {
    document.cookie = `${name}=; path=/; max-age=0`;
  },
  getUserPreferences: () => {
    return {
      theme: Cookie.get(Cookie.jar.theme),
      themeMode: Cookie.get(Cookie.jar.themeMode),
      vimMode: Cookie.get(Cookie.jar.vim) === "true",
      compactMode: Cookie.get(Cookie.jar.compact) === "true",
      itemsPerPage: parseInt(Cookie.get(Cookie.jar.itemsPage)) || 32,
    };
  },
};

export default Cookie;
