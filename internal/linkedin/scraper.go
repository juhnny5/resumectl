// Copyright (c) 2026 Julien Briault
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package linkedin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"resumectl/internal/models"
)

// DebugMode enables debug logs
var DebugMode bool

// LinkedInProfile represents data extracted from a LinkedIn profile
type LinkedInProfile struct {
	FirstName      string
	LastName       string
	Headline       string
	Location       string
	Summary        string
	Experience     []LinkedInExperience
	Education      []LinkedInEducation
	Skills         []string
	Languages      []LinkedInLanguage
	Certifications []LinkedInCertification
}

// LinkedInExperience represents a LinkedIn work experience
type LinkedInExperience struct {
	Title       string
	Company     string
	Location    string
	StartDate   string
	EndDate     string
	Description string
}

// LinkedInEducation represents a LinkedIn education entry
type LinkedInEducation struct {
	School      string
	Degree      string
	Field       string
	StartDate   string
	EndDate     string
	Description string
}

// LinkedInLanguage represents a language on LinkedIn
type LinkedInLanguage struct {
	Name        string
	Proficiency string
}

// LinkedInCertification represents a LinkedIn certification
type LinkedInCertification struct {
	Name         string
	Organization string
	IssueDate    string
}

// ExtractUsernameFromURL extracts the username from a LinkedIn URL
func ExtractUsernameFromURL(linkedinURL string) (string, error) {
	// Clean the URL
	linkedinURL = strings.TrimSpace(linkedinURL)

	// If not a full URL, assume it's just the username
	if !strings.Contains(linkedinURL, "linkedin.com") {
		return linkedinURL, nil
	}

	// Parser l'URL
	parsedURL, err := url.Parse(linkedinURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Add scheme if missing
	if parsedURL.Scheme == "" {
		linkedinURL = "https://" + linkedinURL
		parsedURL, err = url.Parse(linkedinURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
	}

	// Extraire le chemin
	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")

	// Le format attendu est /in/username
	for i, part := range parts {
		if part == "in" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	// Si on a juste un chemin simple
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], nil
	}

	return "", fmt.Errorf("could not extract username from LinkedIn URL: %s", linkedinURL)
}

// FetchProfile retrieves data from a public LinkedIn profile (without authentication)
func FetchProfile(username string) (*LinkedInProfile, error) {
	return FetchProfileWithAuth(username, "")
}

// FetchProfileWithAuth retrieves LinkedIn profile data with optional authentication
// If sessionCookie is provided (li_at cookie), all data will be accessible via Voyager API
func FetchProfileWithAuth(username string, sessionCookie string) (*LinkedInProfile, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// If authenticated, use Voyager API to retrieve complete data
	if sessionCookie != "" {
		profile, err := fetchViaVoyagerAPI(client, username, sessionCookie)
		if err == nil && profile.FirstName != "" {
			return profile, nil
		}
		// In case of API error, fallback to HTML parsing
	}

	// Fallback: retrieve via public HTML page
	profileURL := fmt.Sprintf("https://www.linkedin.com/in/%s/", username)

	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	if sessionCookie != "" {
		req.Header.Set("Cookie", "li_at="+sessionCookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LinkedIn returned status %d - the profile may be private or not exist", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return parseLinkedInHTML(string(body), username, sessionCookie != "")
}

// fetchViaVoyagerAPI retrieves complete data via LinkedIn's internal API
func fetchViaVoyagerAPI(client *http.Client, username string, sessionCookie string) (*LinkedInProfile, error) {
	profile := &LinkedInProfile{}

	// First, fetch the page to get JSESSIONID (CSRF token)
	pageURL := fmt.Sprintf("https://www.linkedin.com/in/%s/", username)
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Cookie", "li_at="+sessionCookie)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract JSESSIONID from response cookies
	var jsessionid string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			jsessionid = strings.Trim(cookie.Value, "\"")
			break
		}
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Si pas de JSESSIONID dans les cookies, essayer dans la page
	if jsessionid == "" {
		jsessionPattern := regexp.MustCompile(`"JSESSIONID":"([^"]+)"`)
		matches := jsessionPattern.FindStringSubmatch(string(body))
		if len(matches) > 1 {
			jsessionid = strings.Trim(matches[1], "\"")
		}
	}

	// Essayer aussi le pattern ajax:
	if jsessionid == "" {
		jsessionPattern := regexp.MustCompile(`JSESSIONID=([^;]+)`)
		matches := jsessionPattern.FindStringSubmatch(string(body))
		if len(matches) > 1 {
			jsessionid = strings.Trim(matches[1], "\"")
		}
	}

	if jsessionid == "" {
		return nil, fmt.Errorf("could not find CSRF token")
	}

	if DebugMode {
		fmt.Printf("[DEBUG] JSESSIONID found: %s\n", jsessionid)
	}

	// Call Voyager API to retrieve complete profile
	apiURL := fmt.Sprintf("https://www.linkedin.com/voyager/api/identity/dash/profiles?q=memberIdentity&memberIdentity=%s&decorationId=com.linkedin.voyager.dash.deco.identity.profile.FullProfileWithEntities-93", username)

	req, err = http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/vnd.linkedin.normalized+json+2.1")
	req.Header.Set("Cookie", fmt.Sprintf("li_at=%s; JSESSIONID=\"%s\"", sessionCookie, jsessionid))
	req.Header.Set("csrf-token", jsessionid)
	req.Header.Set("x-li-lang", "fr_FR")
	req.Header.Set("x-restli-protocol-version", "2.0.0")
	req.Header.Set("x-li-track", `{"clientVersion":"1.13.8677","mpVersion":"1.13.8677","osName":"web","timezoneOffset":1,"timezone":"Europe/Paris","deviceFormFactor":"DESKTOP","mpName":"voyager-web","displayDensity":1,"displayWidth":1920,"displayHeight":1080}`)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if DebugMode {
		fmt.Printf("[DEBUG] API response status: %d\n", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		apiBody, _ := io.ReadAll(resp.Body)
		if DebugMode {
			fmt.Printf("[DEBUG] API error response: %s\n", string(apiBody)[:min(500, len(apiBody))])
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	apiBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if DebugMode {
		fmt.Printf("[DEBUG] API response size: %d bytes\n", len(apiBody))
	}

	// Parse the JSON response from the API
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(apiBody, &apiResponse); err != nil {
		return nil, err
	}

	// Extract data from the "included" response
	if included, ok := apiResponse["included"].([]interface{}); ok {
		if DebugMode {
			fmt.Printf("[DEBUG] Found %d items in 'included'\n", len(included))
		}
		for _, item := range included {
			if itemMap, ok := item.(map[string]interface{}); ok {
				extractFromVoyagerData(itemMap, profile)
			}
		}
	}

	// Also extract from "elements" if it exists
	if data, ok := apiResponse["data"].(map[string]interface{}); ok {
		if elements, ok := data["elements"].([]interface{}); ok && len(elements) > 0 {
			if elem, ok := elements[0].(map[string]interface{}); ok {
				extractFromVoyagerData(elem, profile)
			}
		}
	}

	return profile, nil
}

// parseLinkedInHTML parses the LinkedIn page HTML to extract data
func parseLinkedInHTML(html string, username string, authenticated bool) (*LinkedInProfile, error) {
	profile := &LinkedInProfile{}

	// If authenticated, try to extract from Voyager data (LinkedIn internal API)
	if authenticated {
		extractFromAuthenticatedPage(html, profile)
	}

	// Try to extract JSON-LD data embedded in the page
	jsonLDPattern := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>([\s\S]*?)</script>`)
	matches := jsonLDPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			// Essayer de parser comme un objet avec @graph
			var graphData struct {
				Graph []map[string]interface{} `json:"@graph"`
			}
			if err := json.Unmarshal([]byte(match[1]), &graphData); err == nil && len(graphData.Graph) > 0 {
				for _, item := range graphData.Graph {
					extractFromJSONLD(item, profile)
				}
			} else {
				// Essayer comme objet simple
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(match[1]), &jsonData); err == nil {
					extractFromJSONLD(jsonData, profile)
				}
			}
		}
	}

	// Extract from meta tags
	extractFromMetaTags(html, profile)

	// Extract from visible page content
	extractFromHTMLContent(html, profile)

	// If we couldn't extract the name, use the username
	if profile.FirstName == "" && profile.LastName == "" {
		parts := strings.Split(username, "-")
		if len(parts) >= 2 {
			profile.FirstName = capitalizeFirst(parts[0])
			profile.LastName = capitalizeFirst(parts[1])
		} else {
			profile.FirstName = capitalizeFirst(username)
		}
	}

	return profile, nil
}

// extractFromJSONLD extracts data from JSON-LD
func extractFromJSONLD(data map[string]interface{}, profile *LinkedInProfile) {
	// Check schema type
	schemaType, _ := data["@type"].(string)

	// Ne traiter que les Person
	if schemaType != "Person" && schemaType != "" {
		return
	}

	// Extraire le nom
	if name, ok := data["name"].(string); ok && profile.FirstName == "" {
		parts := strings.SplitN(name, " ", 2)
		if len(parts) >= 1 {
			profile.FirstName = parts[0]
		}
		if len(parts) >= 2 {
			profile.LastName = parts[1]
		}
	}

	// Extraire la description/summary
	if description, ok := data["description"].(string); ok && profile.Summary == "" {
		// Nettoyer les balises HTML
		summary := strings.ReplaceAll(description, "<br>", "\n")
		summary = strings.ReplaceAll(summary, "\\u003Cbr\\u003E", "\n")
		profile.Summary = summary
	}

	// Extraire l'adresse/localisation
	if address, ok := data["address"].(map[string]interface{}); ok {
		if locality, ok := address["addressLocality"].(string); ok && locality != "" {
			profile.Location = locality
		}
	}

	// Extract experiences (worksFor)
	if worksFor, ok := data["worksFor"].([]interface{}); ok {
		for _, work := range worksFor {
			if workMap, ok := work.(map[string]interface{}); ok {
				exp := LinkedInExperience{}

				// Company name
				if name, ok := workMap["name"].(string); ok {
					// Ignore masked names (asterisks)
					if !strings.Contains(name, "***") {
						exp.Company = name
					}
				}

				// Location
				if location, ok := workMap["location"].(string); ok {
					exp.Location = location
				}

				// Role details (member)
				if member, ok := workMap["member"].(map[string]interface{}); ok {
					if desc, ok := member["description"].(string); ok {
						// Ignore masked descriptions
						if !strings.Contains(desc, "***") {
							exp.Description = desc
						}
					}
					if startDate, ok := member["startDate"]; ok {
						exp.StartDate = fmt.Sprintf("%v", startDate)
					}
					if endDate, ok := member["endDate"]; ok {
						exp.EndDate = fmt.Sprintf("%v", endDate)
					}
				}

				// Only add if we have a valid company name
				if exp.Company != "" {
					profile.Experience = append(profile.Experience, exp)
				}
			}
		}
	}

	// Extract education (alumniOf)
	if alumniOf, ok := data["alumniOf"].([]interface{}); ok {
		for _, school := range alumniOf {
			if schoolMap, ok := school.(map[string]interface{}); ok {
				edu := LinkedInEducation{}

				// School name
				if name, ok := schoolMap["name"].(string); ok {
					// Ignore masked names
					if !strings.Contains(name, "***") {
						edu.School = name
					}
				}

				// Details (member)
				if member, ok := schoolMap["member"].(map[string]interface{}); ok {
					if desc, ok := member["description"].(string); ok {
						// Ignore masked descriptions
						if !strings.Contains(desc, "***") {
							edu.Degree = desc
						}
					}
					if startDate, ok := member["startDate"]; ok {
						edu.StartDate = fmt.Sprintf("%v", startDate)
					}
					if endDate, ok := member["endDate"]; ok {
						edu.EndDate = fmt.Sprintf("%v", endDate)
					}
				}

				// Only add if we have a valid school name
				if edu.School != "" {
					profile.Education = append(profile.Education, edu)
				}
			}
		}
	}

	// Extraire les langues (knowsLanguage)
	if languages, ok := data["knowsLanguage"].([]interface{}); ok {
		for _, lang := range languages {
			if langMap, ok := lang.(map[string]interface{}); ok {
				if name, ok := langMap["name"].(string); ok && name != "" {
					profile.Languages = append(profile.Languages, LinkedInLanguage{
						Name: name,
					})
				}
			}
		}
	}

	// Extract skills (knowsAbout) - rarely available
	if skills, ok := data["knowsAbout"].([]interface{}); ok {
		for _, skill := range skills {
			if skillStr, ok := skill.(string); ok && skillStr != "" {
				profile.Skills = append(profile.Skills, skillStr)
			}
		}
	}

	// Extract title from jobTitle (array or string)
	if jobTitles, ok := data["jobTitle"].([]interface{}); ok && len(jobTitles) > 0 {
		// Take the first non-masked title
		for _, jt := range jobTitles {
			if title, ok := jt.(string); ok && !strings.Contains(title, "***") {
				profile.Headline = title
				break
			}
		}
	} else if jobTitle, ok := data["jobTitle"].(string); ok && !strings.Contains(jobTitle, "***") {
		profile.Headline = jobTitle
	}
}

// extractFromMetaTags extracts data from meta tags
func extractFromMetaTags(html string, profile *LinkedInProfile) {
	// Title meta tag
	titlePattern := regexp.MustCompile(`<meta[^>]*property="og:title"[^>]*content="([^"]*)"`)
	if matches := titlePattern.FindStringSubmatch(html); len(matches) > 1 {
		title := matches[1]
		// The format is usually "FirstName LastName - Title | LinkedIn"
		if idx := strings.Index(title, " - "); idx > 0 {
			namePart := title[:idx]
			parts := strings.SplitN(namePart, " ", 2)
			if profile.FirstName == "" && len(parts) >= 1 {
				profile.FirstName = parts[0]
			}
			if profile.LastName == "" && len(parts) >= 2 {
				profile.LastName = parts[1]
			}

			// Extraire le titre
			restPart := title[idx+3:]
			if pipeIdx := strings.Index(restPart, " | "); pipeIdx > 0 {
				if profile.Headline == "" {
					profile.Headline = strings.TrimSpace(restPart[:pipeIdx])
				}
			}
		}
	}

	// Description meta tag
	descPattern := regexp.MustCompile(`<meta[^>]*property="og:description"[^>]*content="([^"]*)"`)
	if matches := descPattern.FindStringSubmatch(html); len(matches) > 1 {
		if profile.Summary == "" {
			profile.Summary = cleanHTMLEntities(matches[1])
		}
	}

	// Location depuis le contenu
	locationPattern := regexp.MustCompile(`<meta[^>]*name="geo.placename"[^>]*content="([^"]*)"`)
	if matches := locationPattern.FindStringSubmatch(html); len(matches) > 1 {
		if profile.Location == "" {
			profile.Location = matches[1]
		}
	}
}

// extractFromHTMLContent extracts data from visible HTML content
func extractFromHTMLContent(html string, profile *LinkedInProfile) {
	// Try to extract location from different patterns
	locationPatterns := []string{
		`class="top-card-subline-item[^"]*"[^>]*>([^<]+)</span>`,
		`class="profile-info-subheader[^"]*"[^>]*>([^<]+)</span>`,
	}

	for _, pattern := range locationPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(html); len(matches) > 1 {
			location := strings.TrimSpace(matches[1])
			if profile.Location == "" && !strings.Contains(location, "@") {
				profile.Location = location
				break
			}
		}
	}
}

// extractFromPageData extracts data from scripts embedded in the page
func extractFromPageData(html string, profile *LinkedInProfile) {
	// Search for data in script tags containing JSON data
	// Pattern for LinkedIn profile data
	dataPatterns := []string{
		`"firstName":"([^"]+)"`,
		`"lastName":"([^"]+)"`,
		`"headline":"([^"]+)"`,
		`"locationName":"([^"]+)"`,
		`"geoLocationName":"([^"]+)"`,
		`"summary":"([^"]+)"`,
	}

	// Extract first name
	if profile.FirstName == "" {
		re := regexp.MustCompile(`"firstName"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(html); len(matches) > 1 {
			profile.FirstName = cleanHTMLEntities(matches[1])
		}
	}

	// Extraire le nom
	if profile.LastName == "" {
		re := regexp.MustCompile(`"lastName"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(html); len(matches) > 1 {
			profile.LastName = cleanHTMLEntities(matches[1])
		}
	}

	// Extraire le titre/headline
	if profile.Headline == "" {
		re := regexp.MustCompile(`"headline"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(html); len(matches) > 1 {
			profile.Headline = cleanHTMLEntities(matches[1])
		}
	}

	// Extraire la localisation
	if profile.Location == "" {
		patterns := []string{
			`"locationName"\s*:\s*"([^"]+)"`,
			`"geoLocationName"\s*:\s*"([^"]+)"`,
			`"location"\s*:\s*\{[^}]*"name"\s*:\s*"([^"]+)"`,
		}
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(html); len(matches) > 1 {
				profile.Location = cleanHTMLEntities(matches[1])
				break
			}
		}
	}

	// Extract summary/about
	if profile.Summary == "" {
		patterns := []string{
			`"summary"\s*:\s*"([^"]{20,})"`,
			`"about"\s*:\s*"([^"]{20,})"`,
		}
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(html); len(matches) > 1 {
				summary := cleanHTMLEntities(matches[1])
				// Decode unicode sequences
				summary = decodeUnicodeEscapes(summary)
				profile.Summary = summary
				break
			}
		}
	}

	// Extract experiences from structured data
	if len(profile.Experience) == 0 {
		// Pattern pour les positions
		expPattern := regexp.MustCompile(`"companyName"\s*:\s*"([^"]+)"[^}]*"title"\s*:\s*"([^"]+)"`)
		expMatches := expPattern.FindAllStringSubmatch(html, -1)
		seen := make(map[string]bool)
		for _, match := range expMatches {
			if len(match) > 2 {
				key := match[1] + "|" + match[2]
				if !seen[key] {
					seen[key] = true
					profile.Experience = append(profile.Experience, LinkedInExperience{
						Company: cleanHTMLEntities(match[1]),
						Title:   cleanHTMLEntities(match[2]),
					})
				}
			}
		}

		// Pattern alternatif
		if len(profile.Experience) == 0 {
			expPattern2 := regexp.MustCompile(`"title"\s*:\s*"([^"]+)"[^}]*"companyName"\s*:\s*"([^"]+)"`)
			expMatches2 := expPattern2.FindAllStringSubmatch(html, -1)
			for _, match := range expMatches2 {
				if len(match) > 2 {
					key := match[2] + "|" + match[1]
					if !seen[key] {
						seen[key] = true
						profile.Experience = append(profile.Experience, LinkedInExperience{
							Company: cleanHTMLEntities(match[2]),
							Title:   cleanHTMLEntities(match[1]),
						})
					}
				}
			}
		}
	}

	// Extract education
	if len(profile.Education) == 0 {
		eduPattern := regexp.MustCompile(`"schoolName"\s*:\s*"([^"]+)"`)
		eduMatches := eduPattern.FindAllStringSubmatch(html, -1)
		seen := make(map[string]bool)
		for _, match := range eduMatches {
			if len(match) > 1 && !seen[match[1]] {
				seen[match[1]] = true
				profile.Education = append(profile.Education, LinkedInEducation{
					School: cleanHTMLEntities(match[1]),
				})
			}
		}
	}

	// Extract skills
	if len(profile.Skills) == 0 {
		skillPattern := regexp.MustCompile(`"skillName"\s*:\s*"([^"]+)"`)
		skillMatches := skillPattern.FindAllStringSubmatch(html, -1)
		seen := make(map[string]bool)
		for _, match := range skillMatches {
			if len(match) > 1 && !seen[match[1]] {
				seen[match[1]] = true
				profile.Skills = append(profile.Skills, cleanHTMLEntities(match[1]))
			}
		}
	}

	// Ignore unused patterns to avoid warnings
	_ = dataPatterns
}

