package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlePostJson(t *testing.T) {
	tests := map[string]struct {
		data         string
		wantRespCode int
	}{
		"empty body": {
			data:         "",
			wantRespCode: http.StatusBadRequest,
		},
		"bad json": {
			data:         `asc0124`,
			wantRespCode: http.StatusBadRequest,
		},
		"simple json": {
			data:         `{"foo": 123}`,
			wantRespCode: http.StatusOK,
		},
		"complex json": {
			data: `{
					"id": 54321,
					"name": "Alice Smith",
					"email": "alice@example.com",
					"address": {
						"city": "Wonderland",
						"zipcode": "12345"
					},
					"preferences": {
						"theme": "light",
						"notifications": {
						"email": true
						}
					},
					"orders": [
						{
						"order_id": "ORD123",
						"total": 59.99,
						"items": [
							{
							"product_id": "PROD100",
							"name": "Book",
							"quantity": 1
							}
						]
						}
					]
					}`,
			wantRespCode: http.StatusOK,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(test.data))
			res := httptest.NewRecorder()

			handler := New(&Config{Addr: ":8080", Id: "g4rble"})
			handler.handlePostJson(res, req)

			if res.Code != test.wantRespCode {
				t.Fatalf("got status %d but wanted %d", res.Code, test.wantRespCode)
			}

			if res.Code != http.StatusOK {
				// can't test body in this case
				return
			}

			responseBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("response body differs: got %s want %s", string(responseBody), test.data)
			}
			if test.data != string(responseBody) {

			}
		})
	}
}
