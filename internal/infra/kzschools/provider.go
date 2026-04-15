package kzschools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"idsai-core-up/internal/services/auth"
)

const maxSuggestions = 8

// Source: https://data.egov.kz/datasets/view?index=onirler_oblystar_kalalar_boi4
// Official registry of general education schools of Kazakhstan, exported from v7.

//go:embed schools.json
var schoolsJSON []byte

var schoolCatalog = mustLoadCatalog()

type sourceSchool struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Kind    string `json:"kind"`
}

type schoolRecord struct {
	suggestion auth.InstitutionSuggestion
	searchName string
	searchRaw  string
	searchAddr string
	searchAll  string
}

type Provider struct {
	schools []schoolRecord
}

func New() *Provider {
	return &Provider{schools: schoolCatalog}
}

func (p *Provider) Suggest(ctx context.Context, in auth.InstitutionSuggestRequest) ([]auth.InstitutionSuggestion, error) {
	if strings.ToUpper(strings.TrimSpace(in.Kind)) != auth.EducationTypeSchool {
		return nil, nil
	}

	query := normalizeSearchText(in.Query)
	if query == "" {
		return nil, auth.ErrInvalidInput
	}

	type scoredRecord struct {
		item  auth.InstitutionSuggestion
		score int
		order int
	}

	matches := make([]scoredRecord, 0, maxSuggestions*2)
	for idx, school := range p.schools {
		if idx%128 == 0 && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		score, ok := scoreSchoolMatch(query, school)
		if !ok {
			continue
		}
		matches = append(matches, scoredRecord{
			item:  school.suggestion,
			score: score,
			order: idx,
		})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score < matches[j].score
		}
		if matches[i].item.Name != matches[j].item.Name {
			return matches[i].item.Name < matches[j].item.Name
		}
		return matches[i].order < matches[j].order
	})

	out := make([]auth.InstitutionSuggestion, 0, maxSuggestions)
	for _, match := range matches {
		out = append(out, match.item)
		if len(out) >= maxSuggestions {
			break
		}
	}
	return out, nil
}

func mustLoadCatalog() []schoolRecord {
	var src []sourceSchool
	if err := json.Unmarshal(schoolsJSON, &src); err != nil {
		panic(fmt.Errorf("load kz schools catalog: %w", err))
	}

	items := make([]schoolRecord, 0, len(src))
	for _, item := range src {
		record := buildSchoolRecord(item)
		if record.suggestion.Name == "" {
			continue
		}
		items = append(items, record)
	}
	return items
}

func buildSchoolRecord(item sourceSchool) schoolRecord {
	name := simplifySchoolName(item.Name)
	address := joinUniqueParts(item.Address, item.Region)
	rawName := cleanDisplayText(item.Name)

	suggestion := auth.InstitutionSuggestion{
		Provider:   auth.InstitutionProviderKZSchools,
		ExternalID: "egov-school:" + strings.TrimSpace(item.ID),
		Name:       name,
		Address:    address,
	}
	if suggestion.Name == "" {
		suggestion.Name = rawName
	}

	searchName := normalizeSearchText(suggestion.Name)
	searchRaw := normalizeSearchText(rawName)
	searchAddr := normalizeSearchText(address)
	return schoolRecord{
		suggestion: suggestion,
		searchName: searchName,
		searchRaw:  searchRaw,
		searchAddr: searchAddr,
		searchAll:  strings.TrimSpace(searchName + " " + searchRaw + " " + searchAddr),
	}
}

func scoreSchoolMatch(query string, school schoolRecord) (int, bool) {
	terms := strings.Fields(query)
	if len(terms) == 0 || !containsAllTerms(school.searchAll, terms) {
		return 0, false
	}

	switch {
	case school.searchName == query:
		return 0, true
	case strings.HasPrefix(school.searchName, query):
		return 10, true
	case hasWordPrefix(school.searchName, query):
		return 20, true
	case strings.Contains(school.searchName, query):
		return 30 + strings.Index(school.searchName, query), true
	case strings.HasPrefix(school.searchRaw, query):
		return 50, true
	case hasWordPrefix(school.searchRaw, query):
		return 60, true
	case strings.Contains(school.searchRaw, query):
		return 70 + strings.Index(school.searchRaw, query), true
	case strings.HasPrefix(school.searchAddr, query):
		return 90, true
	case hasWordPrefix(school.searchAddr, query):
		return 100, true
	case strings.Contains(school.searchAddr, query):
		return 110 + strings.Index(school.searchAddr, query), true
	default:
		return 120, true
	}
}