// decodeUnicodeEscapes decodes \uXXXX unicode sequences
func decodeUnicodeEscapes(s string) string {
	re := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		var r rune
		fmt.Sscanf(match, `\u%04x`, &r)
		return string(r)
	})
}

// cleanHTMLEntities cleans HTML entities
func cleanHTMLEntities(s string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
		"&nbsp;": " ",
	}

	result := s
	for entity, char := range replacements {
		result = strings.ReplaceAll(result, entity, char)
	}
	return result
}

// capitalizeFirst capitalizes the first letter
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
}

// ToCV converts a LinkedIn profile to CV structure
// Data is integrated directly without adding dummy templates
func (p *LinkedInProfile) ToCV(linkedinURL string) *models.CV {
	cv := &models.CV{
		Personal: models.Personal{
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Title:     p.Headline,
			Email:     "your.email@example.com", // To be completed by user
			Phone:     "+33 6 00 00 00 00",      // To be completed by user
			Location:  p.Location,
			LinkedIn:  formatLinkedInURL(linkedinURL),
		},
		Summary: p.Summary,
	}

	// Convert experiences directly
	for _, exp := range p.Experience {
		cvExp := models.Experience{
			Company:     exp.Company,
			Position:    exp.Title,
			Location:    exp.Location,
			StartDate:   exp.StartDate,
			EndDate:     exp.EndDate,
			Description: exp.Description,
		}
		// Ne pas ajouter de highlights vides
		if exp.Description != "" {
			cvExp.Highlights = []string{}
		}
		cv.Experience = append(cv.Experience, cvExp)
	}

	// Convert education directly
	for _, edu := range p.Education {
		cv.Education = append(cv.Education, models.Education{
			Institution: edu.School,
			Degree:      edu.Degree,
			Field:       edu.Field,
			StartDate:   edu.StartDate,
			EndDate:     edu.EndDate,
			Description: edu.Description,
		})
	}

	// Convert skills directly
	if len(p.Skills) > 0 {
		cv.Skills = []models.SkillCategory{
			{
				Category: "Skills",
				Items:    p.Skills,
			},
		}
	}

	// Convert languages directly
	for _, lang := range p.Languages {
		cv.Languages = append(cv.Languages, models.Language{
			Name:  lang.Name,
			Level: mapProficiencyLevel(lang.Proficiency),
		})
	}

	// Convert certifications directly
	for _, cert := range p.Certifications {
		cv.Certifications = append(cv.Certifications, models.Certification{
			Name:   cert.Name,
			Issuer: cert.Organization,
			Date:   cert.IssueDate,
		})
	}

	// No projects by default - user will add them
	cv.Projects = []models.Project{}

	return cv
}

