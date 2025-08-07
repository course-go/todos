package todos_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/course-go/todos/internal/http/dto/request"
	"github.com/course-go/todos/internal/http/dto/response"
	"github.com/course-go/todos/internal/utils/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const apiURLPrefix = "/api/v1"

func TestTodosControllers(t *testing.T) { //nolint: cyclop, maintidx, tparallel
	t.Parallel()

	ctx := t.Context()
	logger := test.NewTestLogger(t)
	r := test.NewTestRouter(ctx, t, logger)

	playGamesTodoID := "62446c85-3798-471f-abb8-75c1cdd7153b"

	t.Run("Get Todos", func(t *testing.T) { //nolint: paralleltest
		req := httptest.NewRequest(http.MethodGet, apiURLPrefix+"/todos", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusOK)

		expectedBodyBytes := []byte(`{
		   "data":{
			  "todos":[
				 {
					"id":"62446c85-3798-471f-abb8-75c1cdd7153b",
					"description":"Mop the floor",
					"createdAt":"2024-07-26T22:48:21.090537Z"
				 },
				 {
					"id":"f52bad23-c201-414e-9bdb-af4327c42aa7",
					"description":"Vacuum",
					"createdAt":"2024-07-26T22:49:47.366006Z",
					"completedAt":"2024-07-27T22:50:19.594495Z",
					"updatedAt":"2024-07-27T22:50:19.594495Z"
				 }
			  ]
		   }
		}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})

	t.Run("Get existing Todo", func(t *testing.T) { //nolint: paralleltest
		req := httptest.NewRequest(http.MethodGet, apiURLPrefix+"/todos/"+playGamesTodoID, nil)
		req.SetPathValue("id", playGamesTodoID)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusOK)

		expectedBodyBytes := []byte(`{
		   "data":{
			  "todo":{
				 "id":"62446c85-3798-471f-abb8-75c1cdd7153b",
				 "description":"Mop the floor",
				 "createdAt":"2024-07-26T22:48:21.090537Z"
			  }
		   }
		}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})

	t.Run("Get non-existing Todo", func(t *testing.T) { //nolint: paralleltest
		todoID := "d8d5141a-ad8c-486a-9d4d-6bda9c7cb33c"
		req := httptest.NewRequest(http.MethodGet, apiURLPrefix+"/todos/"+todoID, nil)
		req.SetPathValue("id", todoID)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusNotFound)

		expectedBodyBytes := []byte(`{"error":"Not Found"}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})

	t.Run("Create Todo", func(t *testing.T) { //nolint: paralleltest
		body := request.CreateTodoRequest{
			Description: "Play some games",
		}

		actualBodyBytes, err := json.Marshal(&body)
		if err != nil {
			t.Fatalf("failed marshaling request body: %v", err)
		}

		reader := bytes.NewReader(actualBodyBytes)
		req := httptest.NewRequest(http.MethodPost, apiURLPrefix+"/todos", reader)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusCreated)

		expectedBodyBytes := []byte(`{
		   "data":{
			  "todo":{
				 "id":"bc931469-bb84-4fd0-aa6d-acfef864580d",
				 "description":"Play some games",
				 "createdAt":"2024-08-18T12:14:45.847679Z"
			  }
		   }
		}`)

		var expectedResponseBody response.Response

		err = json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("could not unmarshal expected body bytes: %v", err)
		}

		actualBodyBytes, err = io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("could not read body bytes: %v", err)
		}

		var actualResponseBody response.Response

		err = json.Unmarshal(actualBodyBytes, &actualResponseBody)
		if err != nil {
			t.Errorf("could not unmarshal actual body bytes: %v", err)
		}

		// The UUIDs that todos use are automatically generated so they are ignored.
		// They could be faked similarly like time, but was not worth the effort.
		ignoreIDs := cmpopts.IgnoreMapEntries(
			func(key string, _ any) bool {
				return key == "id"
			},
		)
		if !cmp.Equal(expectedResponseBody, actualResponseBody, ignoreIDs) {
			t.Errorf("expected and actual response bodies do not match: %s",
				cmp.Diff(expectedResponseBody, actualResponseBody, ignoreIDs),
			)
		}

		assertJSONContentType(t, res)
	})

	t.Run("Create Todo with invalid body", func(t *testing.T) { //nolint: paralleltest
		body := request.CreateTodoRequest{}

		actualBodyBytes, err := json.Marshal(&body)
		if err != nil {
			t.Fatalf("failed marshaling request body: %v", err)
		}

		reader := bytes.NewReader(actualBodyBytes)
		req := httptest.NewRequest(http.MethodPost, apiURLPrefix+"/todos", reader)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusBadRequest)

		expectedBodyBytes := []byte(`{"error":"Bad Request"}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})

	t.Run("Edit existing Todo", func(t *testing.T) { //nolint: paralleltest
		completedAt, err := time.Parse(time.DateTime, "2024-07-28 22:51:00")
		if err != nil {
			t.Errorf("failed parsing time: %v", err)
		}

		body := request.UpdateTodoRequest{
			Description: "Play some games",
			CompletedAt: &completedAt,
		}

		actualBodyBytes, err := json.Marshal(&body)
		if err != nil {
			t.Fatalf("failed marshaling request body: %v", err)
		}

		reader := bytes.NewReader(actualBodyBytes)
		req := httptest.NewRequest(http.MethodPut, apiURLPrefix+"/todos/"+playGamesTodoID, reader)
		req.SetPathValue("id", playGamesTodoID)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusOK)

		expectedBodyBytes := []byte(`{
		   "data":{
			  "todo":{
				 "id":"62446c85-3798-471f-abb8-75c1cdd7153b",
				 "description":"Play some games",
				 "createdAt":"2024-07-26T22:48:21.090537Z",
				 "completedAt":"2024-07-28T22:51:00Z",
				 "updatedAt":"2024-08-18T12:14:45.847679Z"
			  }
		   }
		}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})

	t.Run("Edit non-existing Todo", func(t *testing.T) { //nolint: paralleltest
		completedAt, err := time.Parse(time.DateTime, "2024-07-28 22:51:00")
		if err != nil {
			t.Errorf("failed parsing time: %v", err)
		}

		body := request.UpdateTodoRequest{
			Description: "Play some games",
			CompletedAt: &completedAt,
		}

		actualBodyBytes, err := json.Marshal(&body)
		if err != nil {
			t.Fatalf("failed marshaling request body: %v", err)
		}

		reader := bytes.NewReader(actualBodyBytes)
		todoID := "cba6b1a9-3533-4eff-8649-a075229b1c3d"
		req := httptest.NewRequest(http.MethodPut, apiURLPrefix+"/todos/"+todoID, reader)
		req.SetPathValue("id", todoID)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusNotFound)

		expectedBodyBytes := []byte(`{"error":"Not Found"}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})

	t.Run("Edit Todo with invalid body", func(t *testing.T) { //nolint: paralleltest
		completedAt, err := time.Parse(time.DateTime, "2024-07-28 22:51:00")
		if err != nil {
			t.Errorf("failed parsing time: %v", err)
		}

		body := request.UpdateTodoRequest{
			CompletedAt: &completedAt,
		}

		actualBodyBytes, err := json.Marshal(&body)
		if err != nil {
			t.Fatalf("failed marshaling request body: %v", err)
		}

		reader := bytes.NewReader(actualBodyBytes)
		req := httptest.NewRequest(http.MethodPut, apiURLPrefix+"/todos/"+playGamesTodoID, reader)
		req.SetPathValue("id", playGamesTodoID)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusBadRequest)

		expectedBodyBytes := []byte(`{"error":"Bad Request"}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})

	t.Run("Delete existing Todo", func(t *testing.T) { //nolint: paralleltest
		req := httptest.NewRequest(http.MethodDelete, apiURLPrefix+"/todos/"+playGamesTodoID, nil)
		req.SetPathValue("id", playGamesTodoID)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusNoContent)

		actualBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("could not read body bytes: %v", err)
		}

		if len(actualBody) != 0 {
			t.Errorf("expected no bytes in body but was %d", len(actualBody))
		}

		assertJSONContentType(t, res)
	})

	t.Run("Delete non-existing Todo", func(t *testing.T) { //nolint: paralleltest
		todoID := "d06c0dd1-d7ae-4ca7-8df4-86a6b62f349d"
		req := httptest.NewRequest(http.MethodDelete, apiURLPrefix+"/todos/"+todoID, nil)
		req.SetPathValue("id", todoID)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		res := rr.Result()
		compareResponseCodes(t, res, http.StatusNotFound)

		expectedBodyBytes := []byte(`{"error":"Not Found"}`)
		compareResponseBodies(t, res, expectedBodyBytes)
		assertJSONContentType(t, res)
	})
}

func compareResponseCodes(t *testing.T, res *http.Response, expectedCode int) {
	t.Helper()

	actualCode := res.StatusCode
	if expectedCode != actualCode {
		t.Errorf("expected %d content type but was %d", expectedCode, actualCode)
	}
}

func compareResponseBodies(t *testing.T, res *http.Response, expectedBody []byte) {
	t.Helper()

	actualBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("could not read body bytes: %v", err)
	}

	var expectedResponseBody response.Response

	err = json.Unmarshal(expectedBody, &expectedResponseBody)
	if err != nil {
		t.Errorf("could not unmarshal expected body bytes: %v", err)
	}

	var actualResponseBody response.Response

	err = json.Unmarshal(actualBody, &actualResponseBody)
	if err != nil {
		t.Errorf("could not unmarshal actual body bytes: %v", err)
	}

	if !cmp.Equal(expectedResponseBody, actualResponseBody) {
		t.Errorf("expected and actual response bodies do not match: %s",
			cmp.Diff(expectedResponseBody, actualResponseBody),
		)
	}
}

func assertJSONContentType(t *testing.T, res *http.Response) {
	t.Helper()

	expectedType := "application/json"
	actualType := res.Header.Get("Content-Type")

	if expectedType != actualType {
		t.Errorf("expected %s content type but was %s", expectedType, actualType)
	}
}
