package storage

import (
	"Termile/internal/task"
	"encoding/json"
	"io/ioutil"
	"os"
)

// SaveTasks saves the list of tasks to a specified file in JSON format
func SaveTasks(filename string, tasks []task.Task) error {
	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// LoadTasks loads the list of tasks from a specified JSON file
func LoadTasks(filename string) ([]task.Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var tasks []task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}
