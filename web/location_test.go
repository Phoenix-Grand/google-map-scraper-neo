//nolint:testpackage // tests internal location preparation and injected lookup providers.
package web

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestZIPResolver(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		body      string
		wantError error
	}{
		{name: "valid", status: http.StatusOK, body: `{"places":[{"place name":"Denver","state abbreviation":"CO","latitude":"39.7491","longitude":"-104.9946"}]}`},
		{name: "not found", status: http.StatusNotFound, wantError: errZIPNotFound},
		{name: "provider down", status: http.StatusBadGateway, wantError: errZIPUnavailable},
		{name: "invalid JSON", status: http.StatusOK, body: `<html>error</html>`, wantError: errZIPUnavailable},
		{name: "no places", status: http.StatusOK, body: `{"places":[]}`, wantError: errZIPUnavailable},
		{name: "invalid coordinates", status: http.StatusOK, body: `{"places":[{"latitude":"NaN","longitude":"200"}]}`, wantError: errZIPUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/us/80202" {
					t.Errorf("unexpected lookup path: %s", r.URL.Path)
				}

				w.WriteHeader(test.status)
				_, _ = fmt.Fprint(w, test.body)
			}))
			defer provider.Close()

			resolver := &zipCodeResolver{client: provider.Client(), baseURL: provider.URL}

			location, err := resolver.Resolve(context.Background(), "80202")
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}

			if err == nil && (location.latitude != "39.7491" || location.longitude != "-104.9946" || location.name != "Denver, CO") {
				t.Fatalf("unexpected location: %+v", location)
			}
		})
	}
}

func TestZIPResolverHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resolver := &zipCodeResolver{client: &http.Client{Timeout: time.Second}, baseURL: "http://127.0.0.1:1"}

	_, err := resolver.Resolve(ctx, "80202")
	if !errors.Is(err, context.Canceled) || !errors.Is(err, errZIPUnavailable) {
		t.Fatalf("expected canceled lookup, got %v", err)
	}
}

func TestNormalizeZIPCode(t *testing.T) {
	for _, test := range []struct{ input, want string }{
		{"80202", "80202"}, {" 02108-1234 ", "02108"}, {"00123", "00123"},
		{"8020", ""}, {"80202x", ""}, {"../80202", ""}, {"80202-123", ""},
	} {
		actual, err := normalizeZIPCode(test.input)
		if actual != test.want || (err != nil) != (test.want == "") {
			t.Errorf("normalize %q: got %q, %v", test.input, actual, err)
		}
	}
}

type fixedZIPResolver struct {
	zip string
	err error
}

func (f *fixedZIPResolver) Resolve(_ context.Context, zip string) (zipLocation, error) {
	f.zip = zip

	return zipLocation{latitude: "39.7491", longitude: "-104.9946", name: "Denver, CO"}, f.err
}

func TestZIPJobCreationFromFormAndAPI(t *testing.T) {
	for _, api := range []bool{false, true} {
		t.Run(fmt.Sprintf("api=%t", api), func(t *testing.T) {
			repo := &mutableJobRepo{}
			srv := newTestServerWithRepo(t, repo)
			resolver := &fixedZIPResolver{}
			srv.zipResolver = resolver

			var request *http.Request
			if api {
				request = httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader(`{"name":"ZIP test","keywords":["coffee"],"lang":"en","zip_code":"80202-1234","radius_miles":2.5,"depth":1,"max_time":180,"fast_mode":true,"lat":"0","lon":"0"}`))
			} else {
				form := url.Values{"name": {"ZIP test"}, "keywords": {"coffee"}, "lang": {"en"}, "zoom": {"15"}, "zip_code": {"80202-1234"}, "radius_miles": {"2.5"}, "depth": {"1"}, "maxtime": {"3m"}, "fastmode": {"on"}, "latitude": {"0"}, "longitude": {"0"}}
				request = httptest.NewRequest(http.MethodPost, "/scrape", strings.NewReader(form.Encode()))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			recorder := httptest.NewRecorder()
			srv.srv.Handler.ServeHTTP(recorder, request)

			wantStatus := http.StatusOK
			if api {
				wantStatus = http.StatusCreated
			}

			if recorder.Code != wantStatus || len(repo.jobs) != 1 {
				t.Fatalf("create failed: %d %s", recorder.Code, recorder.Body.String())
			}

			data := repo.jobs[0].Data
			if resolver.zip != "80202" || data.ZIPCode != "80202" || data.Lat != "39.7491" || data.Lon != "-104.9946" || data.LocationName != "Denver, CO" || data.RadiusMiles != 2.5 {
				t.Fatalf("incorrect saved location: %+v", data)
			}

			if math.Abs(data.RadiusMeters()-4023.36) > 0.00001 {
				t.Fatalf("incorrect radius conversion: %f", data.RadiusMeters())
			}
		})
	}
}

func TestLocationErrorsDoNotSaveJobs(t *testing.T) {
	for _, test := range []struct {
		name        string
		zip         string
		miles       string
		lookupError error
		status      int
	}{
		{name: "invalid ZIP", zip: "8020", miles: "1", status: http.StatusUnprocessableEntity},
		{name: "negative miles", zip: "80202", miles: "-1", status: http.StatusUnprocessableEntity},
		{name: "nonfinite miles", zip: "80202", miles: "NaN", status: http.StatusUnprocessableEntity},
		{name: "no center", miles: "1", status: http.StatusUnprocessableEntity},
		{name: "unknown ZIP", zip: "00000", miles: "1", lookupError: errZIPNotFound, status: http.StatusUnprocessableEntity},
		{name: "lookup unavailable", zip: "80202", miles: "1", lookupError: errZIPUnavailable, status: http.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &mutableJobRepo{}
			srv := newTestServerWithRepo(t, repo)
			srv.zipResolver = &fixedZIPResolver{err: test.lookupError}
			form := url.Values{"name": {"ZIP test"}, "keywords": {"coffee"}, "lang": {"en"}, "zoom": {"15"}, "zip_code": {test.zip}, "radius_miles": {test.miles}, "depth": {"1"}, "maxtime": {"3m"}}
			request := httptest.NewRequest(http.MethodPost, "/scrape", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			recorder := httptest.NewRecorder()
			srv.srv.Handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status || len(repo.jobs) != 0 {
				t.Fatalf("unexpected create response: %d %s, saved %d jobs", recorder.Code, recorder.Body.String(), len(repo.jobs))
			}
		})
	}
}
