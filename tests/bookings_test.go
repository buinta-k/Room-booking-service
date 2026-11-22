package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"project/internal/api"
	"project/internal/models"

	"github.com/stretchr/testify/assert"
)

const (
	startDate = "2026-12-01"
	endDate   = "2026-12-05"
)

// Создаёт бронь и возвращает её ID.
func createBookingForTest(t *testing.T, s *api.Api, conn *sql.DB) int {
	t.Helper()

	roomID := insertRoom(t, conn, models.Room{Name: "Room 1"})

	body, err := json.Marshal(models.Booking{
		RoomID:  int(roomID),
		StartAt: startDate,
		EndAt:   endDate,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	s.CreateBooking(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.Booking
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, response.ID, 0)

	return response.ID
}

func TestCreateBooking(t *testing.T) {
	s, conn := newTestApi(t)

	roomID := insertRoom(t, conn, models.Room{Name: "Room 1"})

	body, err := json.Marshal(models.Booking{
		RoomID:  int(roomID),
		StartAt: startDate,
		EndAt:   endDate,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	s.CreateBooking(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var got models.Booking
	err = json.Unmarshal(rr.Body.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, got.ID, 0)
}

func TestCreateBookingInvalid(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid JSON", body: `{"room_id":`},
		{name: "empty values", body: `{"room_id":0,"start_at":"","end_at":""}`},
		{name: "invalid start date", body: `{"room_id":1,"start_at":"01-12-2026","end_at":"2026-12-05"}`},
		{name: "invalid end date", body: `{"room_id":1,"start_at":"2026-12-01","end_at":"05-12-2026"}`},
		{name: "start after end", body: `{"room_id":1,"start_at":"2026-12-05","end_at":"2026-12-01"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := newTestApi(t)

			req := httptest.NewRequest(
				http.MethodPost,
				"/bookings",
				bytes.NewBufferString(tt.body),
			)
			rr := httptest.NewRecorder()

			s.CreateBooking(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestGetBookings(t *testing.T) {
	s, conn := newTestApi(t)

	roomID := insertRoom(t, conn, models.Room{Name: "Room 1"})

	booking := models.Booking{
		RoomID:  int(roomID),
		StartAt: startDate,
		EndAt:   endDate,
	}

	body, err := json.Marshal(booking)
	if err != nil {
		t.Fatal(err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body))
	createRR := httptest.NewRecorder()
	s.CreateBooking(createRR, createReq)

	assert.Equal(t, http.StatusOK, createRR.Code)

	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings?room_id="+strconv.FormatInt(roomID, 10),
		nil,
	)
	rr := httptest.NewRecorder()

	s.GetBookings(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var got []models.Booking
	err = json.Unmarshal(rr.Body.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, got, 1)
	assert.Equal(t, int(roomID), got[0].RoomID)
	assert.Equal(t, startDate, got[0].StartAt)
	assert.Equal(t, endDate, got[0].EndAt)
}

func TestGetBookingsInvalidRoomID(t *testing.T) {
	s, _ := newTestApi(t)

	req := httptest.NewRequest(http.MethodGet, "/bookings?room_id=abc", nil)
	rr := httptest.NewRecorder()

	s.GetBookings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteBooking(t *testing.T) {
	s, conn := newTestApi(t)

	id := createBookingForTest(t, s, conn)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/bookings/"+strconv.Itoa(id),
		nil,
	)
	req.SetPathValue("id", strconv.Itoa(id))

	rr := httptest.NewRecorder()
	s.DeleteBooking(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var got map[string]int
	err := json.Unmarshal(rr.Body.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, id, got["id"])
}

func TestDeleteBookingNotFound(t *testing.T) {
	s, _ := newTestApi(t)

	req := httptest.NewRequest(http.MethodDelete, "/bookings/999", nil)
	req.SetPathValue("id", "999")

	rr := httptest.NewRecorder()
	s.DeleteBooking(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
