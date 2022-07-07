package main

type Todo struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	IsPriority  bool     `json:"priority"`
	IsCompleted bool     `json:"completed"`
}
