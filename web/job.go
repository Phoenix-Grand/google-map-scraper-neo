package web

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

var jobs []Job

const (
	StatusPending = "pending"
	StatusWorking = "working"
	StatusOK      = "ok"
	StatusFailed  = "failed"
)

type SelectParams struct {
	Status string
	Limit  int
	Offset int
}

// JobPage contains one page of jobs and the metadata needed to navigate it.
type JobPage struct {
	Jobs        []Job
	CurrentPage int
	TotalPages  int
	Total       int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	HasPages    bool
}

type JobRepository interface {
	Get(context.Context, string) (Job, error)
	Create(context.Context, *Job) error
	Delete(context.Context, string) error
	Select(context.Context, SelectParams) ([]Job, error)
	Update(context.Context, *Job) error
}

type Job struct {
	ID     string
	Name   string
	Date   time.Time
	Status string
	Data   JobData
}

func (j *Job) Validate() error {
	if j.ID == "" {
		return errors.New("missing id")
	}

	if j.Name == "" {
		return errors.New("missing name")
	}

	if j.Status == "" {
		return errors.New("missing status")
	}

	if j.Date.IsZero() {
		return errors.New("missing date")
	}

	if err := j.Data.Validate(); err != nil {
		return err
	}

	return nil
}

type JobData struct {
	Keywords     []string      `json:"keywords"`
	Lang         string        `json:"lang"`
	Zoom         int           `json:"zoom"`
	Lat          string        `json:"lat"`
	Lon          string        `json:"lon"`
	FastMode     bool          `json:"fast_mode"`
	Radius       int           `json:"radius"`
	ZIPCode      string        `json:"zip_code,omitempty"`
	RadiusMiles  float64       `json:"radius_miles,omitempty"`
	LocationName string        `json:"location_name,omitempty"`
	Depth        int           `json:"depth"`
	Email        bool          `json:"email"`
	ExtraReviews bool          `json:"extra_reviews"`
	MaxTime      time.Duration `json:"max_time"`
	Proxies      []string      `json:"proxies"`
}

func (d *JobData) Validate() error {
	if len(d.Keywords) == 0 {
		return errors.New("missing keywords")
	}

	if d.Lang == "" {
		return errors.New("missing lang")
	}

	if len(d.Lang) != 2 {
		return errors.New("invalid lang")
	}

	if d.Depth == 0 {
		return errors.New("missing depth")
	}

	if d.MaxTime == 0 {
		return errors.New("missing max time")
	}

	if d.ZIPCode != "" {
		if _, err := normalizeZIPCode(d.ZIPCode); err != nil {
			return err
		}
	}

	if !finite(d.RadiusMiles) || d.RadiusMiles < 0 || !finite(d.RadiusMiles*metersPerMile) {
		return errors.New("radius in miles must be a positive number")
	}

	if d.Radius < 0 {
		return errors.New("radius in meters cannot be negative")
	}

	if d.Zoom != 0 && (d.Zoom < 1 || d.Zoom > 21) {
		return errors.New("zoom must be between 1 and 21")
	}

	if d.ZIPCode != "" {
		return nil // Coordinates are resolved from the ZIP before the job is saved.
	}

	if (d.Lat == "") != (d.Lon == "") {
		return errors.New("provide both latitude and longitude")
	}

	if d.Lat != "" {
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(d.Lat), 64)

		lon, lonErr := strconv.ParseFloat(strings.TrimSpace(d.Lon), 64)
		if latErr != nil || lonErr != nil || !validCoordinates(lat, lon) {
			return errors.New("invalid latitude or longitude")
		}
	}

	if d.RadiusMiles > 0 && (d.Lat == "" || d.Lon == "") {
		return errors.New("a miles radius requires a ZIP code or both coordinates")
	}

	if d.FastMode && (d.Lat == "" || d.Lon == "") {
		return errors.New("missing geo coordinates")
	}

	return nil
}

const metersPerMile = 1609.344

// RadiusMeters converts the miles radius, falling back to the legacy fast-mode radius.
func (d *JobData) RadiusMeters() float64 {
	if d.RadiusMiles > 0 {
		return d.RadiusMiles * metersPerMile
	}

	if d.Radius > 0 {
		return float64(d.Radius)
	}

	return 10000 // Preserve the existing 10 km fast-mode default.
}
