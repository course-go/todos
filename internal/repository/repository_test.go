package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/course-go/todos/internal/repository"
	"github.com/course-go/todos/internal/todos"
	"github.com/course-go/todos/internal/utils/test"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	c := test.NewTestContainer(ctx, t)
	t.Cleanup(func() {
		err := c.Terminate(ctx)
		if err != nil {
			t.Logf("failed terminating postgres container: %v", err)
		}
	})
	err := c.Snapshot(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatalf("failed creating database snapshot: %v", err)
	}

	logger := test.NewTestLogger(t)
	t.Run("Create todo", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		todo := todos.Todo{
			Description: "Mop the floor",
		}
		createdTodo, err := r.CreateTodo(ctx, todo)
		if err != nil {
			t.Fatalf("could not create todo: %v", err)
		}

		retrievedTodo, err := r.GetTodo(ctx, createdTodo.ID)
		if err != nil {
			t.Fatalf("could not retrieve created todo: %v", err)
		}

		if todo.Description != retrievedTodo.Description {
			t.Fatalf("todo descriptions do not match: expected: %s != actual: %s",
				todo.Description,
				retrievedTodo.Description,
			)
		}
	})
	t.Run("Get existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		id, err := uuid.Parse("f52bad23-c201-414e-9bdb-af4327c42aa7")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		todo, err := r.GetTodo(ctx, id)
		if err != nil {
			t.Fatalf("could not get todo: %v", err)
		}

		expectedDescription := "Vacuum"
		if todo.Description != expectedDescription {
			t.Fatalf("todo descriptions do not match: expected: %s != actual: %s",
				expectedDescription,
				todo.Description,
			)
		}
	})
	t.Run("Get non-existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		id, err := uuid.Parse("be95c29a-c4dd-4d31-a5c4-d229f3374ab7")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		_, err = r.GetTodo(ctx, id)
		if !errors.Is(err, repository.ErrTodoNotFound) {
			t.Fatalf("todo should not be found: expected: %v != actual: %v", repository.ErrTodoNotFound, err)
		}
	})
	t.Run("Get todos", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		todos, err := r.GetTodos(ctx)
		if err != nil {
			t.Fatalf("could not get todos: %v", err)
		}

		expectedTodosLen := 2
		if len(todos) != expectedTodosLen {
			t.Fatalf("todos length does not match: expected: %d != actual: %d",
				expectedTodosLen,
				len(todos),
			)
		}
	})
	t.Run("Save existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		id, err := uuid.Parse("62446c85-3798-471f-abb8-75c1cdd7153b")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		todo, err := r.GetTodo(ctx, id)
		if err != nil {
			t.Fatalf("could not get todo: %v", err)
		}

		now := time.Now()
		todo.CompletedAt = &now
		savedTodo, err := r.SaveTodo(ctx, todo)
		if err != nil {
			t.Fatalf("could not save todo: %v", err)
		}

		if savedTodo.UpdatedAt == nil {
			t.Fatalf("todo updated timestampt was not changed")
		}

		nowRounded := now.Round(time.Millisecond)
		todoRounded := savedTodo.CompletedAt.Round(time.Millisecond)
		if !todoRounded.Equal(nowRounded) {
			t.Fatalf("todo completed timestampt does not match: expected: %s != actual: %s",
				now.String(),
				savedTodo.UpdatedAt.String(),
			)
		}
	})
	t.Run("Save non-existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		id, err := uuid.Parse("ac4011ce-59c9-4361-8abf-10abd273d5e5")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		todo := todos.Todo{
			ID:          id,
			Description: "Do some shopping",
			CompletedAt: nil,
		}
		_, err = r.SaveTodo(ctx, todo)
		if !errors.Is(err, repository.ErrTodoNotFound) {
			t.Fatalf("todo should not be found: expected: %v != actual: %v", repository.ErrTodoNotFound, err)
		}
	})
	t.Run("Delete existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		id, err := uuid.Parse("f52bad23-c201-414e-9bdb-af4327c42aa7")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		err = r.DeleteTodo(ctx, id)
		if err != nil {
			t.Fatalf("todo should be deleted: expected: nil != actual: %v", err)
		}
	})
	t.Run("Delete non-existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			test.RestoreDatabase(ctx, t, c)
		})

		r := test.NewTestRepository(ctx, t, logger, c)
		id, err := uuid.Parse("4fabcaa9-7fe6-4129-86f2-1d62d142a67b")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		err = r.DeleteTodo(ctx, id)
		if !errors.Is(err, repository.ErrTodoNotFound) {
			t.Fatalf("todo should not be found: expected: %v != actual: %v", repository.ErrTodoNotFound, err)
		}
	})
}
