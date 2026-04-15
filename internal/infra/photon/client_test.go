package photon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/services/auth"

	"github.com/stretchr/testify/require"
)

func TestClientSuggestFiltersByInstitutionKind(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api", r.URL.Path)
		require.Equal(t, "8", r.URL.Query().Get("limit"))
		require.Equal(t, "ru", r.URL.Query().Get("lang"))
		require.Equal(t, "KZ", r.URL.Query().Get("countrycode"))
		require.Equal(t, "71.43", r.URL.Query().Get("lon"))
		require.Equal(t, "51.13", r.URL.Query().Get("lat"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"type": "FeatureCollection",
			"features": [
				{
					"type": "Feature",
					"properties": {
						"osm_type": "R",
						"osm_id": 17,
						"osm_key": "amenity",
						"osm_value": "school",
						"name": "Школа-лицей №17",
						"street": "Абая",
						"housenumber": "1",
						"city": "Астана",
						"country": "Казахстан"
					}
				},
				{
					"type": "Feature",
					"properties": {
						"osm_type": "R",
						"osm_id": 53,
						"osm_key": "amenity",
						"osm_value": "university",
						"name": "Nazarbayev University",
						"street": "Кабанбай батыр",
						"housenumber": "53",
						"city": "Астана",
						"country": "Казахстан"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:        server.URL,
		Lang:           "ru",
		CountryCode:    "KZ",
		DefaultLon:     "71.43",
		DefaultLat:     "51.13",
		RequestTimeout: time.Second,
	})

	schools, err := client.Suggest(context.Background(), auth.InstitutionSuggestRequest{
		Query: "лицей",
		Kind:  auth.EducationTypeSchool,
	})
	require.NoError(t, err)
	require.Len(t, schools, 1)
	require.Equal(t, "r:17", schools[0].ExternalID)
	require.Equal(t, auth.InstitutionProviderPhoton, schools[0].Provider)

	universities, err := client.Suggest(context.Background(), auth.InstitutionSuggestRequest{
		Query: "nazarbayev",
		Kind:  auth.EducationTypeUniversity,
	})
	require.NoError(t, err)
	require.Len(t, universities, 1)
	require.Equal(t, "r:53", universities[0].ExternalID)
}

func TestClientSuggestWorksWithoutGeoBias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "", r.URL.Query().Get("lon"))
		require.Equal(t, "", r.URL.Query().Get("lat"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"type": "FeatureCollection",
			"features": [
				{
					"type": "Feature",
					"properties": {
						"osm_type": "R",
						"osm_id": 2814221,
						"osm_key": "amenity",
						"osm_value": "university",
						"name": "University of California, Berkeley",
						"city": "Berkeley",
						"country": "United States"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:        server.URL,
		RequestTimeout: time.Second,
	})

	items, err := client.Suggest(context.Background(), auth.InstitutionSuggestRequest{
		Query: "berkeley",
		Kind:  auth.EducationTypeUniversity,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "University of California, Berkeley", items[0].Name)
}

func TestClientSuggestPrefixesGenericQueryByEducationType(t *testing.T) {
	t.Run("school", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "school 17", r.URL.Query().Get("q"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"features":[]}`))
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:        server.URL,
			CountryCode:    "KZ",
			RequestTimeout: time.Second,
		})

		_, err := client.Suggest(context.Background(), auth.InstitutionSuggestRequest{
			Query: "17",
			Kind:  auth.EducationTypeSchool,
		})
		require.NoError(t, err)
	})

	t.Run("university", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "university astana it", r.URL.Query().Get("q"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"features":[]}`))
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:        server.URL,
			CountryCode:    "KZ",
			RequestTimeout: time.Second,
		})

		_, err := client.Suggest(context.Background(), auth.InstitutionSuggestRequest{
			Query: "astana it",
			Kind:  auth.EducationTypeUniversity,
		})
		require.NoError(t, err)
	})
}

func TestClientSuggestSkipsUniversitySubunits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"type": "FeatureCollection",
			"features": [
				{
					"type": "Feature",
					"properties": {
						"osm_type": "R",
						"osm_id": 1,
						"osm_key": "amenity",
						"osm_value": "university",
						"name": "Hematology Dept, Enugu State University Teaching Hospital",
						"city": "Enugu",
						"country": "Nigeria"
					}
				},
				{
					"type": "Feature",
					"properties": {
						"osm_type": "R",
						"osm_id": 2,
						"osm_key": "amenity",
						"osm_value": "university",
						"name": "Astana IT University",
						"city": "Astana",
						"country": "Kazakhstan"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:        server.URL,
		CountryCode:    "KZ",
		RequestTimeout: time.Second,
	})

	items, err := client.Suggest(context.Background(), auth.InstitutionSuggestRequest{
		Query: "astana it",
		Kind:  auth.EducationTypeUniversity,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Astana IT University", items[0].Name)
}
