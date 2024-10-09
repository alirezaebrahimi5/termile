package main

import (
	"Termile/internal/task"
	"Termile/internal/ui"
	"Termile/pkg/storage"
	"log"

	"github.com/gizak/termui/v3"
)

const projectFile = "projects.json"

func main() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()
	projects, err := storage.LoadProjects(projectFile)
	if err != nil {
		log.Printf("failed to load projects: %v", err)
	}
	taskManager := task.NewTaskManager()
	taskManager.SetProjects(projects) // Use the setter to load projects into taskManager

	// Start the UI
	ui.StartUI(taskManager)

	// Save tasks when the app exits
	if err := storage.SaveProjects(projectFile, taskManager.ListProjects()); err != nil {
		log.Printf("failed to save projects: %v", err)
	}
}
