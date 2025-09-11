// Package helpers is the package that should not be named like that.
package helpers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/mateconpizza/gm/pkg/bookmark"
	"github.com/mateconpizza/gm/pkg/files"
)

// ShortStr shortens a string to a maximum length.
//
//	string...
func ShortStr(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength-3] + "..."
	}

	return s
}

func TitleFirstLetter(s string) string {
	if s == "" {
		return s
	}

	// Get the first rune and its size
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s // Return original if invalid UTF-8
	}

	// Capitalize the first rune
	titleRune := unicode.ToTitle(r)

	// Combine the capitalized rune with the rest of the string
	return string(titleRune) + s[size:]
}

func FormatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	// Format: "Jan. 2, 2006, 3:04 p.m."
	return t.Format("Jan. 2, 2006, 3:04 p.m.")
}

// FormatTimestamp converts a timestamp string into format "Jan. 2, 2006, 3:04
// p.m.".
func FormatTimestamp(timestamp string) string {
	// FIX: wrong format
	const inputLayout = "20060102150405"
	const outputLayout = "Jan. 2, 2006, 3:04 p.m."
	t, err := time.Parse(inputLayout, timestamp)
	if err != nil {
		return timestamp
	}

	return t.Format(outputLayout)
}

// RelativeISOTime takes a timestamp string in ISO 8601 format (e.g., "2025-02-27T05:03:28Z")
// and returns a relative time description.
func RelativeISOTime(ts string) string {
	if ts == "" {
		return ""
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "invalid timestamp"
	}

	now := time.Now()
	// Normalize to local date only (ignore hour/minute/second)
	t = t.Local()
	now = now.Local()

	// Zero the time component for day comparison
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	diff := now.Sub(t)
	days := int(diff.Hours() / 24)

	switch {
	case days < 0:
		return "in the future"
	case days == 0:
		return "today"
	case days == 1:
		return "yesterday"
	case days < 7:
		return fmt.Sprintf("%d days ago", days)
	case days < 14:
		return "1 week ago"
	case days < 28:
		return fmt.Sprintf("%d weeks ago", days/7)
	case days < 60:
		return "1 month ago"
	case days < 365:
		return fmt.Sprintf("%d months ago", days/30)
	case days < 730:
		return "1 year ago"
	case days < 365*20: // 20 years
		return fmt.Sprintf("%d years ago", days/365)
	default:
		return "never"
	}
}

// ApplyFiltersAndSorting applies all filters and sorting to bookmarks.
func ApplyFiltersAndSorting(
	tag, query, letter, filterBy string,
	bs []*bookmark.Bookmark,
) []*bookmark.Bookmark {
	filtered := filterByTag(bs, tag)
	filtered = filterByQuery(filtered, query)
	filtered = filterByLetter(filtered, letter)

	return SortBy(filterBy, filtered)
}

// filterByQuery filters bookmarks by search query.
func filterByQuery(bookmarks []*bookmark.Bookmark, query string) []*bookmark.Bookmark {
	if query == "" {
		return bookmarks
	}

	var filtered []*bookmark.Bookmark
	queryWords := strings.Fields(strings.ToLower(query))

	for _, b := range bookmarks {
		text := strings.ToLower(b.Title + " " + b.URL + " " + b.Desc + " " + b.Tags)
		matched := true

		for _, word := range queryWords {
			if !strings.Contains(text, word) {
				matched = false
				break
			}
		}

		if matched {
			filtered = append(filtered, b)
		}
	}

	return filtered
}

// filterByTag filters bookmarks by a specific tag.
func filterByTag(bookmarks []*bookmark.Bookmark, tag string) []*bookmark.Bookmark {
	if tag == "" {
		return bookmarks
	}

	var filtered []*bookmark.Bookmark
	normalizedTag := strings.ToLower(strings.TrimPrefix(tag, "#"))

	for _, b := range bookmarks {
		for t := range strings.SplitSeq(b.Tags, ",") {
			if strings.EqualFold(strings.TrimSpace(t), normalizedTag) {
				filtered = append(filtered, b)
				break
			}
		}
	}

	return filtered
}

func filterByLetter(bs []*bookmark.Bookmark, letter string) []*bookmark.Bookmark {
	if letter == "" {
		return bs
	}

	f := make([]*bookmark.Bookmark, 0, len(bs))
	for i := range bs {
		b := bs[i]
		if b.Tags == "" {
			continue // Skip bookmarks without tags
		}

		tags := strings.Split(b.Tags, ",")
		hasMatchingTag := false

		// Check each tag to see if any starts with the letter
		for _, tag := range tags {
			tag = strings.TrimSpace(tag) // Remove whitespace
			if tag != "" && strings.EqualFold(string(tag[0]), letter) {
				hasMatchingTag = true
				break
			}
		}

		if hasMatchingTag {
			f = append(f, b)
		}
	}

	return f
}

