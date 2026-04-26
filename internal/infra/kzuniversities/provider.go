package kzuniversities

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"idsai-core-up/internal/services/auth"
)

const maxSuggestions = 8

type record struct {
	suggestion auth.InstitutionSuggestion
	searchAll  string
	// key is the university_key suffix used in faculty codes (e.g. "ENU", "KBTU")
	Key string
}

var catalog = buildCatalog()

type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) Suggest(_ context.Context, in auth.InstitutionSuggestRequest) ([]auth.InstitutionSuggestion, error) {
	if strings.ToUpper(strings.TrimSpace(in.Kind)) != auth.EducationTypeUniversity {
		return nil, nil
	}
	norm := normalize(in.Query)
	if norm == "" {
		return nil, auth.ErrInvalidInput
	}
	// Also try with Cyrillic→Latin transliteration so "аиту"→"aitu" and vice versa.
	translit := strings.TrimSpace(cyrToLat.Replace(norm))

	type scored struct {
		item  auth.InstitutionSuggestion
		score int
		order int
	}
	matches := make([]scored, 0, maxSuggestions*2)
	for i, r := range catalog {
		score, ok := scoreRecord(norm, translit, r)
		if !ok {
			continue
		}
		matches = append(matches, scored{item: r.suggestion, score: score, order: i})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score < matches[j].score
		}
		return matches[i].order < matches[j].order
	})
	out := make([]auth.InstitutionSuggestion, 0, maxSuggestions)
	for _, m := range matches {
		out = append(out, m.item)
		if len(out) >= maxSuggestions {
			break
		}
	}
	return out, nil
}

// scoreRecord checks if either the original query or its transliterated form
// matches the record, returning the best score.
func scoreRecord(query, queryTranslit string, r record) (int, bool) {
	score1, ok1 := scoreSingleQuery(query, r.searchAll)
	score2, ok2 := scoreSingleQuery(queryTranslit, r.searchAll)
	switch {
	case ok1 && ok2:
		if score1 <= score2 {
			return score1, true
		}
		return score2, true
	case ok1:
		return score1, true
	case ok2:
		return score2, true
	default:
		return 0, false
	}
}

func scoreSingleQuery(query, searchAll string) (int, bool) {
	if query == "" {
		return 0, false
	}
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return 0, false
	}
	for _, t := range terms {
		if !strings.Contains(searchAll, t) {
			return 0, false
		}
	}
	switch {
	case searchAll == query:
		return 0, true
	case strings.HasPrefix(searchAll, query):
		return 10, true
	case strings.Contains(searchAll, query):
		return 30, true
	default:
		return 50, true
	}
}

// cyrToLat converts Cyrillic letters to approximate Latin equivalents so that
// e.g. "аиту" and "aitu" both match the same query.
var cyrToLat = strings.NewReplacer(
	"а", "a", "б", "b", "в", "v", "г", "g", "д", "d",
	"е", "e", "ж", "zh", "з", "z", "и", "i", "й", "y",
	"к", "k", "л", "l", "м", "m", "н", "n", "о", "o",
	"п", "p", "р", "r", "с", "s", "т", "t", "у", "u",
	"ф", "f", "х", "h", "ц", "ts", "ч", "ch", "ш", "sh",
	"щ", "sch", "ы", "y", "э", "e", "ю", "yu", "я", "ya",
)

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer(
		"ё", "е", "ә", "а", "ғ", "г", "қ", "к",
		"ң", "н", "ө", "о", "ұ", "у", "ү", "у",
		"һ", "х", "і", "и",
	).Replace(s)
	var b strings.Builder
	space := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			space = false
		} else if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

// normalizeWithTranslit returns both the normalized form and its Cyrillic→Latin
// transliteration separated by a space, so a single Contains check covers both.
func normalizeWithTranslit(s string) string {
	norm := normalize(s)
	lat := strings.TrimSpace(cyrToLat.Replace(norm))
	if lat == norm {
		return norm
	}
	return norm + " " + lat
}

type uniDef struct {
	name    string
	address string
	key     string
	aliases []string
}

func buildCatalog() []record {
	defs := []uniDef{
		{
			name:    "ЕНУ им. Л.Н. Гумилева",
			address: "010000, Астана, ул. Сатпаева, 2",
			key:     "ENU",
			aliases: []string{"enu", "ену", "евразийский", "eurasian national university", "гумилева", "gumilyov"},
		},
		{
			name:    "КазНУ им. аль-Фараби",
			address: "050040, Алматы, пр. аль-Фараби, 71",
			key:     "KAZNU",
			aliases: []string{"казну", "kaznu", "казахский национальный", "аль-фараби", "al-farabi", "farabi"},
		},
		{
			name:    "КБТУ (Казахстанско-Британский технический университет)",
			address: "050000, Алматы, ул. Толе Би, 59",
			key:     "KBTU",
			aliases: []string{"кбту", "kbtu", "казахстанско-британский", "british technical"},
		},
		{
			name:    "Назарбаев Университет",
			address: "010000, Астана, пр. Кабанбай батыра, 53",
			key:     "NU",
			aliases: []string{"назарбаев", "nazarbayev", "nu", "нун"},
		},
		{
			name:    "МУИТ (Международный университет информационных технологий)",
			address: "050000, Алматы, ул. Манаса, 34/1",
			key:     "MUIT",
			aliases: []string{"муит", "muit", "международный университет информационных", "iitu"},
		},
		{
			name:    "AITU (Astana IT University)",
			address: "010000, Астана, ул. Манаса, 35",
			key:     "AITU",
			aliases: []string{"aitu", "астана ит", "astana it", "астана it university"},
		},
		{
			name:    "Satbayev University (КазНТУ им. К.И. Сатпаева)",
			address: "050013, Алматы, ул. Сатпаева, 22",
			key:     "SAT",
			aliases: []string{"сатбаев", "satbayev", "казнту", "kazntu", "сатпаева", "satpayev"},
		},
		{
			name:    "SDU University (Университет им. С. Демиреля)",
			address: "080003, Каскелен, ул. Абылай хана, 1/1",
			key:     "SDU",
			aliases: []string{"sdu", "suleyman demirel", "демирель", "демиреля", "demirel"},
		},
	}

	records := make([]record, 0, len(defs))
	for _, d := range defs {
		parts := []string{normalizeWithTranslit(d.name), normalize(d.address), normalizeWithTranslit(d.key)}
		for _, a := range d.aliases {
			parts = append(parts, normalizeWithTranslit(a))
		}
		records = append(records, record{
			suggestion: auth.InstitutionSuggestion{
				Provider:   "kzuniversities",
				ExternalID: "kzuni:" + d.key,
				Name:       d.name,
				Address:    d.address,
			},
			searchAll: strings.Join(parts, " "),
			Key:       d.key,
		})
	}
	return records
}