// formatLinkedInURL formate l'URL LinkedIn pour le CV
func formatLinkedInURL(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")
	url = strings.TrimSuffix(url, "/")
	return url
}

// formatSummary formats the summary
func formatSummary(summary string) string {
	if summary == "" {
		return "Write a brief professional summary highlighting your key skills, experience, and career objectives."
	}
	return summary
}

// formatDate formats a date
func formatDate(date string) string {
	if date == "" {
		return "2020"
	}
	return date
}

// formatEndDate formats an end date
func formatEndDate(date string) string {
	if date == "" || strings.ToLower(date) == "present" || strings.ToLower(date) == "current" {
		return "present"
	}
	return date
}

// mapProficiencyLevel converts LinkedIn proficiency level to CV format
func mapProficiencyLevel(proficiency string) string {
	proficiency = strings.ToLower(proficiency)
	switch {
	case strings.Contains(proficiency, "native") || strings.Contains(proficiency, "bilingual"):
		return "Native"
	case strings.Contains(proficiency, "full professional") || strings.Contains(proficiency, "fluent"):
		return "Fluent (C1-C2)"
	case strings.Contains(proficiency, "professional working"):
		return "Professional (B2)"
	case strings.Contains(proficiency, "limited working"):
		return "Intermediate (B1)"
	case strings.Contains(proficiency, "elementary"):
		return "Elementary (A2)"
	default:
		return proficiency
	}
}

