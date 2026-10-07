//nolint:testpackage // exercises the internal result writer used by both web scrape modes.
package webrunner

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"math"
	"testing"

	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/adapters/writers/csvwriter"

	"github.com/gosom/google-maps-scraper/gmaps"
)

func TestRadiusWriterFiltersSingleAndBatchResults(t *testing.T) {
	var output bytes.Buffer

	writer := &businessResultWriter{writer: csvwriter.NewCsvWriter(csv.NewWriter(&output)), latitude: 39.7491, longitude: -104.9946, radius: 1609.344}

	input := make(chan scrapemate.Result, 3)
	input <- scrapemate.Result{Data: &gmaps.Entry{Title: "Inside single", Latitude: 39.75, Longtitude: -104.99}}

	input <- scrapemate.Result{Data: &gmaps.Entry{Title: "Outside single", Latitude: 40, Longtitude: -105}}

	input <- scrapemate.Result{Data: []*gmaps.Entry{
		{Title: "Inside batch", Latitude: 39.7491, Longtitude: -104.9946},
		{Title: "Outside batch", Latitude: 40, Longtitude: -105},
		{Title: "Unknown coordinates"},
		{Title: "Nonfinite", Latitude: math.NaN(), Longtitude: -105},
		nil,
	}}

	close(input)
	// Completed results must still be written when the scrape deadline is reached.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := writer.Run(ctx, input); err != nil {
		t.Fatal(err)
	}

	rows, err := csv.NewReader(&output).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 3 || rows[1][2] != "Inside single" || rows[2][2] != "Inside batch" {
		t.Fatalf("unexpected radius-filtered CSV: %#v", rows)
	}
}

type failingResultWriter struct {
	err error
}

func (w *failingResultWriter) Run(_ context.Context, _ <-chan scrapemate.Result) error {
	return w.err
}

func TestRadiusWriterReturnsOutputFailure(t *testing.T) {
	want := errors.New("output failed")
	writer := &businessResultWriter{writer: &failingResultWriter{err: want}, latitude: 1, longitude: 2, radius: 100}

	input := make(chan scrapemate.Result, 1)
	input <- scrapemate.Result{Data: &gmaps.Entry{Latitude: 1, Longtitude: 2}}

	if err := writer.Run(context.Background(), input); !errors.Is(err, want) {
		t.Fatalf("expected output error, got %v", err)
	}
}

func TestBusinessWriterWithoutRadiusPreservesUnmappedResults(t *testing.T) {
	var output bytes.Buffer

	writer := &businessResultWriter{writer: csvwriter.NewCsvWriter(csv.NewWriter(&output))}
	input := make(chan scrapemate.Result, 1)

	entry := &gmaps.Entry{Title: "Unmapped", Phone: "555", PhoneNumbers: []string{"555", "556"}}
	input <- scrapemate.Result{Data: entry}

	close(input)

	if err := writer.Run(context.Background(), input); err != nil {
		t.Fatal(err)
	}

	rows, err := csv.NewReader(&output).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	lastColumn := len(rows[0]) - 1
	if len(rows) != 2 || rows[0][lastColumn] != "phone_numbers" || rows[1][lastColumn] != `["555","556"]` {
		t.Fatalf("unexpected web business CSV: %#v", rows)
	}

	if len(entry.CsvHeaders()) != lastColumn {
		t.Fatal("web export changed the CLI CSV schema")
	}
}
