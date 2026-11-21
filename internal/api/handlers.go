package api

import (
	"encoding/json"
	"net/http"
	"project/internal/models"
	"strconv"
	"time"
)

const layout = "2006-01-02"

type CreateBookingResponse struct {
	ID int64 `json:"id"`
}

func (s *Api) CreateBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.Booking
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный json")
		return
	}

	if req.RoomID <= 0 || req.StartAt == "" || req.EndAt == "" {
		writeError(w, http.StatusBadRequest, "пустые значения")
		return
	}

	start, err := time.Parse(layout, req.StartAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "не тот формат")
		return
	}

	end, err := time.Parse(layout, req.EndAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "не тот формат")
		return
	}

	if !start.Before(end) {
		writeError(w, http.StatusBadRequest, "start_at должен быть раньше end_at")
		return
	}

	id, err := s.DB.CreateBooking(ctx, req.RoomID, start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка при создании брони")
		return
	}

	writeJson(w, CreateBookingResponse{ID: id})
}

func (s *Api) DeleteBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "неверный id")
		return
	}

	n, err := s.DB.DeleteBooking(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удаляется бронь")
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "бронь не найдена")
		return
	}
	writeJson(w, map[string]int{"id": id})
}

func (s *Api) GetBookings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID, err := strconv.Atoi(r.URL.Query().Get("room_id"))
	if err != nil || roomID <= 0 {
		writeError(w, http.StatusBadRequest, "неверный room_id")
		return
	}

	bookings, err := s.DB.GetBookings(ctx, roomID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка извлечения брони")
		return
	}

	writeJson(w, bookings)

}