// extractFromAuthenticatedPage extracts data from an authenticated LinkedIn page
// LinkedIn loads data via "code" scripts containing JSON with complete info
func extractFromAuthenticatedPage(html string, profile *LinkedInProfile) {
	// Extract data from "code" blocks (Voyager/React format)
	// Pattern for JSON data embedded in scripts
	codePattern := regexp.MustCompile(`<code[^>]*id="bpr-guid-\d+"[^>]*><!--(.+?)--></code>`)
	matches := codePattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			// Decode HTML entities
			jsonData := decodeHTMLEntities(match[1])

			// Try to parse JSON
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
				continue
			}

			// Extract included data
			if included, ok := data["included"].([]interface{}); ok {
				for _, item := range included {
					if itemMap, ok := item.(map[string]interface{}); ok {
						extractFromVoyagerData(itemMap, profile)
					}
				}
			}
		}
	}

	// Alternative pattern for JSON data
	jsonPattern := regexp.MustCompile(`"included":\s*\[([\s\S]*?)\],"meta"`)
	jsonMatches := jsonPattern.FindAllStringSubmatch(html, -1)

	for _, match := range jsonMatches {
		if len(match) > 1 {
			// Parser le tableau included
			var included []map[string]interface{}
			jsonStr := "[" + match[1] + "]"
			if err := json.Unmarshal([]byte(jsonStr), &included); err != nil {
				continue
			}

			for _, item := range included {
				extractFromVoyagerData(item, profile)
			}
		}
	}
}

