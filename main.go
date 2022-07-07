package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/urfave/cli/v2"
)

const TodoFile = ".todos"
const KeyFile = ".key"

func main() {
	app := &cli.App{
		Name:                   "atgore",
		Usage:                  "handle todos",
		Version:                "v0.2.1",
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
					args := ctx.Args()
					if !args.Present() {
						return cli.Exit("Missing arguments", 1)
					}

					todoList := &TodoList{}

					err := todoList.Load()
					if err != nil {
						return err
					}

					newTodo := &Todo{
						Name:        strings.Join(args.Slice(), " "),
						Tags:        ctx.StringSlice("tags"),
						Description: ctx.String("description"),
						IsPriority:  ctx.Bool("priority"),
						IsCompleted: false,
					}

					todoList.Add(newTodo)

					err = todoList.Save()
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
					args := ctx.Args()
					if !args.Present() {
						return cli.Exit("Missing arguments", 1)
					}

					todoList := &TodoList{}

					err := todoList.Load()
					if err != nil {
						return err
					}

					id, err := strconv.ParseInt(ctx.Args().First(), 10, 32)
					if err != nil {
						return err
					}

					todoList.Remove(int(id))
					todoList.Save()

					return nil
				},
			},
			{
				Name:    "complete",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Action: func(ctx *cli.Context) error {
					args := ctx.Args()
					if !args.Present() {
						return cli.Exit("Missing arguments", 1)
					}

					todoList := &TodoList{}

					err := todoList.Load()
					if err != nil {
						return err
					}

					id, err := strconv.ParseInt(ctx.Args().First(), 10, 32)
					if err != nil {
						return err
					}

					todoList.Complete(int(id))
					todoList.Save()

					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list tasks",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "tag", Aliases: []string{"t"}},
					&cli.BoolFlag{Name: "priority", Aliases: []string{"p"}},
					&cli.BoolFlag{Name: "completed", Aliases: []string{"c"}},
				},
				Action: func(ctx *cli.Context) error {
					todoList := &TodoList{}
					err := todoList.Load()
					if err != nil {
						return nil
					}

					todos := todoList.Todos
					if tag := ctx.String("tag"); tag != "" {
						todos = todoList.FindByTag(tag)
					}

					if ctx.Bool("priority") {
						todos = todoList.FindPriority()
					}
					if ctx.Bool("completed") {
						todos = todoList.FindCompleted()
					}

					PrintTodos(todos)

					return nil
				},
			},
			{
				Name: "genkey",
				Action: func(ctx *cli.Context) error {
					args := ctx.Args()
					if !args.Present() {
						return cli.Exit("Missing arguments", 1)
					}

					email := args.First()

					password := prompter.Password("Enter your password")

					masterKey := genMasterKey(password, email)
					hkdfKey, hkdfMacKey := strechMasterKey(masterKey)

					protectedSymKey, err := genProtectedSymKey(hkdfKey, hkdfMacKey)
					if err != nil {
						return err
					}

					err = os.WriteFile(".key", protectedSymKey, 0644)
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
