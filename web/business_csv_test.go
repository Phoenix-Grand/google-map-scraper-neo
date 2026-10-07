//nolint:testpackage // exercises internal CSV parsing and download routing together.
package web

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestBusinessDownloadPreservesAllDetails(t *testing.T) {
	dir := t.TempDir()
	id := "11111111-1111-1111-1111-111111111111"

	var source bytes.Buffer

	writer := csv.NewWriter(&source)
	if err := writer.WriteAll([][]string{
		{"title", "website", "address", "phone", "phone_numbers", "open_hours", "latitude", "longitude"},
		{"Cafe, Test", "https://cafe.test", "1 Main St", "(303) 555-0100", `["3035550100","3035550101"]`, `{"Sunday":["Closed"],"Monday":["9 AM–12 PM","1 PM–5 PM"]}`, "", ""},
		{"Missing details", "", "", "", "null", "null", "1", "2"},
	}); err != nil {
		t.Fatal(err)
	}

	writeCSV(t, dir, id, source.String())
	srv := newTestServer(t, dir)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+id+"/download?format=business", http.NoBody)
	recorder := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Header().Get("Content-Disposition"), "-business.csv") {
		t.Fatalf("download failed: %d %s", recorder.Code, recorder.Body.String())
	}

	rows, err := csv.NewReader(recorder.Body).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	want := [][]string{
		{"Business name", "Website address", "Physical address", "Phone number(s)", "Hours"},
		{"Cafe, Test", "https://cafe.test", "1 Main St", "(303) 555-0100; 3035550101", "Monday: 9 AM–12 PM, 1 PM–5 PM\nSunday: Closed"},
		{"Missing details", "", "", "", ""},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("unexpected business CSV: %#v", rows)
	}

	request = httptest.NewRequest(http.MethodGet, "/download?id="+id, http.NoBody)
	recorder = httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(recorder, request)

	if recorder.Body.String() != source.String() {
		t.Fatal("default full CSV download changed")
	}
}

func TestLegacyBusinessDetails(t *testing.T) {
	places, err := parsePlaces(strings.NewReader("title,phone,open_hours\nLegacy,555,invalid\n"))
	if err != nil {
		t.Fatal(err)
	}

	if len(places) != 1 || !reflect.DeepEqual(places[0].PhoneNumbers, []string{"555"}) || places[0].Hours != "" || places[0].HasCoordinates {
		t.Fatalf("unexpected legacy details: %+v", places)
	}
}
