// asterix/category_test.go
package asterix

import (
	"fmt"
	"testing"
)

func TestCategoryString(t *testing.T) {
	testCases := []struct {
		cat      Category
		expected string
	}{
		{Cat021, "CAT021"},
		{Cat048, "CAT048"},
		{Cat062, "CAT062"},
		{Cat063, "CAT063"},
		{Category(1), "CAT001"},
		{Category(255), "CAT255"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Category%d", tc.cat), func(t *testing.T) {
			result := tc.cat.String()
			if result != tc.expected {
				t.Errorf("String() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestCategoryIsValid(t *testing.T) {
	testCases := []struct {
		cat      Category
		expected bool
	}{
		{Cat021, true},
		{Cat048, true},
		{Cat062, true},
		{Cat063, true},
		{Category(1), true},   // Valid range
		{Category(255), true}, // Valid range
		{Category(0), false},  // Invalid - 0 is not a valid category
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Category%d", tc.cat), func(t *testing.T) {
			result := tc.cat.IsValid()
			if result != tc.expected {
				t.Errorf("IsValid() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCategoryIsBlockable(t *testing.T) {
	testCases := []struct {
		cat      Category
		expected bool
	}{
		{Cat021, true},
		{Cat048, true},
		{Cat062, true},
		{Cat063, true},
		{Cat065, true},
		{Cat247, true},
		{Cat252, false},
		{Category(200), false}, // Unknown category
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Category%d", tc.cat), func(t *testing.T) {
			result := tc.cat.IsBlockable()
			if result != tc.expected {
				t.Errorf("IsBlockable() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCategoryByte(t *testing.T) {
	testCases := []struct {
		cat      Category
		expected byte
	}{
		{Cat021, 21},
		{Cat048, 48},
		{Cat062, 62},
		{Category(255), 255},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Category%d", tc.cat), func(t *testing.T) {
			result := tc.cat.Byte()
			if result != tc.expected {
				t.Errorf("Byte() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestCategoryFromByte(t *testing.T) {
	testCases := []struct {
		b        byte
		expected Category
	}{
		{21, Cat021},
		{48, Cat048},
		{62, Cat062},
		{255, Category(255)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Byte%d", tc.b), func(t *testing.T) {
			result := CategoryFromByte(tc.b)
			if result != tc.expected {
				t.Errorf("CategoryFromByte() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestGetCategoryInfo(t *testing.T) {
	testCases := []struct {
		cat         Category
		expectedNum Category
		expectedVer string
	}{
		{Cat021, Cat021, "2.6"},
		{Cat048, Cat048, "1.30"},
		{Cat062, Cat062, "1.19"},
		{Cat063, Cat063, "1.7"},
		{Category(200), Category(200), ""}, // Unknown category
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Category%d", tc.cat), func(t *testing.T) {
			info := GetCategoryInfo(tc.cat)

			if info.Number != tc.expectedNum {
				t.Errorf("Number = %d, want %d", info.Number, tc.expectedNum)
			}

			if info.Version != tc.expectedVer {
				t.Errorf("Version = %q, want %q", info.Version, tc.expectedVer)
			}

			// Name should always be set
			if info.Name == "" {
				t.Error("Name is empty")
			}

			// Description should always be set
			if info.Description == "" {
				t.Error("Description is empty")
			}
		})
	}
}
