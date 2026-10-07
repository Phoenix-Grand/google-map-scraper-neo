package webrunner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gosom/scrapemate"
	"golang.org/x/sync/errgroup"

	"github.com/gosom/google-maps-scraper/gmaps"
)

type businessResultWriter struct {
	writer    scrapemate.ResultWriter
	latitude  float64
	longitude float64
	radius    float64
}

func (w *businessResultWriter) Run(ctx context.Context, input <-chan scrapemate.Result) error {
	// Like the CSV writer, drain completed results after the scrape deadline.
	// The scraper closes input when it finishes; writer errors stop forwarding.
	group, writerCtx := errgroup.WithContext(context.WithoutCancel(ctx))
	filtered := make(chan scrapemate.Result)

	group.Go(func() error {
		defer close(filtered)

		for {
			select {
			case <-writerCtx.Done():
				return writerCtx.Err()
			case result, ok := <-input:
				if !ok {
					return nil
				}

				entries, err := w.filter(result.Data)
				if err != nil {
					return err
				}

				if len(entries) == 0 {
					continue
				}

				result.Data = entries
				select {
				case filtered <- result:
				case <-writerCtx.Done():
					return writerCtx.Err()
				}
			}
		}
	})
	group.Go(func() error { return w.writer.Run(writerCtx, filtered) })

	return group.Wait()
}

func (w *businessResultWriter) filter(data any) ([]*businessCSVEntry, error) {
	var entries []*gmaps.Entry

	switch value := data.(type) {
	case *gmaps.Entry:
		entries = []*gmaps.Entry{value}
	case []*gmaps.Entry:
		entries = value
	default:
		return nil, fmt.Errorf("unexpected radius result type: %T", data)
	}

	filtered := make([]*businessCSVEntry, 0, len(entries))
	for _, entry := range entries {
		if entry != nil && (w.radius <= 0 || entry.WithinRadius(w.latitude, w.longitude, w.radius)) {
			filtered = append(filtered, &businessCSVEntry{Entry: entry})
		}
	}

	return filtered, nil
}

// Only web exports add a column, preserving the schema of resumable CLI files.
type businessCSVEntry struct {
	*gmaps.Entry
}

func (e *businessCSVEntry) CsvHeaders() []string {
	return append(e.Entry.CsvHeaders(), "phone_numbers")
}

func (e *businessCSVEntry) CsvRow() []string {
	phones, _ := json.Marshal(e.PhoneNumbers) // A slice of strings always encodes successfully.

	return append(e.Entry.CsvRow(), string(phones))
}
