package task

import "time"

// Task represents a task with a title and completion status
type Task struct {
	ID          int
	Title       string
	Description string // Add this line
	Complete    bool
	Subtasks    []Subtask
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// Subtask represents a subtask under a task
type Subtask struct {
	ID          int
	Title       string
	Description string // Add this line
	Complete    bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// TaskManager manages a list of tasks
type TaskManager struct {
	tasks      []Task
	nextTaskID int
	nextSubID  int
}

// NewTaskManager creates a new TaskManager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks:      []Task{},
		nextTaskID: 1,
		nextSubID:  1,
	}
}

// SetTasks replaces the current list of tasks
func (tm *TaskManager) SetTasks(tasks []Task) {
	tm.tasks = tasks

	// Optionally, update nextTaskID and nextSubID to ensure no ID conflicts
	if len(tasks) > 0 {
		maxTaskID := tasks[0].ID
		maxSubID := 0
		for _, task := range tasks {
			if task.ID > maxTaskID {
				maxTaskID = task.ID
			}
			for _, subtask := range task.Subtasks {
				if subtask.ID > maxSubID {
					maxSubID = subtask.ID
				}
			}
		}
		tm.nextTaskID = maxTaskID + 1
		tm.nextSubID = maxSubID + 1
	}
}

func (tm *TaskManager) AddTask(title string, description string) {
	task := Task{
		ID:          tm.nextTaskID,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
	}
	tm.tasks = append(tm.tasks, task)
	tm.nextTaskID++
}

func (tm *TaskManager) AddSubtask(taskID int, title string, description string) {
	for i := range tm.tasks {
		if tm.tasks[i].ID == taskID {
			subtask := Subtask{
				ID:          tm.nextSubID,
				Title:       title,
				Description: description,
				CreatedAt:   time.Now(),
			}
			tm.tasks[i].Subtasks = append(tm.tasks[i].Subtasks, subtask)
			tm.nextSubID++
			break
		}
	}
}

// ListTasks returns the list of tasks
func (tm *TaskManager) ListTasks() []Task {
	return tm.tasks
}

// ListSubtasks returns the list of subtasks for a specific task
func (tm *TaskManager) ListSubtasks(taskID int) []Subtask {
	for _, task := range tm.tasks {
		if task.ID == taskID {
			return task.Subtasks
		}
	}
	return nil
}

// ToggleComplete toggles the completion status of a task by ID
func (tm *TaskManager) ToggleComplete(taskID int) {
	for i := range tm.tasks {
		if tm.tasks[i].ID == taskID {
			tm.tasks[i].Complete = !tm.tasks[i].Complete
			if tm.tasks[i].Complete {
				now := time.Now()
				tm.tasks[i].CompletedAt = &now
			} else {
				tm.tasks[i].CompletedAt = nil
			}
			break
		}
	}
}

// ToggleSubtaskComplete toggles the completion status of a subtask by ID
func (tm *TaskManager) ToggleSubtaskComplete(taskID, subtaskID int) {
	for i := range tm.tasks {
		if tm.tasks[i].ID == taskID {
			for j := range tm.tasks[i].Subtasks {
				if tm.tasks[i].Subtasks[j].ID == subtaskID {
					tm.tasks[i].Subtasks[j].Complete = !tm.tasks[i].Subtasks[j].Complete
					if tm.tasks[i].Subtasks[j].Complete {
						now := time.Now()
						tm.tasks[i].Subtasks[j].CompletedAt = &now
					} else {
						tm.tasks[i].Subtasks[j].CompletedAt = nil
					}
					break
				}
			}
			break
		}
	}
}

// RemoveTask removes a task by ID
func (tm *TaskManager) RemoveTask(taskID int) {
	for i, task := range tm.tasks {
		if task.ID == taskID {
			tm.tasks = append(tm.tasks[:i], tm.tasks[i+1:]...)
			break
		}
	}
}

// RemoveSubtask removes a subtask by ID
func (tm *TaskManager) RemoveSubtask(taskID, subtaskID int) {
	for i := range tm.tasks {
		if tm.tasks[i].ID == taskID {
			for j, subtask := range tm.tasks[i].Subtasks {
				if subtask.ID == subtaskID {
					tm.tasks[i].Subtasks = append(tm.tasks[i].Subtasks[:j], tm.tasks[i].Subtasks[j+1:]...)
					break
				}
			}
			break
		}
	}
}

func (tm *TaskManager) EditTask(taskID int, title string, description string) {
	for i := range tm.tasks {
		if tm.tasks[i].ID == taskID {
			tm.tasks[i].Title = title
			tm.tasks[i].Description = description
			break
		}
	}
}

func (tm *TaskManager) EditSubtask(taskID int, subtaskID int, title string, description string) {
	for i := range tm.tasks {
		if tm.tasks[i].ID == taskID {
			for j := range tm.tasks[i].Subtasks {
				if tm.tasks[i].Subtasks[j].ID == subtaskID {
					tm.tasks[i].Subtasks[j].Title = title
					tm.tasks[i].Subtasks[j].Description = description
					break
				}
			}
			break
		}
	}
}

// GetTasksByCompletion returns tasks filtered by completion status
func (tm *TaskManager) GetTasksByCompletion(completed bool) []Task {
	var tasks []Task
	for _, task := range tm.tasks {
		if task.Complete == completed {
			tasks = append(tasks, task)
		}
	}
	return tasks
}
