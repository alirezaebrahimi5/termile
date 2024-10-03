package main

import (
	"Termile/internal/task"
	"Termile/internal/ui"
	"Termile/pkg/storage"
	"log"

	"github.com/gizak/termui/v3"
)

const taskFile = "tasks.json"

func main() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	// Load tasks from the JSON file
	tasks, err := storage.LoadTasks(taskFile)
	if err != nil {
		log.Printf("failed to load tasks: %v", err)
	}

	taskManager := task.NewTaskManager()
	taskManager.SetTasks(tasks) // Use the setter to load tasks into taskManager

	// Start the UI
	ui.StartUI(taskManager)

	// Save tasks when the app exits
	if err := storage.SaveTasks(taskFile, taskManager.ListTasks()); err != nil {
		log.Printf("failed to save tasks: %v", err)
	}
}
