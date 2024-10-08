package ui

import (
	"Termile/internal/task"
	"Termile/pkg/storage"
	"fmt"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"math"
	"strings"
)

const taskFile = "tasks.json"

// StartUI starts the terminal UI for task and subtask management
func StartUI(tm *task.TaskManager) {
	if err := termui.Init(); err != nil {
		fmt.Printf("failed to initialize termui: %v", err)
		return
	}
	defer termui.Close()

	// Create widgets
	taskList := widgets.NewList()
	taskList.Title = "Tasks"
	taskList.SelectedRowStyle = termui.NewStyle(termui.ColorYellow)

	subtaskList := widgets.NewList()
	subtaskList.Title = "Subtasks"
	subtaskList.SelectedRowStyle = termui.NewStyle(termui.ColorCyan)

	taskInput := widgets.NewParagraph()
	taskInput.Title = "Input"
	taskInput.Text = ""

	description := widgets.NewParagraph()
	description.Title = "Description"
	description.Text = "Select a task or subtask to see its description."
	description.WrapText = true

	// Create a bar chart for task completion statistics
	barChart := widgets.NewBarChart()
	barChart.Title = "Task Completion"
	barChart.Labels = []string{"Completed", "Pending"}
	barChart.BarWidth = 9
	barChart.BarColors = []termui.Color{termui.ColorGreen, termui.ColorRed}
	barChart.NumStyles = []termui.Style{termui.NewStyle(termui.ColorBlack)}

	// Create a gauge for task completion percentage
	gauge := widgets.NewGauge()
	gauge.Title = "⏳ Subtasks Completion ⏳"
	gauge.Percent = 0
	gauge.BarColor = termui.ColorGreen
	gauge.LabelStyle = termui.NewStyle(termui.ColorBlack)
	gauge.TitleStyle = termui.NewStyle(termui.ColorMagenta, termui.ColorClear, termui.ModifierBold)
	gauge.BorderStyle = termui.NewStyle(termui.ColorCyan)
	gauge.PaddingLeft = 1
	gauge.PaddingRight = 1

	// Create a pie chart for task status
	pieChart := widgets.NewPieChart()
	pieChart.Title = "📊 Tasks Status 📊"
	pieChart.AngleOffset = -.5 * math.Pi // Start from the top
	pieChart.Data = []float64{0, 100}    // Initialize with default values
	pieChart.LabelFormatter = func(i int, v float64) string {
		labels := []string{"✅ Completed", "❌ Pending"}
		return fmt.Sprintf("%s\n%.0f%%", labels[i], v)
	}
	pieChart.Colors = []termui.Color{termui.ColorGreen, termui.ColorRed}
	pieChart.BorderStyle = termui.NewStyle(termui.ColorCyan)
	pieChart.TitleStyle = termui.NewStyle(termui.ColorMagenta, termui.ColorClear, termui.ModifierBold)
	pieChart.PaddingTop = 1
	pieChart.PaddingBottom = 1
	pieChart.PaddingLeft = 2
	pieChart.PaddingRight = 2

	// Update the bar chart data
	updateBarChart(barChart, tm)

	// Create a grid and arrange widgets
	grid := termui.NewGrid()
	termWidth, termHeight := termui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		termui.NewRow(1.0,
			termui.NewCol(0.5,
				termui.NewRow(0.4, taskList),
				termui.NewRow(0.4, subtaskList),
				termui.NewRow(0.2, taskInput),
			),
			termui.NewCol(0.5,
				termui.NewRow(0.3, description),
				termui.NewRow(0.3, gauge),
				termui.NewRow(0.4, pieChart), // Increased from 0.35 to 0.4
			),
		),
	)

	// Initial rendering
	updateTaskList(taskList, tm)
	updateSubtaskList(subtaskList, tm, -1) // No task selected initially
	termui.Render(grid)

	// Event handling and other logic remains the same...
	// (You can reuse your existing event loop code here)

	// Event loop variables
	uiEvents := termui.PollEvents()
	typingMode := false
	editingMode := false
	inputBuffer := strings.Builder{}
	selectedTaskIndex := 0
	selectedSubtaskIndex := 0
	inSubtaskMode := false
	inputState := "" // Can be "title" or "description"

	for {
		e := <-uiEvents

		switch e.ID {
		case "<C-q>", "<C-c>":
			return // Exit the app

		case "<C-A>": // Add subtask (when in subtask mode)
			if inSubtaskMode {
				typingMode = true
				editingMode = false
				inputBuffer.Reset()
				taskInput.Text = ""
				taskInput.Title = "Enter new subtask"
				termui.Render(taskInput)
			}
		case "<C-x>":
			if err := storage.SaveTasks(taskFile, tm.ListTasks()); err != nil {
				log.Printf("failed to save tasks: %v", err)
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

		case "<C-e>": // Enter edit mode for the selected task or subtask title
			if !inSubtaskMode && len(tm.ListTasks()) > 0 {
				typingMode = true
				editingMode = true
				inputState = "title"
				inputBuffer.Reset()
				selectedTask := tm.ListTasks()[selectedTaskIndex]
				taskInput.Title = "Edit task title"
				taskInput.Text = selectedTask.Title
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 {
				typingMode = true
				editingMode = true
				inputState = "title"
				inputBuffer.Reset()
				selectedSubtask := tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex]
				taskInput.Title = "Edit subtask title"
				taskInput.Text = selectedSubtask.Title
				inputBuffer.WriteString(taskInput.Text)
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
				updateGauge(gauge, tm, tm.ListTasks()[selectedTaskIndex].ID)
				termui.Render(taskList)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 {
				tm.RemoveSubtask(tm.ListTasks()[selectedTaskIndex].ID, tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].ID)
				if selectedSubtaskIndex > 0 {
					selectedSubtaskIndex--
				}
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				termui.Render(subtaskList, taskInput)
			}

		case "<Enter>":
			if typingMode {
				inputText := inputBuffer.String()
				if inputState == "title" {
					taskTitle := strings.TrimSpace(inputText)
					if editingMode {
						if !inSubtaskMode {
							// Update title, keep existing description
							tm.EditTask(
								tm.ListTasks()[selectedTaskIndex].ID,
								taskTitle,
								tm.ListTasks()[selectedTaskIndex].Description,
							)
						} else {
							tm.EditSubtask(
								tm.ListTasks()[selectedTaskIndex].ID,
								tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].ID,
								taskTitle,
								tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].Description,
							)
						}
					} else {
						if !inSubtaskMode {
							// Add new task with empty description
							tm.AddTask(taskTitle, "")
						} else {
							// Add new subtask with empty description
							tm.AddSubtask(tm.ListTasks()[selectedTaskIndex].ID, taskTitle, "")
						}
					}
				} else if inputState == "description" {
					taskDescription := strings.TrimSpace(inputText)
					if editingMode {
						if !inSubtaskMode {
							// Update description, keep existing title
							tm.EditTask(
								tm.ListTasks()[selectedTaskIndex].ID,
								tm.ListTasks()[selectedTaskIndex].Title,
								taskDescription,
							)
						} else {
							tm.EditSubtask(
								tm.ListTasks()[selectedTaskIndex].ID,
								tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].ID,
								tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].Title,
								taskDescription,
							)
						}
					}
				}
				typingMode = false
				editingMode = false
				inputState = ""
				taskInput.Text = ""
				taskInput.Title = "Input"
				inputBuffer.Reset()
				updateTaskList(taskList, tm)
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				updateDescription(description, tm, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(taskList, subtaskList, description)

			}

		case "<C-a>": // Enter task typing mode (add mode)
			typingMode = true
			editingMode = false
			inputState = "title"
			inputBuffer.Reset()
			taskInput.Text = ""
			if !inSubtaskMode {
				taskInput.Title = "Enter new task title"
			} else {
				taskInput.Title = "Enter new subtask title"
			}
			termui.Render(taskInput)

		case "<C-j>", "<Down>": // Move selection down
			if !inSubtaskMode && len(tm.ListTasks()) > 0 && selectedTaskIndex < len(tm.ListTasks())-1 {
				selectedTaskIndex++
				taskList.SelectedRow = selectedTaskIndex
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				updateGauge(gauge, tm, tm.ListTasks()[selectedTaskIndex].ID)
				updateDescription(description, tm, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(taskList, subtaskList, description)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 && selectedSubtaskIndex < len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID))-1 {
				selectedSubtaskIndex++
				subtaskList.SelectedRow = selectedSubtaskIndex
				updateDescription(description, tm, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(subtaskList, description)
			}

		case "<C-l>": // Add or edit description for the selected task or subtask
			if !inSubtaskMode && len(tm.ListTasks()) > 0 {
				typingMode = true
				editingMode = true
				inputState = "description"
				inputBuffer.Reset()
				selectedTask := tm.ListTasks()[selectedTaskIndex]
				taskInput.Title = "Edit task description"
				taskInput.Text = selectedTask.Description
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 {
				typingMode = true
				editingMode = true
				inputState = "description"
				inputBuffer.Reset()
				selectedSubtask := tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex]
				taskInput.Title = "Edit subtask description"
				taskInput.Text = selectedSubtask.Description
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			}

		case "<C-k>", "<Up>": // Move selection up
			if !inSubtaskMode && len(tm.ListTasks()) > 0 && selectedTaskIndex > 0 {
				selectedTaskIndex--
				taskList.SelectedRow = selectedTaskIndex
				// Reset subtask index when changing tasks
				selectedSubtaskIndex = 0
				subtaskList.SelectedRow = selectedSubtaskIndex
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				updateGauge(gauge, tm, tm.ListTasks()[selectedTaskIndex].ID)
				updateDescription(description, tm, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(taskList, subtaskList, description)
			} else if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 && selectedSubtaskIndex > 0 {
				selectedSubtaskIndex--
				subtaskList.SelectedRow = selectedSubtaskIndex
				updateDescription(description, tm, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(subtaskList, description)
			}

		case "<C-o>":
			showHelpModal()
			termui.Clear() // Clear the screen after closing the help modal
			termui.Render(grid)

		case "<C-t>": // Toggle task completion (mark as done/undone)
			if inSubtaskMode && len(tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)) > 0 {
				// Subtask mode: Toggle the selected subtask's completion status
				tm.ToggleSubtaskComplete(
					tm.ListTasks()[selectedTaskIndex].ID,
					tm.ListSubtasks(tm.ListTasks()[selectedTaskIndex].ID)[selectedSubtaskIndex].ID,
				)
				updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
				updateGauge(gauge, tm, tm.ListTasks()[selectedTaskIndex].ID)
				termui.Render(subtaskList)
			} else if !inSubtaskMode && len(tm.ListTasks()) > 0 {
				// Task mode: Toggle the selected task's completion status
				tm.ToggleComplete(
					tm.ListTasks()[selectedTaskIndex].ID,
				)
				updateTaskList(taskList, tm)
				termui.Render(taskList)
			}

		default:
			if typingMode {
				if e.Type == termui.KeyboardEvent {
					switch e.ID {
					case "<Space>":
						inputBuffer.WriteString(" ")
					case "<Tab>":
						inputBuffer.WriteString("\t")
					case "<Backspace>":
						// Handled in a separate case
					case "<Enter>":
						// Ignore, handled separately
					default:
						if len(e.ID) == 1 {
							inputBuffer.WriteString(e.ID)
						}
					}
					taskInput.Text = inputBuffer.String()
					termui.Render(taskInput)
				}
			}

		}

		termui.Render(taskList, subtaskList, taskInput)
		updateTaskList(taskList, tm)
		if len(tm.ListTasks()) > 0 {
			updateSubtaskList(subtaskList, tm, tm.ListTasks()[selectedTaskIndex].ID)
			updateGauge(gauge, tm, tm.ListTasks()[selectedTaskIndex].ID)
		} else {
			updateSubtaskList(subtaskList, tm, -1)
			updateGauge(gauge, tm, -1)
		}
		updatePieChart(pieChart, tm)
		updateBarChart(barChart, tm)

		updateDescription(description, tm, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
		termWidth, termHeight = termui.TerminalDimensions()
		grid.SetRect(0, 0, termWidth, termHeight)
		termui.Clear()
		termui.Render(grid)

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
	if taskID == -1 {
		subtaskList.Rows = []string{"No subtasks available"}
		return
	}
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

// updateBarChart updates the bar chart with current task completion statistics
func updateBarChart(barChart *widgets.BarChart, tm *task.TaskManager) {
	totalTasks := len(tm.ListTasks())
	if totalTasks == 0 {
		barChart.Data = []float64{0, 0}
		return
	}
	completedTasks := 0
	for _, task := range tm.ListTasks() {
		if task.Complete {
			completedTasks++
		}
	}
	pendingTasks := totalTasks - completedTasks
	barChart.Data = []float64{float64(completedTasks), float64(pendingTasks)}
}

// calculateCompletionPercentage calculates the percentage of completed tasks
func calculateCompletionPercentage(tm *task.TaskManager) int {
	totalTasks := len(tm.ListTasks())
	if totalTasks == 0 {
		return 0
	}
	completedTasks := 0
	for _, task := range tm.ListTasks() {
		if task.Complete {
			completedTasks++
		}
	}
	percentage := (completedTasks * 100) / totalTasks
	return percentage
}

// updateGauge updates the gauge with the percentage of completed subtasks for the selected task
func updateGauge(gauge *widgets.Gauge, tm *task.TaskManager, taskID int) {
	if taskID == -1 {
		// No task selected
		gauge.Percent = 0
		gauge.Label = "No Task Selected"
		gauge.BarColor = termui.ColorYellow
		gauge.LabelStyle = termui.NewStyle(termui.ColorBlack, termui.ColorYellow)
		return
	}

	// Get the subtasks for the selected task
	subtasks := tm.ListSubtasks(taskID)
	totalSubtasks := len(subtasks)
	if totalSubtasks == 0 {
		// No subtasks
		gauge.Percent = 0
		gauge.Label = "No Subtasks"
		gauge.BarColor = termui.ColorYellow
		gauge.LabelStyle = termui.NewStyle(termui.ColorBlack, termui.ColorYellow)
		return
	}
	completedSubtasks := 0
	for _, subtask := range subtasks {
		if subtask.Complete {
			completedSubtasks++
		}
	}
	percentage := (completedSubtasks * 100) / totalSubtasks
	gauge.Percent = percentage
	gauge.Label = fmt.Sprintf("✔ %d%% Complete ✔", percentage)

	// Change bar color based on percentage
	switch {
	case percentage < 20:
		gauge.BarColor = termui.ColorRed
		gauge.LabelStyle = termui.NewStyle(termui.ColorWhite, termui.ColorRed)
	case percentage < 40:
		gauge.BarColor = termui.ColorMagenta
		gauge.LabelStyle = termui.NewStyle(termui.ColorBlack, termui.ColorMagenta)
	case percentage < 60:
		gauge.BarColor = termui.ColorYellow
		gauge.LabelStyle = termui.NewStyle(termui.ColorBlack, termui.ColorYellow)
	case percentage < 80:
		gauge.BarColor = termui.ColorBlue
		gauge.LabelStyle = termui.NewStyle(termui.ColorBlack, termui.ColorBlue)
	default:
		gauge.BarColor = termui.ColorGreen
		gauge.LabelStyle = termui.NewStyle(termui.ColorBlack, termui.ColorGreen)
	}
}

func updatePieChart(pieChart *widgets.PieChart, tm *task.TaskManager) {
	totalTasks := len(tm.ListTasks())
	if totalTasks == 0 {
		pieChart.Data = []float64{0, 100}
		return
	}
	completedTasks := 0
	for _, task := range tm.ListTasks() {
		if task.Complete {
			completedTasks++
		}
	}
	completedPercentage := (float64(completedTasks) / float64(totalTasks)) * 100
	pendingPercentage := 100 - completedPercentage
	pieChart.Data = []float64{completedPercentage, pendingPercentage}
	pieChart.LabelFormatter = func(i int, v float64) string {
		labels := []string{"✅ Completed", "❌ Pending"}
		return fmt.Sprintf("%s\n%.0f%%", labels[i], v)
	}
	pieChart.Colors = []termui.Color{termui.ColorGreen, termui.ColorRed}
}

func updateDescription(description *widgets.Paragraph, tm *task.TaskManager, taskIndex int, subtaskIndex int, inSubtaskMode bool) {
	if inSubtaskMode {
		if taskIndex >= 0 && subtaskIndex >= 0 && len(tm.ListTasks()) > taskIndex {
			task := tm.ListTasks()[taskIndex]
			if len(task.Subtasks) > subtaskIndex {
				subtask := task.Subtasks[subtaskIndex]
				description.Text = subtask.Description
			}
		} else {
			description.Text = "No Subtask Selected"
		}
	} else {
		if taskIndex >= 0 && len(tm.ListTasks()) > taskIndex {
			task := tm.ListTasks()[taskIndex]
			description.Text = task.Description
		} else {
			description.Text = "No Task Selected"
		}
	}
}

// showHelpModal displays a modal with help information
func showHelpModal() {
	helpText := widgets.NewParagraph()
	helpText.Title = "Help"
	helpText.Text = "\n" +
		"Ctrl+q / Ctrl+c: Quit the application\n" +
		"Ctrl+a: Add new task\n" +
		"Ctrl+A: Add new subtask\n" +
		"Ctrl+e: Edit selected task or subtask\n" +
		"Ctrl+d: Delete selected task or subtask\n" +
		"Ctrl+s: Switch to subtask view\n" +
		"Ctrl+l: Add/edit description for task/subtask\n" +
		"Ctrl+k / Up Arrow: Move selection up\n" +
		"Ctrl+j / Down Arrow: Move selection down\n" +
		"Ctrl+t: Toggle task or subtask completion\n"
	helpText.WrapText = true

	termWidth, termHeight := termui.TerminalDimensions()
	helpText.SetRect(termWidth/4, termHeight/4, 3*termWidth/4, 3*termHeight/4)
	termui.Render(helpText)

	// Wait for a key event to close the modal
	for e := range termui.PollEvents() {
		if e.Type == termui.KeyboardEvent {
			break
		}
	}
}
