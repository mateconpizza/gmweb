/**
 * @module types
 *
 * This module defines the data structures used for bookmarks in the application.
 * It includes the `Bookmark` and `BookmarkJSON` types, which represent the
 * properties and structure of bookmark objects, both in the database and
 * when serialized to JSON format.
 */

// FIX: should i keep this?

/**
 * @typedef {object} Bookmark
 * @property {number} ID - The unique identifier for the bookmark.
 * @property {string} URL - The URL of the bookmark.
 * @property {string} Tags - Tags for the bookmark, stored as a comma-separated string.
 * @property {string} Title - The title of the bookmark.
 * @property {string} Desc - The description of the bookmark.
 * @property {string} CreatedAt - Timestamp when the bookmark was created.
 * @property {string} LastVisit - Timestamp of the last visit.
 * @property {string} UpdatedAt - Timestamp of the last update.
 * @property {number} VisitCount - The number of times the bookmark has been visited.
 * @property {boolean} Favorite - Indicates if the bookmark is a favorite.
 * @property {string} FaviconURL - URL for the bookmark's favicon.
 * @property {string} FaviconLocal - Local path to the cached favicon.
 * @property {string} Checksum - Checksum of the bookmark's data (URL, Title, Description, Tags).
 * @property {string} ArchiveURL - The Internet Archive URL for the bookmark.
 * @property {string} ArchiveTimestamp - The timestamp from the Internet Archive.
 * @property {string} LastStatusChecked - Timestamp of the last status check.
 * @property {number} HTTPStatusCode - The HTTP status code of the URL (e.g., 200, 404).
 * @property {string} HTTPStatusText - The HTTP status text (e.g., 'OK', 'Not Found').
 * @property {boolean} IsActive - Indicates if the URL is active (status code 200-299).
 */

/**
 * @typedef {object} BookmarkJSON
 * @property {number} ID - The unique identifier for the bookmark.
 * @property {string} URL - The URL of the bookmark.
 * @property {string[]} Tags - An array of tags for the bookmark.
 * @property {string} Title - The title of the bookmark.
 * @property {string} Desc - The description of the bookmark.
 * @property {string} CreatedAt - The timestamp when the bookmark was created.
 * @property {string} LastVisit - The timestamp of the last visit.
 * @property {string} UpdatedAt - The timestamp of the last update.
 * @property {number} VisitCount - The number of times the bookmark has been visited.
 * @property {boolean} Favorite - A boolean indicating if the bookmark is a favorite.
 * @property {string} FaviconURL - The URL for the bookmark's favicon.
 * @property {string} FaviconLocal - The local path to the cached favicon file.
 * @property {string} Checksum - The checksum or hash of the bookmark's data.
 * @property {string} ArchiveURL - The Internet Archive URL.
 * @property {string} ArchiveTimestamp - The Internet Archive timestamp.
 * @property {string} LastStatusChecked - The timestamp of the last status check.
 * @property {number} HTTPStatusCode - The HTTP status code (e.g., 200, 404).
 * @property {string} HTTPStatusText - The HTTP status text (e.g., "OK", "Not Found").
 * @property {boolean} IsActive - A boolean indicating if the URL is active (status code 200-299).
 */
