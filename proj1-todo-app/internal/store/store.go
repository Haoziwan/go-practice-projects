package store

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"tasks/internal/task"
)

const (
	defaultFilename = "tasks.csv"
	timeFormat      = time.RFC3339
)

// Store manages the task data file
type Store struct {
	filepath string
	file     *os.File
	tasks    []task.Task
}

// New creates a new Store instance
func New() (*Store, error) {
	// Use current working directory for data file
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	fp := filepath.Join(cwd, defaultFilename)

	return &Store{
		filepath: fp,
		tasks:    []task.Task{},
	}, nil
}

// Open opens the data file and loads tasks
func (s *Store) Open() error {
	// Open or create the data file
	f, err := os.OpenFile(s.filepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open file for reading: %w", err)
	}
	s.file = f

	if err := s.loadTasks(); err != nil {
		s.Close()
		return err
	}

	return nil
}

// Close closes the file
func (s *Store) Close() error {
	if s.file != nil {
		err := s.file.Close()
		s.file = nil
		return err
	}
	return nil
}

// loadTasks reads all tasks from the CSV file
func (s *Store) loadTasks() error {
	s.tasks = []task.Task{}

	// Seek to beginning of file
	if _, err := s.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	reader := csv.NewReader(s.file)

	// Read header
	header, err := reader.Read()
	if err == io.EOF {
		// Empty file, no tasks
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	expectedHeader := []string{"ID", "Description", "CreatedAt", "CompletedAt"}
	if len(header) < 4 {
		return fmt.Errorf("invalid CSV header: expected %v", expectedHeader)
	}

	// Read records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV record: %w", err)
		}

		t, err := s.parseTask(record)
		if err != nil {
			return fmt.Errorf("failed to parse task: %w", err)
		}
		s.tasks = append(s.tasks, t)
	}

	return nil
}

// parseTask parses a CSV record into a Task
func (s *Store) parseTask(record []string) (task.Task, error) {
	if len(record) < 4 {
		return task.Task{}, fmt.Errorf("invalid record: expected 4 fields, got %d", len(record))
	}

	id, err := strconv.Atoi(record[0])
	if err != nil {
		return task.Task{}, fmt.Errorf("invalid ID: %w", err)
	}

	createdAt, err := time.Parse(timeFormat, record[2])
	if err != nil {
		return task.Task{}, fmt.Errorf("invalid CreatedAt: %w", err)
	}

	var completedAt *time.Time
	if record[3] != "" {
		t, err := time.Parse(timeFormat, record[3])
		if err != nil {
			return task.Task{}, fmt.Errorf("invalid CompletedAt: %w", err)
		}
		completedAt = &t
	}

	return task.Task{
		ID:          id,
		Description: record[1],
		CreatedAt:   createdAt,
		CompletedAt: completedAt,
	}, nil
}

// Save writes all tasks to the CSV file
func (s *Store) Save() error {
	if s.file == nil {
		return fmt.Errorf("file not opened")
	}

	// Truncate and seek to beginning
	if err := s.file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	if _, err := s.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	writer := csv.NewWriter(s.file)

	// Write header
	if err := writer.Write([]string{"ID", "Description", "CreatedAt", "CompletedAt"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write records
	for _, t := range s.tasks {
		completedAt := ""
		if t.CompletedAt != nil {
			completedAt = t.CompletedAt.Format(timeFormat)
		}

		record := []string{
			strconv.Itoa(t.ID),
			t.Description,
			t.CreatedAt.Format(timeFormat),
			completedAt,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return nil
}

// Add creates a new task with the given description
func (s *Store) Add(description string) task.Task {
	// Find next ID
	maxID := 0
	for _, t := range s.tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}

	newTask := task.Task{
		ID:          maxID + 1,
		Description: description,
		CreatedAt:   time.Now(),
		CompletedAt: nil,
	}

	s.tasks = append(s.tasks, newTask)
	return newTask
}

// List returns all tasks, optionally filtering by completion status
func (s *Store) List(showAll bool) []task.Task {
	if showAll {
		return s.tasks
	}

	// Filter incomplete tasks
	var result []task.Task
	for _, t := range s.tasks {
		if !t.IsComplete() {
			result = append(result, t)
		}
	}
	return result
}

// Complete marks a task as completed by ID
func (s *Store) Complete(id int) error {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			if s.tasks[i].IsComplete() {
				return fmt.Errorf("task %d is already completed", id)
			}
			s.tasks[i].Complete()
			return nil
		}
	}
	return fmt.Errorf("task %d not found", id)
}

// Delete removes a task by ID
func (s *Store) Delete(id int) error {
	for i, t := range s.tasks {
		if t.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task %d not found", id)
}

// GetByID returns a task by its ID
func (s *Store) GetByID(id int) (*task.Task, error) {
	for _, t := range s.tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task %d not found", id)
}
