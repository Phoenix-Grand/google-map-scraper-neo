package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	zipPattern        = regexp.MustCompile(`^\d{5}(-\d{4})?$`)
	errZIPNotFound    = errors.New("ZIP code was not found; check the ZIP or enter coordinates")
	errZIPUnavailable = errors.New("ZIP lookup is temporarily unavailable; try again or enter coordinates")
)

type zipLocation struct {
	latitude  string
	longitude string
	name      string
}

type zipResolver interface {
	Resolve(context.Context, string) (zipLocation, error)
}

type zipCodeResolver struct {
	client  *http.Client
	baseURL string
}

func normalizeZIPCode(zip string) (string, error) {
	zip = strings.TrimSpace(zip)
	if !zipPattern.MatchString(zip) {
		return "", errors.New("enter a US ZIP code (5 digits or ZIP+4)")
	}

	return zip[:5], nil
}

func validCoordinates(latitude, longitude float64) bool {
	return finite(latitude) && finite(longitude) && latitude >= -90 && latitude <= 90 && longitude >= -180 && longitude <= 180
}

func (z *zipCodeResolver) Resolve(ctx context.Context, zip string) (zipLocation, error) {
	zip, err := normalizeZIPCode(zip)
	if err != nil {
		return zipLocation{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, z.baseURL+"/us/"+zip, http.NoBody) //nolint:gosec // The server fixes baseURL; normalized ZIPs contain exactly five digits.
	if err != nil {
		return zipLocation{}, fmt.Errorf("%w: %w", errZIPUnavailable, err)
	}

	response, err := z.client.Do(request) //nolint:gosec // The request targets the fixed ZIP provider with a validated numeric path.
	if err != nil {
		return zipLocation{}, fmt.Errorf("%w: %w", errZIPUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return zipLocation{}, errZIPNotFound
	}

	if response.StatusCode != http.StatusOK {
		return zipLocation{}, errZIPUnavailable
	}

	var result struct {
		Places []struct {
			Name      string `json:"place name"`
			State     string `json:"state abbreviation"`
			Latitude  string `json:"latitude"`
			Longitude string `json:"longitude"`
		} `json:"places"`
	}

	const maxZIPResponseBytes = 64 * 1024
	if err := json.NewDecoder(io.LimitReader(response.Body, maxZIPResponseBytes)).Decode(&result); err != nil || len(result.Places) == 0 {
		return zipLocation{}, errZIPUnavailable
	}

	place := result.Places[0]
	latitude, latErr := strconv.ParseFloat(place.Latitude, 64)

	longitude, lonErr := strconv.ParseFloat(place.Longitude, 64)
	if latErr != nil || lonErr != nil || !validCoordinates(latitude, longitude) {
		return zipLocation{}, errZIPUnavailable
	}

	return zipLocation{latitude: place.Latitude, longitude: place.Longitude, name: strings.TrimSpace(place.Name + ", " + place.State)}, nil
}

func (s *Server) prepareLocation(ctx context.Context, data *JobData) error {
	data.Lat = strings.TrimSpace(data.Lat)
	data.Lon = strings.TrimSpace(data.Lon)

	data.ZIPCode = strings.TrimSpace(data.ZIPCode)
	if data.ZIPCode == "" {
		return nil
	}

	zip, err := normalizeZIPCode(data.ZIPCode)
	if err != nil {
		return err
	}

	location, err := s.zipResolver.Resolve(ctx, zip)
	if err != nil {
		return err
	}

	data.ZIPCode = zip
	data.Lat = location.latitude
	data.Lon = location.longitude
	data.LocationName = location.name

	return nil
}

func newZIPResolver() zipResolver {
	return &zipCodeResolver{client: &http.Client{Timeout: 8 * time.Second}, baseURL: "https://api.zippopotam.us"}
}

func (s *Server) prepareJob(ctx context.Context, job *Job) error {
	if job.Data.Zoom == 0 {
		job.Data.Zoom = 15
	}

	if err := job.Validate(); err != nil {
		return err
	}

	return s.prepareLocation(ctx, &job.Data)
}

func locationErrorStatus(err error) int {
	if errors.Is(err, errZIPUnavailable) {
		return http.StatusServiceUnavailable
	}

	return http.StatusUnprocessableEntity
}
