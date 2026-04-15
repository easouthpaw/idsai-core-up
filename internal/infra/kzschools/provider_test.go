package kzschools

import (
	"context"
	"testing"

	"idsai-core-up/internal/services/auth"

	"github.com/stretchr/testify/require"
)

func TestSimplifySchoolNameUsesQuotedSchoolName(t *testing.T) {
	raw := "Коммунальное государственное учреждение «Средняя школа имени А.Розыбакиева» государственного учреждения «Отдел образования по Панфиловскому району Управления образования области Жетісу»"
	require.Equal(t, "Средняя школа имени А.Розыбакиева", simplifySchoolName(raw))
}

func TestSimplifySchoolNameHandlesNestedQuotes(t *testing.T) {
	raw := "Коммунальное государственное учреждение \"Лицей-интернат \"БІЛІМ-ИННОВАЦИЯ\" №3\" управления образования области Ұлытау"
	require.Equal(t, "Лицей-интернат \"БІЛІМ-ИННОВАЦИЯ\" №3", simplifySchoolName(raw))
}

func TestProviderSuggestFindsSchoolByCleanedName(t *testing.T) {
	provider := &Provider{
		schools: []schoolRecord{
			buildSchoolRecord(sourceSchool{
				ID:      "1",
				Name:    "Коммунальное государственное учреждение «Средняя школа имени Абая» отдела образования города Шымкент",
				Address: "Улица Абая 12",
				Region:  "г. Шымкент",
			}),
			buildSchoolRecord(sourceSchool{
				ID:      "2",
				Name:    "Коммунальное государственное учреждение «Школа-лицей №17» отдела образования города Астана",
				Address: "Проспект Тауелсиздик 8",
				Region:  "г. Астана",
			}),
		},
	}

	items, err := provider.Suggest(context.Background(), auth.InstitutionSuggestRequest{
		Query: "лицей 17",
		Kind:  auth.EducationTypeSchool,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Школа-лицей №17", items[0].Name)
	require.Equal(t, auth.InstitutionProviderKZSchools, items[0].Provider)
}

func TestProviderSuggestCanMatchAddressAndRegion(t *testing.T) {
	provider := &Provider{
		schools: []schoolRecord{
			buildSchoolRecord(sourceSchool{
				ID:      "1",
				Name:    "Коммунальное государственное учреждение «Средняя школа имени Абая» отдела образования города Шымкент",
				Address: "Улица Абая 12",
				Region:  "г. Шымкент",
			}),
		},
	}

	items, err := provider.Suggest(context.Background(), auth.InstitutionSuggestRequest{
		Query: "шымкент абая",
		Kind:  auth.EducationTypeSchool,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Средняя школа имени Абая", items[0].Name)
}
