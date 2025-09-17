// Set of useless query parameters to remove
const uselessParams = new Set([
  // Tracking
  "utm_source",
  "utm_medium",
  "utm_campaign",
  "utm_term",
  "utm_content",
  "gclid",
  "fbclid",
  "msclkid",
  "dclid",
  "yclid",
  "_hsenc",
  "_hsmi",
  "mc_cid",
  "mc_eid",

  // Analytics & sessions
  "ga",
  "_ga",
  "_gac",
  "scid",
  "sessionid",
  "sid",
  "phpsessid",
  "ref",
  "refid",
  "refsrc",
  "cid",
  "affid",

  // Cache busters
  "cb",
  "_",
  "nocache",
  "rand",
  "ts",
  "t",
  "v",

  // A/B testing
  "exp",
  "experiment",
  "variant",
  "ab",
  "abtest",
  "split",
  "testgroup",

  // Affiliate / Ads
  "aff_source",
  "partnerid",
  "srcid",
  "campaignid",
  "adid",
  "adgroupid",

  // Additional common tracking params
  "source",
  "medium",
  "campaign",
  "term",
  "content",
  "_branch_match_id",
  "igshid",
  "share",
  "at_medium",
  "at_campaign",
  "wt_mc_id",
  "wt_mc_ev",
  "trk",
  "trkid",
  "li_fat_id",
  "vero_conv",
  "vero_id",
  "pk_source",
  "pk_medium",
  "pk_campaign",
  "email",
  "hash",
  "amp_js_v",
  "amp_gsa",
  "usqp",
  "sa",
  "ved",
  "usg",
  "cd",
  "cad",
  "rct",
  "ei",
  "biw",
  "bih",
]);

/**
 * Removes useless query parameters from URLSearchParams
 * @param {URLSearchParams} searchParams - URLSearchParams object to clean
 * @returns {string[]} Array of deleted parameter names
 */
function cleanParams(searchParams) {
  const deletedParams = [];

  // Convert to array to avoid mutation during iteration
  const paramsArray = Array.from(searchParams.keys());

  for (const param of paramsArray) {
    if (uselessParams.has(param.toLowerCase())) {
      deletedParams.push(param);
      searchParams.delete(param);
    }
  }

  console.log({ deletedParams });

  return deletedParams;
}

/**
 * Alternative function that also returns which parameters were deleted
 * @param {string} rawURL - The URL string to clean
 * @returns {{cleanedURL: string, deletedParams: string[]}} Object with cleaned URL and deleted params
 */
export default function cleanURL(rawURL) {
  let parsedURL;

  try {
    parsedURL = new URL(rawURL);
  } catch (error) {
    throw new Error(`Invalid URL: ${error.message}`);
  }

  const deletedParams = cleanParams(parsedURL.searchParams);

  return {
    cleanedURL: parsedURL.toString(),
    deletedParams: deletedParams,
  };
}
