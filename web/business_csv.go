package web

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func writeBusinessCSV(output io.Writer, places []Place) error {
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"Business name", "Website address", "Physical address", "Phone number(s)", "Hours"}); err != nil {
		return err
	}

	for i := range places {
		place := &places[i]
		if err := writer.Write([]string{place.Title, place.Website, place.Address, strings.Join(place.PhoneNumbers, "; "), place.Hours}); err != nil {
			return err
		}
	}

	writer.Flush()

	return writer.Error()
}

func (s *Server) downloadBusiness(w http.ResponseWriter, r *http.Request, id string) {
	places, err := s.svc.GetPlaces(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrPlacesNotFound) {
			status = http.StatusNotFound
		}

		http.Error(w, "Unable to read business results", status)

		return
	}

	var output bytes.Buffer
	if err := writeBusinessCSV(&output, places); err != nil {
		http.Error(w, "Unable to prepare business CSV", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-business.csv", id))
	_, _ = output.WriteTo(w)
}
