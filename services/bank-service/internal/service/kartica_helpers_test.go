package service

// White-box tests for unexported kartica_service helpers.
// Uses package service (not service_test) to access unexported symbols.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"banka-backend/services/bank-service/internal/domain"
)

// ─── luhnCheckDigit ───────────────────────────────────────────────────────────

func TestLuhnCheckDigit_KnownValues(t *testing.T) {
	tests := []struct {
		name   string
		digits []int
		want   int
	}{
		// Classic Luhn example: partial "497916650" should give check=3 → full "4979166503"
		{"visa example", []int{4, 9, 7, 9, 1, 6, 6, 5, 0}, 3},
		// Single zero → check must be 0 so the full number [0, 0] sums to 0 mod 10
		{"single zero", []int{0}, 0},
		// [1] → double 1 = 2; check = (10-2)%10 = 8
		{"single one", []int{1}, 8},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := luhnCheckDigit(tc.digits)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestLuhnCheckDigit_FullNumberValidation verifies that appending the check
// digit to the partial number produces a valid Luhn sequence (total sum % 10 == 0).
func TestLuhnCheckDigit_FullNumberValidation(t *testing.T) {
	partials := [][]int{
		{4, 5, 3, 9, 1, 4, 8, 8, 0, 3, 4, 3, 6, 4, 6},
		{3, 7, 1, 4, 4, 9, 6, 3, 5, 3, 9, 8, 4, 3},
		{6, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	for _, partial := range partials {
		check := luhnCheckDigit(partial)
		full := append(partial, check)

		// Validate the full number with the standard Luhn algorithm.
		sum := 0
		for i := len(full) - 1; i >= 0; i-- {
			d := full[i]
			if (len(full)-1-i)%2 == 1 {
				d *= 2
				if d > 9 {
					d -= 9
				}
			}
			sum += d
		}
		assert.Equal(t, 0, sum%10, "Luhn check failed for partial %v (check=%d)", partial, check)
	}
}

// ─── generateBrojKartice ──────────────────────────────────────────────────────

func TestGenerateBrojKartice_LengthAndPrefix(t *testing.T) {
	tests := []struct {
		tipKartice     string
		expectedIIN    string
		expectedLength int
	}{
		{domain.TipKarticaVisa, "466666", 16},
		{domain.TipKarticaMastercard, "512345", 16},
		{domain.TipKarticaDinaCard, "989100", 16},
		{domain.TipKarticaAmex, "341234", 15},
		{"UNKNOWN", "466666", 16}, // unknown falls back to Visa
	}

	for _, tc := range tests {
		t.Run(tc.tipKartice, func(t *testing.T) {
			num, err := generateBrojKartice(tc.tipKartice)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedLength, len(num),
				"card number %q should be %d digits", num, tc.expectedLength)
			assert.True(t, strings.HasPrefix(num, tc.expectedIIN),
				"card number %q should start with %q", num, tc.expectedIIN)

			// Verify Luhn checksum on the full generated number.
			sum := 0
			for i, ch := range num {
				d := int(ch - '0')
				if (len(num)-1-i)%2 == 1 {
					d *= 2
					if d > 9 {
						d -= 9
					}
				}
				sum += d
			}
			assert.Equal(t, 0, sum%10,
				"Luhn validation failed for generated card number %q", num)
		})
	}
}

func TestGenerateBrojKartice_Uniqueness(t *testing.T) {
	// Generate 10 numbers and check they're not all the same (non-deterministic).
	seen := make(map[string]struct{})
	for i := 0; i < 10; i++ {
		num, err := generateBrojKartice(domain.TipKarticaVisa)
		require.NoError(t, err)
		seen[num] = struct{}{}
	}
	// At least 2 distinct numbers expected from 10 draws over a 10^9 space.
	assert.Greater(t, len(seen), 1)
}

// ─── generateCVV ──────────────────────────────────────────────────────────────

func TestGenerateCVV_Format(t *testing.T) {
	for i := 0; i < 20; i++ {
		cvv, err := generateCVV()
		require.NoError(t, err)
		assert.Len(t, cvv, 3, "CVV %q should always be 3 chars", cvv)
		for _, ch := range cvv {
			assert.True(t, ch >= '0' && ch <= '9', "CVV char %q should be a digit", ch)
		}
	}
}

// ─── validateOvlascenoLice ────────────────────────────────────────────────────

func TestValidateOvlascenoLice(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.OvlascenoLiceInput
		wantErr error
	}{
		{
			name:    "valid",
			input:   domain.OvlascenoLiceInput{Ime: "Ana", Prezime: "Anic", EmailAdresa: "ana@test.com"},
			wantErr: nil,
		},
		{
			name:    "missing ime",
			input:   domain.OvlascenoLiceInput{Ime: "", Prezime: "Anic", EmailAdresa: "ana@test.com"},
			wantErr: domain.ErrOvlascenoLiceMissingData,
		},
		{
			name:    "missing prezime",
			input:   domain.OvlascenoLiceInput{Ime: "Ana", Prezime: "", EmailAdresa: "ana@test.com"},
			wantErr: domain.ErrOvlascenoLiceMissingData,
		},
		{
			name:    "missing email",
			input:   domain.OvlascenoLiceInput{Ime: "Ana", Prezime: "Anic", EmailAdresa: ""},
			wantErr: domain.ErrOvlascenoLiceMissingData,
		},
		{
			name:    "invalid email no @",
			input:   domain.OvlascenoLiceInput{Ime: "Ana", Prezime: "Anic", EmailAdresa: "notanemail.com"},
			wantErr: domain.ErrInvalidEmailFormat,
		},
		{
			name:    "invalid email no dot",
			input:   domain.OvlascenoLiceInput{Ime: "Ana", Prezime: "Anic", EmailAdresa: "ana@testcom"},
			wantErr: domain.ErrInvalidEmailFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOvlascenoLice(&tc.input)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
