package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

type Todo struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Tags        []string  `json:"tags"`
	Description string    `json:"description"`
	IsPriority  bool      `json:"priority"`
	IsCompleted bool      `json:"completed"`
}

const TodoFile = ".todos.json"

func Load() ([]Todo, error) {
	jsonFile, err := os.Open(TodoFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var todos []Todo
	json.Unmarshal(bytes, &todos)

	return todos, nil
}

func Save(todos []Todo) error {
	bytes, err := json.Marshal(todos)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(TodoFile, bytes, 0644)
}

func HasTag(todo Todo, tags []string) bool {
	for _, todoTag := range todo.Tags {
		for _, filterTag := range tags {
			if todoTag == filterTag {
				return true
			}
		}
	}

	return false
}

func main() {
	app := &cli.App{
		Name:                   "atgore",
		Usage:                  "handle todos",
		Version:                "v0.1.0",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a task to the list",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "description", Aliases: []string{"d"}},
					&cli.BoolFlag{Name: "priority", Aliases: []string{"p"}},
					&cli.StringSliceFlag{Name: "tags", Aliases: []string{"t"}},
				},
				Action: func(ctx *cli.Context) error {
					todos, err := Load()
					if err != nil {
						return err
					}

					newTodo := Todo{
						uuid.New(),
						"name",
						ctx.StringSlice("tags"),
						"description",
						ctx.Bool("priority"),
						false,
					}

					todos = append(todos, newTodo)

					err = Save(todos)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "remove a task from the list",
				Action: func(ctx *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "complete",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Action: func(ctx *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list tasks",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "tags", Aliases: []string{"t"}},
				},
				Action: func(ctx *cli.Context) error {
					todos, err := Load()
					if err != nil {
						return err
					}

					n := len(todos)
					if tags := ctx.StringSlice("tags"); tags != nil {
						n = 0

						for _, todo := range todos {
							if HasTag(todo, tags) {
								todos[n] = todo
								n++
							}
						}
					}

					fmt.Println(todos[:n])

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
