// api.js
import config from "../config.js";

/**
 * Endpoints for API and Web routes.
 *
 * Provides structured access to REST API paths and web routes
 * for bookmarks, imports, database management, and related operations.
 * @typedef {object} APIEndpoints
 * @property {string} scrapeUrl - Endpoint for scraping metadata.
 * @property {string} archiveUrl - Endpoint for archiving a URL.
 * @property {(db: string) => string} importHtml - Import bookmarks from HTML.
 * @property {(db: string) => string} importRepoJson - Import repo in JSON format.
 * @property {(db: string) => string} importRepoGpg - Import repo with GPG verification.
 * @property {(db: string) => string} listTags - Fetch all tags.
 * @property {(db: string) => string} createBookmark - Create a new bookmark.
 * @property {(db: string, id: string) => string} toggleFavorite - Mark a bookmark as favorite.
 * @property {(db: string, id: string) => string} recordVisit - Record a bookmark visit.
 * @property {(db: string, id: string) => string} updateBookmark - Update a bookmark.
 * @property {(db: string, id: string) => string} updateNotes - Update a bookmark's notes.
 * @property {(db: string, id: string) => string} deleteBookmark - Delete a bookmark.
 * @property {(db: string, id: string) => string} updateStatus - Get bookmark status.
 * @property {(db: string, id: string) => string} getBookmarkById - Get a bookmark by ID.
 * @property {(db: string) => string} getDbInfo - Get database info.
 * @property {(db: string) => string} createDb - Create a new database.
 * @property {(db: string) => string} deleteDb - Delete a database.
 * @property {string} listDatabases - List available databases.
 * @property {string} getAllDbInfo - Get info about all databases.
 */

const API_BASE_PATH = config.paths.api;

/** @type {APIEndpoints} */
const API = {
  // Global Endpoints
  scrapeUrl: `${API_BASE_PATH}/scrape`,
  archiveUrl: `${API_BASE_PATH}/archive`,

  // Import Endpoints
  importHtml: (db) => `${API_BASE_PATH}/${db}/import/html`,
  importRepoJson: (db) => `${API_BASE_PATH}/${db}/import/repojson`,
  importRepoGpg: (db) => `${API_BASE_PATH}/${db}/import/repogpg`,

  // Bookmark/Record Endpoints
  listTags: (db) => `${API_BASE_PATH}/${db}/bookmarks/tags`,
  createBookmark: (db) => `${API_BASE_PATH}/${db}/bookmarks/new`,
  toggleFavorite: (db, id) => `${API_BASE_PATH}/${db}/bookmarks/${id}/favorite`,
  recordVisit: (db, id) => `${API_BASE_PATH}/${db}/bookmarks/${id}/visit`,
  updateBookmark: (db, id) => `${API_BASE_PATH}/${db}/bookmarks/${id}/update`,
  updateNotes: (db, id) => `${API_BASE_PATH}/${db}/bookmarks/${id}/notes`,
  deleteBookmark: (db, id) => `${API_BASE_PATH}/${db}/bookmarks/${id}/delete`,
  updateStatus: (db, id) => `${API_BASE_PATH}/${db}/bookmarks/${id}/status`,
  getBookmarkById: (db, id) => `${API_BASE_PATH}/${db}/bookmarks/${id}`,

  // Individual Database Endpoints
  getDbInfo: (db) => `${API_BASE_PATH}/${db}/info`,
  createDb: (db) => `${API_BASE_PATH}/${db}/new`,
  deleteDb: (db) => `${API_BASE_PATH}/${db}/delete`,
  // Database Management Endpoints
  listDatabases: `${API_BASE_PATH}/repo/list`,
  getAllDbInfo: `${API_BASE_PATH}/repo/all`,
};

/**
 * @typedef {object} WebEndpoints
 * @property {(db: string) => string} viewAllBookmarks - Web route to view all bookmarks.
 * @property {(db: string) => string} createBookmarkPage - Web route to add a new bookmark.
 * @property {(db: string, id: string) => string} viewBookmark - Web route to view a single bookmark.
 * @property {(db: string, id: string) => string} viewQrCode - Web route for bookmark QR code.
 * @property {string} changeTheme - Web route for changing theme.
 */

const WEB_BASE_PATH = config.paths.web;

/** @type {WebEndpoints} */
const WEB = {
  viewAllBookmarks: (db) => `${WEB_BASE_PATH}/${db}/bookmarks/all`,
  createBookmarkPage: (db) => `${WEB_BASE_PATH}/${db}/bookmarks/new`,
  viewBookmark: (db, id) => `${WEB_BASE_PATH}/${db}/bookmarks/view/${id}`,
  viewQrCode: (db, id) => `${WEB_BASE_PATH}/${db}/bookmarks/qr/${id}`,
  changeTheme: `${WEB_BASE_PATH}/theme/change`,
};

const routes = { api: API, front: WEB };

export default routes;
