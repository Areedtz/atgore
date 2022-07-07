package main

import (
	"strings"

	"github.com/cheynewallace/tabby"
)

func PrintTodos(todos []Todo) {
	t := tabby.New()
	t.AddHeader("ID", "Name", "Tags", "Description", "Is Priority", "Is Completed")
	for _, todo := range todos {
		completed := "[ ]"
		if todo.IsCompleted {
			completed = "[x]"
		}

		priority := "[ ]"
		if todo.IsPriority {
			priority = "[x]"
		}

		t.AddLine(todo.Id, todo.Name, strings.Join(todo.Tags, ", "), todo.Description, priority, completed)
	}
	t.Print()
}