// extractFromVoyagerData extracts data from LinkedIn's Voyager format
func extractFromVoyagerData(data map[string]interface{}, profile *LinkedInProfile) {
	// Determine the data type
	entityUrn, _ := data["entityUrn"].(string)
	typeField, _ := data["$type"].(string)

	// Debug: display encountered types
	if DebugMode && typeField != "" && (strings.Contains(typeField, "position") || strings.Contains(typeField, "Position") ||
		strings.Contains(typeField, "education") || strings.Contains(typeField, "Education") ||
		strings.Contains(typeField, "profile") || strings.Contains(typeField, "Profile") ||
		strings.Contains(typeField, "skill") || strings.Contains(typeField, "Skill") ||
		strings.Contains(typeField, "language") || strings.Contains(typeField, "Language")) {
		fmt.Printf("[DEBUG] Type: %s, EntityUrn: %s\n", typeField, entityUrn)
	}

	// Extract data by type - use broader patterns
	typeLower := strings.ToLower(typeField)
	urnLower := strings.ToLower(entityUrn)

	switch {
	case strings.Contains(typeLower, "profile") && !strings.Contains(typeLower, "position") && !strings.Contains(typeLower, "education"):
		// Main profile data
		if firstName, ok := data["firstName"].(string); ok && firstName != "" {
			profile.FirstName = firstName
		}
		if lastName, ok := data["lastName"].(string); ok && lastName != "" {
			profile.LastName = lastName
		}
		if headline, ok := data["headline"].(string); ok && headline != "" {
			profile.Headline = headline
		}
		if summary, ok := data["summary"].(string); ok && summary != "" {
			profile.Summary = summary
		}
		if locationName, ok := data["locationName"].(string); ok && locationName != "" {
			profile.Location = locationName
		}
		if geoLocation, ok := data["geoLocationName"].(string); ok && geoLocation != "" && profile.Location == "" {
			profile.Location = geoLocation
		}
		// Aussi chercher dans geoLocation object
		if geoLoc, ok := data["geoLocation"].(map[string]interface{}); ok {
			if geoName, ok := geoLoc["geoLocationName"].(string); ok && geoName != "" && profile.Location == "" {
				profile.Location = geoName
			}
		}

	case strings.Contains(typeLower, "position") || strings.Contains(urnLower, "profileposition") || strings.Contains(urnLower, "fs_position"):
		// Work experience data
		exp := LinkedInExperience{}

		if title, ok := data["title"].(string); ok {
			exp.Title = title
		}
		if companyName, ok := data["companyName"].(string); ok {
			exp.Company = companyName
		}
		// Also search in company object
		if company, ok := data["company"].(map[string]interface{}); ok {
			if name, ok := company["name"].(string); ok && exp.Company == "" {
				exp.Company = name
			}
		}
		// Search in companyName as object with text
		if companyNameObj, ok := data["companyName"].(map[string]interface{}); ok {
			if text, ok := companyNameObj["text"].(string); ok {
				exp.Company = text
			}
		}
		if locationName, ok := data["locationName"].(string); ok {
			exp.Location = locationName
		}
		if description, ok := data["description"].(string); ok {
			exp.Description = description
		}

		// Extraire les dates depuis dateRange ou timePeriod
		if dateRange, ok := data["dateRange"].(map[string]interface{}); ok {
			if startDate, ok := dateRange["start"].(map[string]interface{}); ok {
				exp.StartDate = formatTimePeriodDate(startDate)
			}
			if endDate, ok := dateRange["end"].(map[string]interface{}); ok {
				exp.EndDate = formatTimePeriodDate(endDate)
			}
		}
		if timePeriod, ok := data["timePeriod"].(map[string]interface{}); ok {
			if startDate, ok := timePeriod["startDate"].(map[string]interface{}); ok {
				exp.StartDate = formatTimePeriodDate(startDate)
			}
			if endDate, ok := timePeriod["endDate"].(map[string]interface{}); ok {
				exp.EndDate = formatTimePeriodDate(endDate)
			}
		}

		if exp.Company != "" || exp.Title != "" {
			if DebugMode {
				fmt.Printf("[DEBUG] Adding experience: %s @ %s\n", exp.Title, exp.Company)
			}
			profile.Experience = append(profile.Experience, exp)
		}

	case strings.Contains(typeLower, "education") || strings.Contains(urnLower, "profileeducation") || strings.Contains(urnLower, "fs_education"):
		// Education data
		edu := LinkedInEducation{}

		if schoolName, ok := data["schoolName"].(string); ok {
			edu.School = schoolName
		}
		// Search in school object
		if school, ok := data["school"].(map[string]interface{}); ok {
			if name, ok := school["name"].(string); ok && edu.School == "" {
				edu.School = name
			}
		}
		// Search in schoolName as object with text
		if schoolNameObj, ok := data["schoolName"].(map[string]interface{}); ok {
			if text, ok := schoolNameObj["text"].(string); ok {
				edu.School = text
			}
		}
		if degreeName, ok := data["degreeName"].(string); ok {
			edu.Degree = degreeName
		}
		if fieldOfStudy, ok := data["fieldOfStudy"].(string); ok {
			edu.Field = fieldOfStudy
		}
		if description, ok := data["description"].(string); ok {
			edu.Description = description
		}

		// Extract dates
		if dateRange, ok := data["dateRange"].(map[string]interface{}); ok {
			if startDate, ok := dateRange["start"].(map[string]interface{}); ok {
				edu.StartDate = formatTimePeriodDate(startDate)
			}
			if endDate, ok := dateRange["end"].(map[string]interface{}); ok {
				edu.EndDate = formatTimePeriodDate(endDate)
			}
		}
		if timePeriod, ok := data["timePeriod"].(map[string]interface{}); ok {
			if startDate, ok := timePeriod["startDate"].(map[string]interface{}); ok {
				edu.StartDate = formatTimePeriodDate(startDate)
			}
			if endDate, ok := timePeriod["endDate"].(map[string]interface{}); ok {
				edu.EndDate = formatTimePeriodDate(endDate)
			}
		}

		if edu.School != "" {
			if DebugMode {
				fmt.Printf("[DEBUG] Adding education: %s - %s\n", edu.School, edu.Degree)
			}
			profile.Education = append(profile.Education, edu)
		}

	case strings.Contains(typeLower, "skill") || strings.Contains(urnLower, "fs_skill"):
		// Skills
		if name, ok := data["name"].(string); ok && name != "" {
			profile.Skills = append(profile.Skills, name)
		}

	case strings.Contains(typeField, "Language") || strings.Contains(entityUrn, "fs_language"):
		// Languages
		lang := LinkedInLanguage{}
		if name, ok := data["name"].(string); ok && name != "" {
			lang.Name = name
		}
		if proficiency, ok := data["proficiency"].(string); ok {
			lang.Proficiency = proficiency
		}
		if lang.Name != "" {
			profile.Languages = append(profile.Languages, lang)
		}

	case strings.Contains(typeField, "Certification") || strings.Contains(entityUrn, "fs_certification"):
		// Certifications
		cert := LinkedInCertification{}
		if name, ok := data["name"].(string); ok && name != "" {
			cert.Name = name
		}
		if authority, ok := data["authority"].(string); ok {
			cert.Organization = authority
		}
		// Date d'obtention
		if timePeriod, ok := data["timePeriod"].(map[string]interface{}); ok {
			if startDate, ok := timePeriod["startDate"].(map[string]interface{}); ok {
				cert.IssueDate = formatTimePeriodDate(startDate)
			}
		}
		if cert.Name != "" {
			profile.Certifications = append(profile.Certifications, cert)
		}
	}
}

// formatTimePeriodDate formate une date depuis le format timePeriod de LinkedIn
func formatTimePeriodDate(dateData map[string]interface{}) string {
	year, _ := dateData["year"].(float64)
	month, _ := dateData["month"].(float64)

	if year > 0 {
		if month > 0 {
			return fmt.Sprintf("%02d/%d", int(month), int(year))
		}
		return fmt.Sprintf("%d", int(year))
	}
	return ""
}

// decodeHTMLEntities decodes HTML entities in a string
func decodeHTMLEntities(s string) string {
	// Replace common HTML entities
	replacements := map[string]string{
		"&lt;":   "<",
		"&gt;":   ">",
		"&amp;":  "&",
		"&quot;": "\"",
		"&#39;":  "'",
		"&apos;": "'",
	}

	for entity, char := range replacements {
		s = strings.ReplaceAll(s, entity, char)
	}

	return s
}
