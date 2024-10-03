package ui

import (
	"Termile/internal/task"
	"fmt"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"strings"
)

// StartUI starts the terminal UI for task and subtask management
func StartUI(tm *task.TaskManager) {
	taskList := widgets.NewList()
	taskList.Title = "Tasks"
	taskList.Rows = []string{"No tasks available"}
	taskList.SelectedRowStyle = termui.NewStyle(termui.ColorYellow)

	subtaskList := widgets.NewList()
	subtaskList.Title = "Subtasks"
	subtaskList.Rows = []string{"No subtasks available"}
	subtaskList.SelectedRowStyle = termui.NewStyle(termui.ColorCyan)
	subtaskList.SetRect(0, 18, 80, 30)

	taskInput := widgets.NewParagraph()
	taskInput.Title = "Enter new task"
	taskInput.Text = ""
	taskInput.SetRect(0, 0, 80, 3)

	taskList.SetRect(0, 3, 80, 18)

	termui.Render(taskList, subtaskList, taskInput)

	uiEvents := termui.PollEvents()

	typingMode := false
	editingMode := false
	inputBuffer := strings.Builder{}
	selectedTaskIndex := 0
	selectedSubtaskIndex := 0
	inSubtaskMode := false // Track if user is managing subtasks

	updateTaskList(taskList, tm)

	for {
		e := <-uiEvents

		switch e.ID {
		case "<C-q>", "<C-c>":
			return // Exit the app

		case "<C-a>": // Enter task typing mode (add mode)
			typingMode = true
			editingMode = false
			inputBuffer.Reset()
			taskInput.Text = ""
			taskInput.Title = "Enter new task"
			termui.Render(taskInput)

		case "<C-A>": // Add subtask (when in subtask mode)
			if inSubtaskMode {
				typingMode = true
				editingMode = false
				inputBuffer.Reset()
				taskInput.Text = ""
				taskInput.Title = "Enter new subtask"
				termui.Render(taskInput)
			}

		case "<Backspace>": // Handle backspace during task input
			if typingMode {
				currentText := inputBuffer.String()
				if len(currentText) > 0 {
					// Remove the last character from the buffer
					currentText = currentText[:len(currentText)-1]
					inputBuffer.Reset()
					inputBuffer.WriteString(currentText)
					taskInput.Text = inputBuffer.String()
					termui.Render(taskInput)
				}
			}

		case "<C-e>": // Enter edit mode for the selected task or subtask
			if !inSubtaskMode && len(tm.ListTasks()) > 0 {
				typingMode = true
				editingMode = true
				inputBuffer.Reset()
				taskInput.Title = "Edit task"
				taskInput.Text = tm.ListTasks()[selectedTaskIndex].Title
				inputBuffer.WriteString(tm.ListTasks()[selectedTaskIndex].Title)
				termui.Render(taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 {
				typingMode = true
				editingMode = true
				inputBuffer.Reset()
				taskInput.Title = "Edit subtask"
				taskInput.Text = tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].Title
				inputBuffer.WriteString(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].Title)
				termui.Render(taskInput)
			}

		case "<C-s>": // Switch to subtask mode
			if len(tm.ListTasks()) > 0 {
				inSubtaskMode = !inSubtaskMode
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				termui.Render(subtaskList, taskList, taskInput)
			}

		case "<C-d>": // Delete the selected task or subtask
			if !inSubtaskMode && len(tm.ListTasks()) > 0 {
				tm.RemoveTask(tm.ListTasks()[selectedTaskIndex].ID)
				if selectedTaskIndex > 0 {
					selectedTaskIndex--
				}
				updateTaskList(taskList, tm)
				termui.Render(taskList, taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 {
				tm.RemoveSubtask(tm.ListTasks()[selectedTaskIndex].ID, tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].ID)
				if selectedSubtaskIndex > 0 {
					selectedSubtaskIndex--
				}
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				termui.Render(subtaskList, taskInput)
			}

		case "<Enter>": // Submit the task or subtask (add/edit)
			if typingMode {
				taskTitle := inputBuffer.String()
				if taskTitle != "" {
					if editingMode {
						if !inSubtaskMode {
							tm.EditTask(tm.ListTasks()[selectedTaskIndex].ID, taskTitle)
						} else {
							tm.EditSubtask(tm.ListTasks()[selectedTaskIndex].ID, tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].ID, taskTitle)
						}
					} else {
						if !inSubtaskMode {
							tm.AddTask(taskTitle)
						} else {
							tm.AddSubtask(tm.ListTasks()[selectedTaskIndex].ID, taskTitle)
						}
					}
					updateTaskList(taskList, tm)
					updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				}
				typingMode = false
				taskInput.Text = ""
				taskInput.Title = "Enter new task"
				termui.Render(taskList, subtaskList, taskInput)
			}

		case "<C-j>", "<Down>": // Move selection down
			if !inSubtaskMode && len(tm.ListTasks()) > 0 && selectedTaskIndex < len(tm.ListTasks())-1 {
				selectedTaskIndex++
				taskList.SelectedRow = selectedTaskIndex
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID) // Show subtasks for the new selected task
				termui.Render(taskList, subtaskList)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 && selectedSubtaskIndex < len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID))-1 {
				selectedSubtaskIndex++
				subtaskList.SelectedRow = selectedSubtaskIndex
				termui.Render(subtaskList)
			}

		case "<C-k>", "<Up>": // Move selection up
			if !inSubtaskMode && len(tm.ListTasks()) > 0 && selectedTaskIndex > 0 {
				selectedTaskIndex--
				taskList.SelectedRow = selectedTaskIndex
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID) // Show subtasks for the new selected task
				termui.Render(taskList, subtaskList)
			} else if inSubtaskMode && selectedSubtaskIndex > 0 {
				selectedSubtaskIndex--
				subtaskList.SelectedRow = selectedSubtaskIndex
				termui.Render(subtaskList)
			}

		case "<C-t>": // Toggle task completion (mark as done/undone)
			if !inSubtaskMode && len(tm.ListTasks()) > 0 {
				// Task mode: Toggle the selected task's completion status
				tm.ToggleComplete(tm.ListTasks()[selectedTaskIndex].ID)
				updateTaskList(taskList, tm)
				termui.Render(taskList, subtaskList, taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 {
				// Subtask mode: Toggle the selected subtask's completion status
				tm.ToggleSubtaskComplete(tm.ListTasks()[selectedTaskIndex].ID, tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].ID)
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				termui.Render(taskList, subtaskList, taskInput)
			}

		default:
			if typingMode {
				if len(e.ID) == 1 {
					inputBuffer.WriteString(e.ID)
					taskInput.Text = inputBuffer.String()
					termui.Render(taskInput)
				}
			}
		}

		termui.Render(taskList, subtaskList, taskInput)
	}
}

// updateTaskList updates the task list with the current tasks
func updateTaskList(taskList *widgets.List, tm *task.TaskManager) {
	tasks := tm.ListTasks()
	rows := []string{}
	for _, task := range tasks {
		status := "[ ]"
		if task.Complete {
			status = "[x]"
		}
		rows = append(rows, fmt.Sprintf("%d. %s %s", task.ID, status, task.Title))
	}
	taskList.Rows = rows

	if len(rows) == 0 {
		taskList.Rows = []string{"No tasks available"}
	}
}

// updateSubtaskList updates the subtask list with the current subtasks for a specific task
func updateSubtaskList(subtaskList *widgets.List, tm *task.TaskManager, taskID int) {
	subtasks := tm.ListSubtasks(taskID)
	rows := []string{}
	for _, subtask := range subtasks {
		status := "[ ]"
		if subtask.Complete {
			status = "[x]"
		}
		rows = append(rows, fmt.Sprintf("%d. %s %s", subtask.ID, status, subtask.Title))
	}
	subtaskList.Rows = rows

	if len(rows) == 0 {
		subtaskList.Rows = []string{"No subtasks available"}
	}
}
