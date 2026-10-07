package web

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ErrPlacesNotFound is returned by GetPlaces when the job's CSV output does not
// exist. Callers use it to distinguish a missing job (404) from other errors.
var ErrPlacesNotFound = errors.New("places not found")

// Place contains business details extracted from a job's CSV output.
type Place struct {
	Title          string   `json:"title"`
	Address        string   `json:"address"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
	Link           string   `json:"link"`
	Category       string   `json:"category"`
	Phone          string   `json:"phone"`
	Website        string   `json:"website"`
	ReviewRating   float64  `json:"review_rating"`
	PhoneNumbers   []string `json:"phone_numbers"`
	Hours          string   `json:"hours"`
	HasCoordinates bool     `json:"has_coordinates"`
}

// GetPlaces parses the job's business results, including places without map coordinates.
// In web mode each job writes exactly one {id}.csv, so that file is the single
// source of truth for the map.
func (s *Service) GetPlaces(_ context.Context, id string) ([]Place, error) {
	path, err := s.csvPath(id)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("csv file not found for job %s: %w", id, ErrPlacesNotFound)
		}

		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	return parsePlaces(f)
}

// parsePlaces reads scraped results from a CSV stream and returns the places
// with their available details. Columns are resolved by header name so the
// parser tolerates reordering; the names mirror gmaps.Entry.CsvHeaders().
func parsePlaces(r io.Reader) ([]Place, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return []Place{}, nil
		}

		return nil, err
	}

	col := make(map[string]int, len(header))
	for i, name := range header {
		col[name] = i
	}

	get := func(row []string, name string) string {
		idx, ok := col[name]
		if !ok || idx >= len(row) {
			return ""
		}

		return row[idx]
	}

	places := []Place{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		lat, errLat := strconv.ParseFloat(get(row, "latitude"), 64)
		lon, errLon := strconv.ParseFloat(get(row, "longitude"), 64)

		hasCoordinates := errLat == nil && errLon == nil && validCoordinates(lat, lon) && (lat != 0 || lon != 0)
		if !hasCoordinates {
			lat, lon = 0, 0
		}

		rating, _ := strconv.ParseFloat(get(row, "review_rating"), 64)
		if !finite(rating) {
			rating = 0
		}

		places = append(places, Place{
			Title:          get(row, "title"),
			Address:        get(row, "address"),
			Latitude:       lat,
			Longitude:      lon,
			Link:           get(row, "link"),
			Category:       get(row, "category"),
			Phone:          get(row, "phone"),
			Website:        get(row, "website"),
			ReviewRating:   rating,
			HasCoordinates: hasCoordinates,
			PhoneNumbers:   parsePhoneNumbers(get(row, "phone"), get(row, "phone_numbers")),
			Hours:          formatOpeningHours(get(row, "open_hours")),
		})
	}

	return places, nil
}

func parsePhoneNumbers(primary, raw string) []string {
	var numbers []string

	_ = json.Unmarshal([]byte(raw), &numbers)
	numbers = append([]string{primary}, numbers...)
	result := make([]string, 0, len(numbers))
	seen := make(map[string]bool)
	normalize := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "")

	for _, number := range numbers {
		number = strings.TrimSpace(number)

		key := normalize.Replace(number)
		if key == "" || seen[key] {
			continue
		}

		seen[key] = true

		result = append(result, number)
	}

	return result
}

func formatOpeningHours(raw string) string {
	var hours map[string][]string
	if err := json.Unmarshal([]byte(raw), &hours); err != nil {
		return ""
	}

	days := make([]string, 0, len(hours))
	for day := range hours {
		days = append(days, day)
	}

	dayOrder := map[string]int{"monday": 1, "tuesday": 2, "wednesday": 3, "thursday": 4, "friday": 5, "saturday": 6, "sunday": 7}

	sort.Slice(days, func(i, j int) bool {
		left, right := dayOrder[strings.ToLower(days[i])], dayOrder[strings.ToLower(days[j])]
		if left == 0 {
			left = 8
		}

		if right == 0 {
			right = 8
		}

		if left == right {
			return days[i] < days[j]
		}

		return left < right
	})

	lines := make([]string, 0, len(days))
	for _, day := range days {
		if len(hours[day]) > 0 {
			lines = append(lines, day+": "+strings.Join(hours[day], ", "))
		}
	}

	return strings.Join(lines, "\n")
}

// finite reports whether f is a usable, real number (not NaN or ±Inf).
func finite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