func containsAllTerms(source string, terms []string) bool {
	for _, term := range terms {
		if term == "" {
			continue
		}
		if !strings.Contains(source, term) {
			return false
		}
	}
	return true
}

func hasWordPrefix(source, query string) bool {
	for _, part := range strings.Fields(source) {
		if strings.HasPrefix(part, query) {
			return true
		}
	}
	return false
}

func simplifySchoolName(raw string) string {
	text := cleanDisplayText(raw)
	if text == "" {
		return ""
	}
	if candidate := extractSmartQuoted(text); candidate != "" {
		return cleanDisplayText(candidate)
	}
	if candidate := extractASCIIQuoted(text); candidate != "" {
		return cleanDisplayText(candidate)
	}

	replaced := stripLeadingEntityForm(text)
	replaced = stripOwnershipTail(replaced)
	replaced = strings.TrimSpace(strings.Trim(replaced, ",.-"))
	if replaced == "" {
		return text
	}
	return cleanDisplayText(replaced)
}

func extractSmartQuoted(text string) string {
	start := strings.IndexRune(text, '«')
	if start < 0 {
		return ""
	}
	rest := text[start+len("«"):]
	end := strings.IndexRune(rest, '»')
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func extractASCIIQuoted(text string) string {
	start := strings.IndexRune(text, '"')
	if start < 0 {
		return ""
	}
	end := strings.LastIndex(text, "\"")
	if end <= start {
		return ""
	}
	return text[start+1 : end]
}

func stripLeadingEntityForm(text string) string {
	lower := strings.ToLower(text)
	prefixes := []string{
		"коммунальное государственное учреждение ",
		"коммунальное государственное учереждение ",
		"коммунальное госудрственное учреждение ",
		"государственное учреждение ",
		"товарищество с ограниченной ответственностью ",
		"частное учреждение ",
		"республиканское государственное учреждение ",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSpace(text[len(prefix):])
		}
	}
	return text
}

func stripOwnershipTail(text string) string {
	lower := strings.ToLower(text)
	cutMarkers := []string{
		" государственного учреждения ",
		" государственого учреждения ",
		" отдела образования ",
		" управления образования ",
		" управления человеческого потенциала ",
		" акимата ",
	}
	cutAt := -1
	for _, marker := range cutMarkers {
		idx := strings.Index(lower, marker)
		if idx >= 0 && (cutAt == -1 || idx < cutAt) {
			cutAt = idx
		}
	}
	if cutAt >= 0 {
		return strings.TrimSpace(text[:cutAt])
	}
	return text
}

func joinUniqueParts(parts ...string) string {
	items := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := cleanDisplayText(part)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, value)
	}
	return strings.Join(items, ", ")
}

func cleanDisplayText(text string) string {
	text = strings.NewReplacer(
		"–", "-",
		"—", "-",
		"“", "\"",
		"”", "\"",
	).Replace(strings.TrimSpace(text))
	return strings.Join(strings.Fields(text), " ")
}

func normalizeSearchText(text string) string {
	text = strings.ToLower(cleanDisplayText(text))
	text = strings.NewReplacer(
		"ё", "е",
		"ә", "а",
		"ғ", "г",
		"қ", "к",
		"ң", "н",
		"ө", "о",
		"ұ", "у",
		"ү", "у",
		"һ", "х",
		"і", "и",
		"№", " ",
		"#", " ",
	).Replace(text)

	var b strings.Builder
	b.Grow(len(text))
	space := true
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			space = false
			continue
		}
		if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}