func GroupTagsByLetter(tags []string) map[string][]string {
	sort.Strings(tags)
	grouped := make(map[string][]string)
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		first := strings.ToUpper(string(tag[0]))
		grouped[first] = append(grouped[first], tag)
	}
	return grouped
}

func SortBy(s string, bs []*bookmark.Bookmark) []*bookmark.Bookmark {
	switch s {
	case "newest":
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].CreatedAt > bs[j].CreatedAt
		})
		return bs
	case "oldest":
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].CreatedAt < bs[j].CreatedAt
		})
		return bs
	case "last_visit":
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].LastVisit > bs[j].LastVisit
		})
		return bs
	case "favorites":
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].Favorite
		})
		return bs
	case "more_visits":
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].VisitCount > bs[j].VisitCount
		})
	case "inactive":
		sort.Slice(bs, func(i, j int) bool {
			if !bs[i].IsActive && bs[j].IsActive {
				return true
			}
			if !bs[i].IsActive == !bs[j].IsActive {
				return false
			}
			return false
		})
	case "never_visited":
		neverVisited := make([]*bookmark.Bookmark, 0, len(bs))
		for i := range bs {
			if bs[i].VisitCount != 0 {
				continue
			}
			neverVisited = append(neverVisited, bs[i])
		}
		return neverVisited
	}

	return bs
}

func ExtractTags(bs []*bookmark.Bookmark) []string {
	tagsMap := map[string]bool{}
	for _, b := range bs {
		for t := range strings.SplitSeq(b.Tags, ",") {
			if t != "" {
				if !tagsMap[t] {
					tagsMap[t] = true
				}
				continue
			}
		}
	}

	uniqueTags := make([]string, 0, len(tagsMap))
	for t := range tagsMap {
		uniqueTags = append(uniqueTags, t)
	}

	return uniqueTags
}

func SortCurrentDB(dbs []string, currentDB string) []string {
	files.PrioritizeFile(dbs, currentDB)
	return dbs
}

func TagsWithPoundList(s string) []string {
	return strings.FieldsFunc(TagsWithPound(s), func(r rune) bool { return r == ' ' })
}

// TagsWithPound returns a prettified tags with #.
//
//	#tag1 #tag2 #tag3
func TagsWithPound(s string) string {
	var sb strings.Builder

	tagsSplit := strings.Split(s, ",")
	sort.Strings(tagsSplit)

	for _, t := range tagsSplit {
		if t == "" {
			continue
		}

		sb.WriteString(t + " ")
	}

	return sb.String()
}

func GetTagsFn(tag, query string, bs, paginatedBs []*bookmark.Bookmark) func() []string {
	items := paginatedBs
	if tag == "" && query == "" {
		items = bs
	}

	return func() []string {
		return ExtractTags(items)
	}
}

// GenChecksum generates a checksum for the bookmark.
func GenChecksum(rawURL, title, desc, tags string) string {
	data := fmt.Sprintf("u:%s|t:%s|d:%s|tags:%s", rawURL, title, desc, tags)
	return generateHash(data, 8)
}

func generateHash(s string, c int) string {
	hash := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(hash[:])[:c]
}

// PrintTable prints data in a markdown-style table format.
func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))

	// Initialize with header lengths
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// Check row data for maximum width
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Add padding
	for i := range colWidths {
		colWidths[i] += 2 // Add 1 space on each side
	}

	// Print header
	fmt.Print("|")
	for i, header := range headers {
		fmt.Printf("%-*s|", colWidths[i], " "+header+" ")
	}
	fmt.Println()

	// Print separator
	fmt.Print("|")
	for _, width := range colWidths {
		fmt.Print(strings.Repeat("-", width) + "|")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("|")
		for i, cell := range row {
			if i < len(colWidths) {
				fmt.Printf("%-*s|", colWidths[i], " "+cell+" ")
			}
		}
		// Fill empty cells if row is shorter than headers
		for i := len(row); i < len(colWidths); i++ {
			fmt.Printf("%-*s|", colWidths[i], " ")
		}
		fmt.Println()
	}
}

func HashString(s string, n int) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:n]
}

func HashDomain(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", HashString(parsedURL.Host, 12)), nil
}

func IsWithinLastWeek(dateString string) bool {
	t, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		fmt.Printf("Error parsing date '%s': %v\n", dateString, err)
		return false
	}
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	return t.After(sevenDaysAgo) || t.Equal(sevenDaysAgo)
}
