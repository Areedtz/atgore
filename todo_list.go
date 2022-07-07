package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
)

type TodoList struct {
	Todos []Todo
}

func (t *TodoList) Load() error {
	encBytes, err := os.ReadFile(TodoFile)
	if err != nil {
		return err
	}

	var todos []Todo
	if len(encBytes) > 1 {
		masterKey := genMasterKey("example", "example@example.com")
		hkdfKey, hkdfMacKey := strechMasterKey(masterKey)

		protectedSymKey, err := os.ReadFile(".key")
		if err != nil {
			return err
		}

		symkey, err := decryptProtectedSymKey(hkdfKey, hkdfMacKey, protectedSymKey)
		if err != nil {
			return err
		}

		block, err := aes.NewCipher(symkey[:32])
		if err != nil {
			return err
		}

		iv := encBytes[:aes.BlockSize]
		bytes := make([]byte, len(encBytes)-len(iv))
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(bytes, encBytes[aes.BlockSize:])

		bytes, err = unpadPKCS7(bytes, aes.BlockSize)
		if err != nil {
			return err
		}

		json.Unmarshal(bytes, &todos)
	} else {
		todos = make([]Todo, 0)
	}

	t.Todos = todos

	return nil
}

func (t *TodoList) Save() error {
	sort.Slice(t.Todos, func(i, j int) bool {
		return t.Todos[i].Id < t.Todos[j].Id
	})

	bytes, err := json.Marshal(t.Todos)
	if err != nil {
		return err
	}

	masterKey := genMasterKey("example", "example@example.com")
	hkdfKey, hkdfMacKey := strechMasterKey(masterKey)

	protectedSymKey, err := os.ReadFile(".key")
	if err != nil {
		return err
	}

	symkey, err := decryptProtectedSymKey(hkdfKey, hkdfMacKey, protectedSymKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(symkey[:32])
	if err != nil {
		return err
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return err
	}

	bytes, err = padPKCS7(bytes, aes.BlockSize)
	if err != nil {
		return err
	}

	encBytes := make([]byte, len(bytes))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encBytes, bytes)

	return ioutil.WriteFile(TodoFile, append(iv, encBytes...), 0644)
}

func (t *TodoList) MaxId() int {
	id := 0
	for _, todo := range t.Todos {
		if todo.Id > id {
			id = todo.Id
		}
	}

	return id
}

func (t *TodoList) Add(todo *Todo) {
	todo.Id = t.MaxId() + 1

	t.Todos = append(t.Todos, *todo)
}

func (t *TodoList) Find(id int) *Todo {
	for _, todo := range t.Todos {
		if todo.Id == id {
			return &todo
		}
	}

	return nil
}

func (t *TodoList) FindByTag(tag string) []Todo {
	todos := make([]Todo, len(t.Todos))
	copy(todos, t.Todos)
	n := 0

	for _, todo := range todos {
		for _, todoTag := range todo.Tags {
			if todoTag == tag {
				todos[n] = todo
				n++
			}
		}
	}

	return todos[:n]
}

func (t *TodoList) FindCompleted() []Todo {
	todos := make([]Todo, len(t.Todos))
	copy(todos, t.Todos)
	n := 0

	for _, todo := range todos {
		if todo.IsCompleted {
			todos[n] = todo
			n++
		}
	}

	return todos[:n]
}

func (t *TodoList) FindPriority() []Todo {
	todos := make([]Todo, len(t.Todos))
	copy(todos, t.Todos)
	n := 0

	for _, todo := range todos {
		if todo.IsCompleted {
			todos[n] = todo
			n++
		}
	}

	return todos[:n]
}

func (t *TodoList) Remove(id int) {
	i := -1
	for index, todo := range t.Todos {
		if todo.Id == id {
			i = index
		}
	}

	if i >= 0 {
		t.Todos[i] = t.Todos[len(t.Todos)-1]
		t.Todos = t.Todos[:len(t.Todos)-1]
	}
}

func (t *TodoList) Complete(id int) {
	todo := t.Find(id)
	if todo == nil {
		return
	}

	todo.IsCompleted = true

	t.Remove(todo.Id)
	t.Todos = append(t.Todos, *todo)
}
