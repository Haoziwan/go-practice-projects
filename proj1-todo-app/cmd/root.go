package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"tasks/internal/store"

	"github.com/mergestat/timediff"
)

// Run starts the interactive CLI
func Run() {
	fmt.Println("Tasks - Interactive Task Manager")
	fmt.Println("Type 'help' for available commands, 'quit' to exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("tasks> ")

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := parseArgs(line)
		if len(args) == 0 {
			continue
		}

		cmd := strings.ToLower(args[0])

		switch cmd {
		case "help", "h":
			printHelp()
		case "add", "a":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Error: missing task description")
				fmt.Println("Usage: add <description>")
				continue
			}
			addTask(strings.Join(args[1:], " "))
		case "list", "ls", "l":
			showAll := false
			if len(args) > 1 && (args[1] == "-a" || args[1] == "--all") {
				showAll = true
			}
			listTasks(showAll)
		case "complete", "done", "c":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Error: missing task ID")
				fmt.Println("Usage: complete <taskid>")
				continue
			}
			id, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error: invalid task ID")
				continue
			}
			completeTask(id)
		case "delete", "del", "d":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Error: missing task ID")
				fmt.Println("Usage: delete <taskid>")
				continue
			}
			id, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error: invalid task ID")
				continue
			}
			deleteTask(id)
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
		}
	}
}

// parseArgs splits a line into arguments, respecting quoted strings
func parseArgs(line string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range line {
		if !inQuote {
			if r == '"' || r == '\'' {
				inQuote = true
				quoteChar = r
			} else if r == ' ' || r == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		} else {
			if r == quoteChar {
				inQuote = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "  add <description>\tAdd a new task")
	fmt.Fprintln(w, "  list [-a]\tList tasks (-a to show all including completed)")
	fmt.Fprintln(w, "  complete <id>\tMark a task as completed")
	fmt.Fprintln(w, "  delete <id>\tDelete a task")
	fmt.Fprintln(w, "  help\tShow this help message")
	fmt.Fprintln(w, "  quit\tExit the application")
	w.Flush()
	fmt.Println()
	fmt.Println("Shortcuts: a=add, l/ls=list, c/done=complete, d/del=delete, h=help, q=quit")
}

func addTask(description string) {
	s, err := store.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}
	defer s.Close()

	task := s.Add(description)

	if err := s.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	fmt.Printf("Added task %d: %s\n", task.ID, task.Description)
}

func listTasks(showAll bool) {
	s, err := store.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}
	defer s.Close()

	tasks := s.List(showAll)

	if len(tasks) == 0 {
		if showAll {
			fmt.Println("No tasks found.")
		} else {
			fmt.Println("No uncompleted tasks found. Use 'list -a' to show all tasks.")
		}
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	if showAll {
		fmt.Fprintln(w, "ID\tTask\tCreated\tDone")
		for _, t := range tasks {
			done := "false"
			if t.IsComplete() {
				done = "true"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
				t.ID,
				t.Description,
				timediff.TimeDiff(t.CreatedAt),
				done,
			)
		}
	} else {
		fmt.Fprintln(w, "ID\tTask\tCreated")
		for _, t := range tasks {
			fmt.Fprintf(w, "%d\t%s\t%s\n",
				t.ID,
				t.Description,
				timediff.TimeDiff(t.CreatedAt),
			)
		}
	}

	w.Flush()
}

func completeTask(id int) {
	s, err := store.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}
	defer s.Close()

	task, err := s.GetByID(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Complete(id); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	fmt.Printf("Completed task %d: %s\n", id, task.Description)
}

func deleteTask(id int) {
	s, err := store.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}
	defer s.Close()

	task, err := s.GetByID(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Delete(id); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	if err := s.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	fmt.Printf("Deleted task %d: %s\n", id, task.Description)
}
