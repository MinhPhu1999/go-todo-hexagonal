package platform

import (
	"errors"
	"testing"
)

func TestFetchTodosRequestNormalizeDefaults(t *testing.T) {
	request, err := (FetchTodosRequest{}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() returned error: %v", err)
	}

	if request.PageSize != defaultTodoPageSize {
		t.Fatalf("expected default page size %d, got %d", defaultTodoPageSize, request.PageSize)
	}
	if request.PageNumber != defaultTodoPage {
		t.Fatalf("expected default page number %d, got %d", defaultTodoPage, request.PageNumber)
	}
}

func TestFetchTodosRequestNormalizeRejectsLargePageSize(t *testing.T) {
	_, err := (FetchTodosRequest{PageSize: maxTodoPageSize + 1, PageNumber: 1}).Normalize()
	if !errors.Is(err, ErrInvalidTodo) {
		t.Fatalf("expected ErrInvalidTodo, got %v", err)
	}
}

func TestTodoListTotalPages(t *testing.T) {
	list := TodoList{
		PageSize: 10,
		Total:    21,
	}

	if totalPages := list.TotalPages(); totalPages != 3 {
		t.Fatalf("expected 3 total pages, got %d", totalPages)
	}
}
