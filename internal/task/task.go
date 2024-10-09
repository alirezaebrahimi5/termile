package task

import (
	"fmt"
	"time"
)

// Project represents a project with tasks.
type Project struct {
	ID          int
	Name        string
	Description string
	Tasks       []Task
	CreatedAt   time.Time
}

// Task represents a task with subtasks.
type Task struct {
	ID          int
	Title       string
	Description string
	AssignedTo  string // New field
	Complete    bool
	Subtasks    []Subtask
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// Subtask represents a subtask.
type Subtask struct {
	ID          int
	Title       string
	Description string
	AssignedTo  string // New field
	Complete    bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// TaskManager manages a list of projects, tasks, and subtasks.
type TaskManager struct {
	projects      []Project
	tasks         []Task
	subtasks      []Subtask
	nextProjectID int
	nextTaskID    int
	nextSubID     int
}

// NewTaskManager creates a new TaskManager.
func NewTaskManager() *TaskManager {
	return &TaskManager{
		projects:      []Project{},
		tasks:         []Task{},
		subtasks:      []Subtask{},
		nextProjectID: 1,
		nextTaskID:    1,
		nextSubID:     1,
	}
}

// AddProject adds a new project to the TaskManager.
func (tm *TaskManager) AddProject(project Project) {
	project.ID = tm.getNextProjectID()
	tm.projects = append(tm.projects, project)
}

// AddTask adds a new task to a specific project.
func (tm *TaskManager) AddTask(projectID int, task Task) {
	for i, project := range tm.projects {
		if project.ID == projectID {
			task.ID = tm.getNextTaskID()
			tm.projects[i].Tasks = append(tm.projects[i].Tasks, task)
			break
		}
	}
}

// AddSubtask adds a new subtask to a specific task within a project.
func (tm *TaskManager) AddSubtask(projectID int, taskID int, subtask Subtask) {
	for i, project := range tm.projects {
		if project.ID == projectID {
			for j, taskItem := range project.Tasks {
				if taskItem.ID == taskID {
					subtask.ID = tm.getNextSubtaskID()
					tm.projects[i].Tasks[j].Subtasks = append(tm.projects[i].Tasks[j].Subtasks, subtask)
					break
				}
			}
			break
		}
	}
}

// AssignTaskTo assigns a task to someone.
func (tm *TaskManager) AssignTaskTo(projectID int, taskID int, assignedTo string) {
	for i := range tm.projects {
		if tm.projects[i].ID == projectID {
			for j := range tm.projects[i].Tasks {
				if tm.projects[i].Tasks[j].ID == taskID {
					tm.projects[i].Tasks[j].AssignedTo = assignedTo
					return
				}
			}
		}
	}
}

// AssignSubtaskTo assigns a subtask to someone.
func (tm *TaskManager) AssignSubtaskTo(projectID int, taskID int, subtaskID int, assignedTo string) {
	for i := range tm.projects {
		if tm.projects[i].ID == projectID {
			for j := range tm.projects[i].Tasks {
				if tm.projects[i].Tasks[j].ID == taskID {
					for k := range tm.projects[i].Tasks[j].Subtasks {
						if tm.projects[i].Tasks[j].Subtasks[k].ID == subtaskID {
							tm.projects[i].Tasks[j].Subtasks[k].AssignedTo = assignedTo
							return
						}
					}
				}
			}
		}
	}
}

// ListTasks returns the list of tasks for a given project.
func (tm *TaskManager) ListTasks(projectID int) []Task {
	for _, project := range tm.projects {
		if project.ID == projectID {
			return project.Tasks
		}
	}
	return nil
}

// ListSubtasks returns the list of subtasks for a given task in a project.
func (tm *TaskManager) ListSubtasks(projectID, taskID int) []Subtask {
	for _, project := range tm.projects {
		if project.ID == projectID {
			for _, task := range project.Tasks {
				if task.ID == taskID {
					return task.Subtasks
				}
			}
		}
	}
	return nil
}

// ToggleComplete toggles the completion status of a task by ID.
func (tm *TaskManager) ToggleComplete(projectID int, taskID int) {
	for i := range tm.projects {
		if tm.projects[i].ID == projectID {
			for j := range tm.projects[i].Tasks {
				if tm.projects[i].Tasks[j].ID == taskID {
					tm.projects[i].Tasks[j].Complete = !tm.projects[i].Tasks[j].Complete
					if tm.projects[i].Tasks[j].Complete {
						now := time.Now()
						tm.projects[i].Tasks[j].CompletedAt = &now
					} else {
						tm.projects[i].Tasks[j].CompletedAt = nil
					}
					break
				}
			}
			break
		}
	}
}

// ToggleSubtaskComplete toggles the completion status of a subtask by ID.
func (tm *TaskManager) ToggleSubtaskComplete(projectID int, taskID int, subtaskID int) {
	for i := range tm.projects {
		if tm.projects[i].ID == projectID {
			for j := range tm.projects[i].Tasks {
				if tm.projects[i].Tasks[j].ID == taskID {
					for k := range tm.projects[i].Tasks[j].Subtasks {
						if tm.projects[i].Tasks[j].Subtasks[k].ID == subtaskID {
							tm.projects[i].Tasks[j].Subtasks[k].Complete = !tm.projects[i].Tasks[j].Subtasks[k].Complete
							if tm.projects[i].Tasks[j].Subtasks[k].Complete {
								now := time.Now()
								tm.projects[i].Tasks[j].Subtasks[k].CompletedAt = &now
							} else {
								tm.projects[i].Tasks[j].Subtasks[k].CompletedAt = nil
							}
							break
						}
					}
					break
				}
			}
			break
		}
	}
}

// RemoveTask removes a task by ID from a specific project.
func (tm *TaskManager) RemoveTask(projectID int, taskID int) {
	for i := range tm.projects {
		if tm.projects[i].ID == projectID {
			tasks := &tm.projects[i].Tasks
			for j, task := range *tasks {
				if task.ID == taskID {
					*tasks = append((*tasks)[:j], (*tasks)[j+1:]...)
					return
				}
			}
		}
	}
}

// RemoveSubtask removes a subtask by ID from a specific task in a project.
func (tm *TaskManager) RemoveSubtask(projectID int, taskID int, subtaskID int) {
	for i := range tm.projects {
		if tm.projects[i].ID == projectID {
			for j := range tm.projects[i].Tasks {
				if tm.projects[i].Tasks[j].ID == taskID {
					subtasks := &tm.projects[i].Tasks[j].Subtasks
					for k, subtask := range *subtasks {
						if subtask.ID == subtaskID {
							*subtasks = append((*subtasks)[:k], (*subtasks)[k+1:]...)
							return
						}
					}
				}
			}
		}
	}
}

// ListProjects returns the list of all projects.
func (tm *TaskManager) ListProjects() []Project {
	return tm.projects
}

// SetProjects sets the projects and updates the next IDs accordingly.
func (tm *TaskManager) SetProjects(projects []Project) {
	tm.projects = projects
	// Update nextProjectID, nextTaskID, and nextSubID based on existing IDs
	maxProjectID := 0
	maxTaskID := 0
	maxSubID := 0
	for _, project := range projects {
		if project.ID > maxProjectID {
			maxProjectID = project.ID
		}
		for _, task := range project.Tasks {
			if task.ID > maxTaskID {
				maxTaskID = task.ID
			}
			for _, subtask := range task.Subtasks {
				if subtask.ID > maxSubID {
					maxSubID = subtask.ID
				}
			}
		}
	}
	tm.nextProjectID = maxProjectID + 1
	tm.nextTaskID = maxTaskID + 1
	tm.nextSubID = maxSubID + 1
}

// EditTask updates the title and description of a task
func (tm *TaskManager) EditTask(projectID, taskID int, newTitle, newDescription string) error {
	// Validate project existence
	project := tm.projects[projectID]

	// Locate the subtask by ID
	for i, task := range project.Tasks {
		if task.ID == taskID {
			tm.projects[projectID].Tasks[i].Title = newTitle
			tm.projects[projectID].Tasks[i].Description = newDescription
			break
		}
	}

	return fmt.Errorf("subtask with ID %d not found in task %d", taskID, projectID)
}

// EditSubtask updates the title and description of a subtask
func (tm *TaskManager) EditSubtask(projectID int, taskID int, subtaskID int, newTitle string, newDescription string) error {
	// Validate project existence
	project := tm.projects[projectID]

	// Validate task existence
	task := project.Tasks[taskID]

	// Locate the subtask by ID
	for i, subtask := range task.Subtasks {
		if subtask.ID == subtaskID {
			tm.projects[projectID].Tasks[taskID].Subtasks[i].Title = newTitle
			tm.projects[projectID].Tasks[taskID].Subtasks[i].Description = newDescription
			break
		}
	}

	return fmt.Errorf("subtask with ID %d not found in task %d", subtaskID, taskID)
}

// EditProject edits the name and description of a project by ID.
func (tm *TaskManager) EditProject(projectID int, newName string, newDescription string) {
	for i := range tm.projects {
		if tm.projects[i].ID == projectID {
			tm.projects[i].Name = newName
			tm.projects[i].Description = newDescription
			break
		}
	}
}

// RemoveProject removes a project by ID.
func (tm *TaskManager) RemoveProject(projectID int) {
	for i, project := range tm.projects {
		if project.ID == projectID {
			tm.projects = append(tm.projects[:i], tm.projects[i+1:]...)
			return
		}
	}
}

// getNextProjectID generates the next project ID.
func (tm *TaskManager) getNextProjectID() int {
	return tm.nextProjectID
}

// getNextTaskID generates the next task ID.
func (tm *TaskManager) getNextTaskID() int {
	return tm.nextTaskID
}

// getNextSubtaskID generates the next subtask ID.
func (tm *TaskManager) getNextSubtaskID() int {
	return tm.nextSubID
}
func (tm *TaskManager) GetProjectByIndex(index int) (*Project, error) {
	projects := tm.ListProjects()
	if index < 0 || index >= len(projects) {
		return nil, fmt.Errorf("project index %d out of range", index)
	}
	return &projects[index], nil
}

func (tm *TaskManager) GetTaskByIndex(projectID, index int) (*Task, error) {
	tasks := tm.ListTasks(projectID)
	if index < 0 || index >= len(tasks) {
		return nil, fmt.Errorf("task index %d out of range", index)
	}
	return &tasks[index], nil
}

// GetProject safely retrieves a project by index
func (tm *TaskManager) GetProject(index int) (*Project, error) {
	projects := tm.ListProjects()
	if index < 0 || index >= len(projects) {
		return nil, fmt.Errorf("project index %d out of range", index)
	}
	return &projects[index], nil
}

// GetTask safely retrieves a task by index
func (tm *TaskManager) GetTask(projectID, index int) (*Task, error) {
	tasks := tm.projects[projectID].Tasks

	return &tasks[index], nil
}

// GetSubtask safely retrieves a subtask by index
func (tm *TaskManager) GetSubtask(projectID, taskID, index int) (*Subtask, error) {
	subtasks := tm.projects[projectID].Tasks[taskID].Subtasks
	return &subtasks[index], nil
}
