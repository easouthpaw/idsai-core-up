package photon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"idsai-core-up/internal/services/auth"
)

var ErrUnavailable = errors.New("photon request failed")

type Config struct {
	BaseURL        string
	Lang           string
	CountryCode    string
	DefaultLon     string
	DefaultLat     string
	RequestTimeout time.Duration
}

type Client struct {
	baseURL     string
	lang        string
	countryCode string
	defaultLon  string
	defaultLat  string
	httpClient  *http.Client
}

type searchResponse struct {
	Features []feature `json:"features"`
}

type feature struct {
	Properties featureProperties `json:"properties"`
}

type featureProperties struct {
	OSMType     string `json:"osm_type"`
	OSMID       int64  `json:"osm_id"`
	OSMKey      string `json:"osm_key"`
	OSMValue    string `json:"osm_value"`
	Name        string `json:"name"`
	Street      string `json:"street"`
	Housenumber string `json:"housenumber"`
	Locality    string `json:"locality"`
	District    string `json:"district"`
	City        string `json:"city"`
	County      string `json:"county"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Postcode    string `json:"postcode"`
}

func New(cfg Config) *Client {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://photon.komoot.io"
	}
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	return &Client{
		baseURL:     baseURL,
		lang:        strings.TrimSpace(cfg.Lang),
		countryCode: strings.ToUpper(strings.TrimSpace(cfg.CountryCode)),
		defaultLon:  strings.TrimSpace(cfg.DefaultLon),
		defaultLat:  strings.TrimSpace(cfg.DefaultLat),
		httpClient:  &http.Client{Timeout: timeout},
	}
}

func (c *Client) Suggest(ctx context.Context, in auth.InstitutionSuggestRequest) ([]auth.InstitutionSuggestion, error) {
	kind := strings.ToUpper(strings.TrimSpace(in.Kind))
	if kind != auth.EducationTypeUniversity && kind != auth.EducationTypeSchool {
		return nil, auth.ErrInvalidInput
	}

	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, auth.ErrInvalidInput
	}

	items, err := c.fetch(ctx, query, in)
	if err != nil {
		return nil, err
	}
	return filterInstitutionSuggestions(kind, items), nil
}

func (c *Client) fetch(ctx context.Context, query string, in auth.InstitutionSuggestRequest) ([]feature, error) {
	endpoint, err := url.Parse(c.baseURL + "/api")
	if err != nil {
		return nil, ErrUnavailable
	}

	params := endpoint.Query()
	kind := strings.ToUpper(strings.TrimSpace(in.Kind))
	params.Set("q", queryWithKindPrefix(kind, query))
	params.Set("limit", "8")
	if c.lang != "" {
		params.Set("lang", c.lang)
	}
	if c.countryCode != "" {
		params.Set("countrycode", c.countryCode)
	}
	switch {
	case in.Lon != nil && in.Lat != nil:
		params.Set("lon", fmt.Sprintf("%g", *in.Lon))
		params.Set("lat", fmt.Sprintf("%g", *in.Lat))
	case c.defaultLon != "" && c.defaultLat != "":
		params.Set("lon", c.defaultLon)
		params.Set("lat", c.defaultLat)
	}
	if kind == auth.EducationTypeSchool {
		params.Add("osm_tag", "amenity:school")
	} else {
		params.Add("osm_tag", "amenity:university")
		params.Add("osm_tag", "amenity:college")
	}
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status=%d", ErrUnavailable, resp.StatusCode)
	}

	var payload searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return payload.Features, nil
}

func filterInstitutionSuggestions(kind string, items []feature) []auth.InstitutionSuggestion {
	out := make([]auth.InstitutionSuggestion, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		suggestion := auth.InstitutionSuggestion{
			Provider:   auth.InstitutionProviderPhoton,
			ExternalID: externalID(item.Properties),
			Name:       strings.TrimSpace(item.Properties.Name),
			Address:    formatAddress(item.Properties),
		}
		if suggestion.Name == "" || !matchesInstitutionKind(kind, item.Properties, suggestion) {
			continue
		}

		key := strings.ToLower(suggestion.ExternalID + "|" + suggestion.Name + "|" + suggestion.Address)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, suggestion)
		if len(out) >= 8 {
			break
		}
	}

	return out
}

func externalID(props featureProperties) string {
	osmType := strings.TrimSpace(props.OSMType)
	if osmType == "" && props.OSMID == 0 {
		return ""
	}
	return strings.ToLower(osmType) + ":" + fmt.Sprintf("%d", props.OSMID)
}

func formatAddress(props featureProperties) string {
	parts := []string{
		strings.TrimSpace(strings.TrimSpace(props.Street + " " + props.Housenumber)),
		strings.TrimSpace(props.Locality),
		strings.TrimSpace(props.District),
		strings.TrimSpace(props.City),
		strings.TrimSpace(props.County),
		strings.TrimSpace(props.State),
		strings.TrimSpace(props.Postcode),
		strings.TrimSpace(props.Country),
	}
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		key := strings.ToLower(part)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, part)
	}
	return strings.Join(out, ", ")
}

func matchesInstitutionKind(kind string, props featureProperties, suggestion auth.InstitutionSuggestion) bool {
	osmValue := strings.ToLower(strings.TrimSpace(props.OSMValue))
	nameText := strings.ToLower(strings.TrimSpace(suggestion.Name))

	if containsAny(nameText, []string{"автошкол", "driving school"}) {
		return false
	}
	if containsAny(nameText, []string{"hospital", "clinic", "dept", "department", "theatre", "лаборатор", "больниц", "клиник"}) {
		return false
	}

	switch kind {
	case auth.EducationTypeSchool:
		if osmValue == "school" {
			return true
		}
		return containsAny(nameText, []string{"school", "школ", "лицей", "lyceum", "гимназ", "gymnasium"})
	default:
		if osmValue == "university" || osmValue == "college" {
			return true
		}
		return containsAny(nameText, []string{"university", "универс", "вуз", "institute", "институт", "academy", "академ"})
	}
}

func containsAny(source string, terms []string) bool {
	for _, term := range terms {
		if term != "" && strings.Contains(source, term) {
			return true
		}
	}
	return false
}

func queryWithKindPrefix(kind, query string) string {
	text := strings.TrimSpace(query)
	if text == "" {
		return text
	}
	lower := strings.ToLower(text)
	switch kind {
	case auth.EducationTypeSchool:
		if containsAny(lower, []string{"school", "школ", "лицей", "гимназ", "lyceum", "gymnasium"}) {
			return text
		}
		if hasCyrillic(text) {
			return "школа " + text
		}
		return "school " + text
	default:
		if containsAny(lower, []string{"university", "универс", "вуз", "институт", "академ", "college", "institute", "academy"}) {
			return text
		}
		if hasCyrillic(text) {
			return "университет " + text
		}
		return "university " + text
	}
}

func hasCyrillic(text string) bool {
	for _, r := range text {
		if (r >= 'А' && r <= 'я') || r == 'Ё' || r == 'ё' {
			return true
		}
	}
	return false
}
