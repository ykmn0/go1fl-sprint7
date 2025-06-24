package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, len(cafeList["moscow"])},
	}

	for _, tt := range requests {
		url := "/cafe?city=moscow"
		if tt.count >= 0 {
			url += "&count=" + strconv.Itoa(tt.count)
		}
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		if body == "" {
			assert.Equal(t, 0, tt.want, "expected zero cafes")
			continue
		}

		cafes := strings.Split(body, ",")
		for i := range cafes {
			cafes[i] = strings.TrimSpace(cafes[i])
		}
		assert.Equal(t, tt.want, len(cafes), "count mismatch for count=%d", tt.count)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, tt := range requests {
		url := "/cafe?city=moscow&search=" + tt.search
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		if body == "" {
			assert.Equal(t, 0, tt.wantCount, "expected zero cafes for search=%q", tt.search)
			continue
		}

		cafes := strings.Split(body, ",")
		for i := range cafes {
			cafes[i] = strings.TrimSpace(cafes[i])
			assert.True(t, strings.Contains(strings.ToLower(cafes[i]), strings.ToLower(tt.search)),
				"cafe %q does not contain search %q", cafes[i], tt.search)
		}

		assert.Equal(t, tt.wantCount, len(cafes), "count mismatch for search=%q", tt.search)
	}
}
