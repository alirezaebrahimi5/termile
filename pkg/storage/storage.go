package storage

import (
	"Termile/internal/task"
	"encoding/json"
	"io"
	"os"
)

// SaveProjects saves the list of projects to a specified file in JSON format
func SaveProjects(filename string, projects []task.Project) error {
	data, err := json.Marshal(projects)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadProjects loads the list of projects from a specified JSON file
func LoadProjects(filename string) ([]task.Project, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var projects []task.Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}
