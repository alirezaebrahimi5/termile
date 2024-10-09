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

const projectFile = "projects.json"

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

	projectList := widgets.NewList()
	projectList.Title = "Projects"
	projectList.SelectedRowStyle = termui.NewStyle(termui.ColorGreen)

	// Create a grid and arrange widgets
	grid := termui.NewGrid()
	termWidth, termHeight := termui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		termui.NewRow(1.0,
			termui.NewCol(0.25,
				termui.NewRow(0.9, projectList),
				termui.NewRow(0.1, taskInput), // Adjust proportions as needed
			),
			termui.NewCol(0.35,
				termui.NewRow(0.5, taskList),
				termui.NewRow(0.5, subtaskList),
			),
			termui.NewCol(0.4,
				termui.NewRow(0.3, description),
				termui.NewRow(0.3, gauge),
				termui.NewRow(0.4, pieChart),
			),
		),
	)

	// Event handling and other logic remains the same...
	// (You can reuse your existing event loop code here)

	// Event loop variables
	uiEvents := termui.PollEvents()
	typingMode := false
	inputBuffer := strings.Builder{}
	selectedTaskIndex := 0
	selectedSubtaskIndex := 0
	selectedProjectIndex := 0
	inSubtaskMode := false
	inputState := ""        // Can be "title" or "description"
	selectedProjectID := -1 // ID of the currently selected project
	selectedTaskID := -1    // ID of the currently selected task
	selectedSubTaskID := -1
	inProjectMode := true

	projects := tm.ListProjects()
	if len(projects) > 0 {
		selectedProjectIndex = 0
		selectedProjectID = projects[0].ID
		projectList.SelectedRow = selectedProjectIndex
	}

	updateProjectList(projectList, tm, 0)
	updateTaskList(taskList, tm, selectedProjectID)
	updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
	updateBarChart(barChart, tm, selectedProjectID)
	updatePieChart(pieChart, tm, selectedProjectID)
	updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
	updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
	termui.Render(grid)

	for {
		e := <-uiEvents

		switch e.ID {
		case "<C-q>", "<C-c>":
			return // Exit the app

		case "<C-p>": // Switch to project mode
			inProjectMode = true
			inSubtaskMode = false
			updateProjectList(projectList, tm, 0)
			termui.Render(projectList)

		case "<C-A>": // Add subtask (when in subtask mode)
			if inSubtaskMode {
				typingMode = true

				inputState = "subtask"
				inputBuffer.Reset()
				taskInput.Text = ""
				taskInput.Title = "Enter new subtask title"
				termui.Render(taskInput)
			}
		case "<C-x>":
			if err := storage.SaveProjects(projectFile, tm.ListProjects()); err != nil {
				log.Printf("failed to save projects: %v", err)
			}

		case "<Backspace>": // Handle backspace during getTask input
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

		case "<C-b>":
			showTaskTreeModal(tm)
			termui.Clear() // Clear the screen after closing the tree modal
			termui.Render(grid)

		case "<C-e>": // Enter edit mode for the selected getTask or subtask title
			if inProjectMode && len(tm.ListProjects()) > 0 {
				typingMode = true

				inputState = "edit_project_name"
				inputBuffer.Reset()
				if selectedProjectIndex >= 0 && selectedProjectIndex < len(tm.ListProjects()) {
					selectedProject := tm.ListProjects()[selectedProjectIndex]
					taskInput.Title = "Edit project name"
					taskInput.Text = selectedProject.Name
					inputBuffer.WriteString(taskInput.Text)
					termui.Render(taskInput)
				} else {
					log.Printf("Selected project index %d is out of range", selectedProjectIndex)
				}
			} else if !inSubtaskMode && len(tm.ListTasks(selectedProjectID)) > 0 {
				// Editing a Task Title
				typingMode = true

				inputState = "edit_task_name"
				inputBuffer.Reset()
				selectedTask := tm.ListTasks(selectedProjectID)[selectedTaskIndex]
				taskInput.Title = "Edit getTask title"
				taskInput.Text = selectedTask.Title
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(selectedProjectID, selectedTaskID)) > 0 {
				// Editing a Subtask Title
				typingMode = true

				inputState = "edit_subtask_name"
				inputBuffer.Reset()
				selectedSubtask := tm.ListSubtasks(selectedProjectID, selectedTaskID)[selectedSubtaskIndex]
				taskInput.Title = "Edit subtask title"
				taskInput.Text = selectedSubtask.Title
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			}

		case "<C-s>": // Switch to subtask mode
			if len(tm.ListTasks(selectedProjectID)) > 0 {
				inSubtaskMode = !inSubtaskMode
				selectedProjectIndex = 0
				subtaskList.SelectedRow = selectedTaskIndex
				updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
				termui.Render(subtaskList, taskList, taskInput)
			}

		case "<C-d>": // Delete the selected getTask or subtask
			if inProjectMode && len(tm.ListProjects()) > 0 {
				if selectedProjectIndex >= 0 && selectedProjectIndex < len(tm.ListProjects()) {
					selectedProjectID = tm.ListProjects()[selectedProjectIndex].ID
					projectList.SelectedRow = selectedProjectIndex
					tm.RemoveProject(tm.ListProjects()[selectedProjectIndex].ID)
					// Adjust selectedProjectIndex
					if selectedProjectIndex >= len(tm.ListProjects()) && selectedProjectIndex > 0 {
						selectedProjectIndex--
					}
					// Update projectList
					updateProjectList(projectList, tm, 0)
					// Set selectedProjectID
					if selectedProjectIndex >= 0 && selectedProjectIndex < len(tm.ListProjects()) {
						selectedProjectID = tm.ListProjects()[selectedProjectIndex].ID
						projectList.SelectedRow = selectedProjectIndex
					} else {
						selectedProjectID = -1
						projectList.SelectedRow = -1
					}
					// Reset getTask and subtask indices and IDs
					selectedTaskIndex = 0
					selectedTaskID = -1
					selectedSubTaskID = -1
					selectedSubtaskIndex = 0
					updateTaskList(taskList, tm, selectedProjectID)
					updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
					updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
					termui.Render(projectList, taskList, subtaskList, description)
				} else {
					log.Printf("Selected project index %d is out of range during deletion", selectedProjectIndex)
				}
			} else if !inSubtaskMode && len(tm.ListTasks(selectedProjectID)) > 0 {
				tm.RemoveTask(selectedProjectID, tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID)
				if selectedTaskIndex > 0 {
					selectedTaskIndex--
				}
				updateTaskList(taskList, tm, selectedProjectID)
				updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
				termui.Render(taskList)
			} else if inSubtaskMode && len(tm.ListSubtasks(selectedProjectID, selectedTaskID)) > 0 {
				tm.RemoveSubtask(selectedProjectID, selectedTaskID, tm.ListSubtasks(selectedProjectID, selectedTaskID)[selectedSubtaskIndex].ID)
				if selectedSubtaskIndex >= len(tm.ListSubtasks(selectedProjectID, selectedTaskID)) && selectedSubtaskIndex > 0 {
					selectedSubtaskIndex--
				}
				updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
				termui.Render(subtaskList, taskInput)
			}

		case "<Enter>":
			if typingMode {
				inputText := strings.TrimSpace(inputBuffer.String())
				switch inputState {
				case "edit_project_name":
					project, err := tm.GetProject(selectedProjectIndex)
					if err != nil {
						log.Printf("Error editing project name: %v", err)
						break
					}
					tm.EditProject(project.ID, inputText, project.Description)

					// Prompt for project description
					typingMode = true

					inputState = "edit_project_description"
					inputBuffer.Reset()
					taskInput.Title = "Edit project description"
					taskInput.Text = project.Description
					inputBuffer.WriteString(project.Description)
					termui.Render(taskInput)

				case "edit_task_name":
					selectedTask, err := tm.GetTask(selectedProjectIndex, selectedTaskID)

					if err != nil {
						log.Printf("Error editing project name: %v", err)
						break
					}

					if err := tm.EditTask(selectedProjectIndex, selectedTask.ID, inputText, selectedTask.Description); err != nil {
						log.Println(err)
					}

					// Prompt for project description
					typingMode = true

					inputState = "edit_task_description"
					inputBuffer.Reset()
					taskInput.Title = "Edit getTask description"
					taskInput.Text = selectedTask.Description
					inputBuffer.WriteString(selectedTask.Description)
					termui.Render(taskInput)
				case "edit_subtask_name":
					selectedSubTask, err := tm.GetSubtask(selectedProjectIndex, selectedTaskIndex, selectedSubTaskID)

					if err != nil {
						log.Printf("Error editing project name: %v", err)
						break
					}
					if err := tm.EditSubtask(selectedProjectIndex, selectedTaskIndex, selectedSubTask.ID, inputText, selectedSubTask.Description); err != nil {
						log.Println(err)
					}
					// Prompt for project description
					typingMode = true

					inputState = "edit_subtask_description"
					inputBuffer.Reset()
					taskInput.Title = "Edit subtask description"
					taskInput.Text = selectedSubTask.Description
					inputBuffer.WriteString(selectedSubTask.Description)
					termui.Render(taskInput)
				case "edit_project_description":
					project, err := tm.GetProject(selectedProjectIndex)
					if err != nil {
						log.Printf("Error editing project description: %v", err)
						break
					}
					tm.EditProject(project.ID, project.Name, inputText)

					// Reset input states
					typingMode = false

					inputState = ""
					inputBuffer.Reset()
					taskInput.Text = ""
					taskInput.Title = "Input"

					// Update UI
					updateProjectList(projectList, tm, selectedProjectIndex)
					projectList.SelectedRow = selectedProjectIndex
					termui.Render(projectList)

				case "description":
					if !inSubtaskMode {
						task, err := tm.GetTask(selectedProjectID, selectedTaskIndex)
						if err != nil {
							log.Printf("Error editing getTask description: %v", err)
							break
						}
						tm.EditTask(selectedProjectID, task.ID, task.Title, inputText)
					} else {
						subtask, err := tm.GetSubtask(selectedProjectID, selectedTaskID, selectedSubtaskIndex)
						if err != nil {
							log.Printf("Error editing subtask description: %v", err)
							break
						}
						tm.EditSubtask(selectedProjectID, selectedTaskID, subtask.ID, subtask.Title, inputText)
					}

				case "assign":
					if !inSubtaskMode {
						task, err := tm.GetTask(selectedProjectID, selectedTaskIndex)
						if err != nil {
							log.Printf("Error assigning getTask: %v", err)
							break
						}
						tm.AssignTaskTo(selectedProjectID, task.ID, inputText)
					} else {
						subtask, err := tm.GetSubtask(selectedProjectID, selectedTaskID, selectedSubtaskIndex)
						if err != nil {
							log.Printf("Error assigning subtask: %v", err)
							break
						}
						tm.AssignSubtaskTo(selectedProjectID, selectedTaskID, subtask.ID, inputText)
					}

				case "title":
					if !inSubtaskMode {
						task, err := tm.GetTask(selectedProjectID, selectedTaskIndex)
						if err != nil {
							log.Printf("Error editing getTask title: %v", err)
							break
						}
						err = tm.EditTask(selectedProjectID, task.ID, inputText, task.Description)
						if err != nil {
							log.Print("")
						}
					} else {
						subtask, err := tm.GetSubtask(selectedProjectID, selectedTaskID, selectedSubtaskIndex)
						if err != nil {
							log.Printf("Error editing getTask title: %v", err)
							break
						}
						tm.EditSubtask(selectedProjectID, selectedTaskID, subtask.ID, inputText, subtask.Description)
					}

					// Reset input states
					typingMode = false

					inputState = ""
					inputBuffer.Reset()
					taskInput.Text = ""
					taskInput.Title = "Input"

				case "project":
					if inputText != "" {
						newProject := task.Project{
							Name:        inputText,
							Description: "", // Optionally prompt for description
							Tasks:       []task.Task{},
						}
						tm.AddProject(newProject)

						// Update and select the new project
						projects := tm.ListProjects()
						selectedProjectIndex = len(projects) - 1
						selectedProjectID = projects[selectedProjectIndex].ID

						updateProjectList(projectList, tm, selectedProjectIndex)
						updateTaskList(taskList, tm, selectedProjectID)
						updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
						updateBarChart(barChart, tm, selectedProjectID)
						updatePieChart(pieChart, tm, selectedProjectID)
						updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
						updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
						termui.Render(projectList, taskList, subtaskList, description, barChart, pieChart, gauge)
						inProjectMode = true
						inSubtaskMode = false
					}

				case "getTask":
					if inputText != "" && selectedProjectID != -1 {
						newTask := task.Task{
							Title:       inputText,
							Description: "",
							Complete:    false,
							AssignedTo:  "",
							Subtasks:    []task.Subtask{},
						}
						tm.AddTask(selectedProjectID, newTask)

						// Update and select the new getTask
						tasks := tm.ListTasks(selectedProjectID)
						selectedTaskIndex = len(tasks) - 1
						selectedTaskID = tasks[selectedTaskIndex].ID

						updateTaskList(taskList, tm, selectedProjectID)
					}

				case "subtask":
					if inputText != "" && selectedProjectID != -1 && selectedTaskID != -1 {
						newSubtask := task.Subtask{
							Title:       inputText,
							Description: "",
							Complete:    false,
							AssignedTo:  "",
						}
						tm.AddSubtask(selectedProjectID, selectedTaskID, newSubtask)

						// Update and select the new subtask
						subtasks := tm.ListSubtasks(selectedProjectID, selectedTaskID)
						selectedSubtaskIndex = len(subtasks) - 1

						updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
					}

				default:
					log.Printf("Unhandled inputState: %s", inputState)
				}

				// After handling inputState, perform common resets and updates if not already done
				if inputState != "edit_project_description" && inputState != "project" && inputState != "getTask" && inputState != "subtask" {
					typingMode = false

					inputState = ""
					inputBuffer.Reset()
					taskInput.Text = ""
					taskInput.Title = "Input"
				}

				// Update UI elements after handling
				updateProjectList(projectList, tm, selectedProjectID)
				updateTaskList(taskList, tm, selectedProjectID)
				updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
				updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				updatePieChart(pieChart, tm, selectedProjectID)
				updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
				termWidth, termHeight = termui.TerminalDimensions()
				grid.SetRect(0, 0, termWidth, termHeight)
				termui.Clear()
				termui.Render(grid)
			}

		case "<C-a>": // Enter getTask typing mode (add mode)
			typingMode = true

			inputBuffer.Reset()
			if inProjectMode {
				inputState = "project"
				taskInput.Title = "Enter new project name"
			} else if !inSubtaskMode {
				inputState = "getTask"
				taskInput.Title = "Enter new getTask title"
			} else {
				inputState = "subtask"
				taskInput.Title = "Enter new subtask title"
			}
			taskInput.Text = ""
			termui.Render(taskInput)

		case "<C-j>", "<Down>": // Move selection down
			if inProjectMode && len(tm.ListProjects()) > 0 && selectedProjectIndex < len(tm.ListProjects())-1 {
				selectedProjectIndex++
				projectList.SelectedRow = selectedProjectIndex
				selectedProjectID = tm.ListProjects()[selectedProjectIndex].ID

				// Reset getTask and subtask indices and IDs
				selectedTaskIndex = 0
				selectedTaskID = -1
				selectedSubtaskIndex = 0

				updateTaskList(taskList, tm, selectedProjectID)
				updateSubtaskList(subtaskList, tm, selectedProjectID, -1) // No getTask selected
				termui.Render(taskList, subtaskList, description, projectList)
			} else if !inProjectMode && !inSubtaskMode && len(tm.ListTasks(selectedProjectID)) > 0 && selectedTaskIndex < len(tm.ListTasks(selectedProjectID))-1 {
				selectedTaskIndex++
				taskList.SelectedRow = selectedTaskIndex
				selectedTaskID = tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID

				// Reset subtask index and ID
				selectedSubtaskIndex = 0

				updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
				updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
				updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(taskList, subtaskList, description)
			} else if inSubtaskMode && len(tm.ListSubtasks(selectedProjectID, selectedTaskID)) > 0 && selectedSubtaskIndex < len(tm.ListSubtasks(selectedProjectID, selectedTaskID))-1 {
				selectedSubtaskIndex++
				subtaskList.SelectedRow = selectedSubtaskIndex

				updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(description)
			}

		case "<C-l>": // Add or edit description for the selected getTask or subtask
			if !inSubtaskMode && len(tm.ListTasks(selectedProjectID)) > 0 {
				typingMode = true

				inputState = "description"
				inputBuffer.Reset()
				selectedTask := tm.ListTasks(selectedProjectID)[selectedTaskIndex]
				taskInput.Title = "Edit getTask description"
				taskInput.Text = selectedTask.Description
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(selectedProjectID, tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID)) > 0 {
				typingMode = true

				inputState = "description"
				inputBuffer.Reset()
				selectedSubtask := tm.ListSubtasks(selectedProjectID, tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID)[selectedSubtaskIndex]
				taskInput.Title = "Edit subtask description"
				taskInput.Text = selectedSubtask.Description
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			}

		case "<C-k>", "<Up>": // Move selection up
			if inProjectMode && len(tm.ListProjects()) > 0 && selectedProjectIndex > 0 {
				selectedProjectIndex--
				projectList.SelectedRow = selectedProjectIndex
				selectedProjectID = tm.ListProjects()[selectedProjectIndex].ID
				// Reset getTask and subtask indices and IDs
				selectedTaskIndex = 0
				selectedTaskID = -1
				selectedSubtaskIndex = 0
				// Update UI elements
				termui.Render(taskList, subtaskList, description, projectList)
				updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(description)
			} else if !inProjectMode && !inSubtaskMode && len(tm.ListTasks(selectedProjectID)) > 0 && selectedTaskIndex > 0 {
				selectedTaskIndex--
				taskList.SelectedRow = selectedTaskIndex
				selectedTaskID = tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID

				// Reset subtask index and ID
				selectedSubtaskIndex = 0

				updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
				updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
				updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(taskList, subtaskList, description)
			} else if inSubtaskMode && len(tm.ListSubtasks(selectedProjectID, selectedTaskID)) > 0 && selectedSubtaskIndex > 0 {
				selectedSubtaskIndex--
				subtaskList.SelectedRow = selectedSubtaskIndex

				updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
				termui.Render(description)
			}

		case "<C-o>":
			showHelpModal()
			termui.Clear() // Clear the screen after closing the help modal
			termui.Render(grid)

		case "<C-t>": // Switch to getTask mode
			inProjectMode = false
			inSubtaskMode = false
			// Optionally reset selected indices
			selectedTaskIndex = 0
			selectedTaskID = 0

			selectedSubtaskIndex = 0
			taskList.SelectedRow = selectedTaskIndex
			// Update UI elements
			updateTaskList(taskList, tm, selectedProjectID)
			updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
			updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
			updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
			termui.Render(taskList, subtaskList, taskInput)

		case "<C-g>": // Toggle getTask completion (mark as done/undone)
			if inSubtaskMode && len(tm.ListSubtasks(selectedProjectID, tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID)) > 0 {
				// Subtask mode: Toggle the selected subtask's completion status
				tm.ToggleSubtaskComplete(
					selectedProjectID,
					tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID,
					tm.ListSubtasks(selectedProjectID, selectedTaskID)[selectedSubtaskIndex].ID,
				)
				updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
				updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
				termui.Render(subtaskList)
			} else if !inSubtaskMode && len(tm.ListTasks(selectedProjectID)) > 0 {
				// Task mode: Toggle the selected getTask's completion status
				tm.ToggleComplete(
					selectedProjectID, tm.ListTasks(selectedProjectID)[selectedTaskIndex].ID,
				)
				updateTaskList(taskList, tm, selectedProjectID)
				termui.Render(taskList)
			}

		case "<C-m>": // Assign to someone
			if !inSubtaskMode && len(tm.ListTasks(selectedProjectID)) > 0 {
				typingMode = true

				inputState = "assign"
				inputBuffer.Reset()
				selectedTask := tm.ListTasks(selectedProjectID)[selectedTaskIndex]
				taskInput.Title = "Assign getTask to"
				taskInput.Text = selectedTask.AssignedTo
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
			} else if inSubtaskMode && len(tm.ListSubtasks(selectedProjectID, selectedTaskID)) > 0 {
				typingMode = true

				inputState = "assign"
				inputBuffer.Reset()
				selectedSubtask := tm.ListSubtasks(selectedProjectID, selectedTaskID)[selectedSubtaskIndex]
				taskInput.Title = "Assign subtask to"
				taskInput.Text = selectedSubtask.AssignedTo
				inputBuffer.WriteString(taskInput.Text)
				termui.Render(taskInput)
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
		// After any update
		updateTaskList(taskList, tm, selectedProjectID)
		if len(tm.ListTasks(selectedProjectID)) > 0 {
			updateSubtaskList(subtaskList, tm, selectedProjectID, selectedTaskID)
			updateGauge(gauge, tm, selectedProjectID, selectedTaskID)
		}
		updatePieChart(pieChart, tm, selectedProjectID)
		updateDescription(description, tm, selectedProjectID, selectedTaskIndex, selectedSubtaskIndex, inSubtaskMode)
		termWidth, termHeight = termui.TerminalDimensions()
		grid.SetRect(0, 0, termWidth, termHeight)
		termui.Clear()
		termui.Render(grid)

	}
}

// updateTaskList updates the task list with the current tasks
func updateTaskList(taskList *widgets.List, tm *task.TaskManager, projectID int) {
	if projectID == -1 {
		taskList.Rows = []string{"No tasks available"}
		return
	}
	tasks := tm.ListTasks(projectID)
	rows := []string{}
	for _, task := range tasks {
		status := "[ ]"
		if task.Complete {
			status = "[x]"
		}
		rows = append(rows, fmt.Sprintf("%d. %s %s (Assigned to: %s)", task.ID, status, task.Title, task.AssignedTo))
	}
	taskList.Rows = rows

	if len(rows) == 0 {
		taskList.Rows = []string{"No tasks available"}
	}
}

func updateProjectList(projectList *widgets.List, tm *task.TaskManager, selectedProjectIndex int) {
	projects := tm.ListProjects()
	rows := []string{}
	for _, project := range projects {
		rows = append(rows, fmt.Sprintf("%d. %s", project.ID, project.Name))
	}
	if len(rows) == 0 {
		projectList.Rows = []string{"No projects available"}
		projectList.SelectedRow = -1
	} else {
		projectList.Rows = rows
		if selectedProjectIndex >= 0 && selectedProjectIndex < len(rows) {
			projectList.SelectedRow = selectedProjectIndex
		} else {
			projectList.SelectedRow = 0
		}
	}
}

// updateSubtaskList updates the subtask list with the current subtasks for a specific task
func updateSubtaskList(subtaskList *widgets.List, tm *task.TaskManager, projectID int, taskID int) {
	if taskID == -1 {
		subtaskList.Rows = []string{"No subtasks available"}
		return
	}
	subtasks := tm.ListSubtasks(projectID, taskID)
	rows := []string{}
	for _, subtask := range subtasks {
		status := "[ ]"
		if subtask.Complete {
			status = "[x]"
		}
		rows = append(rows, fmt.Sprintf("%d. %s %s (Assigned to: %s)", subtask.ID, status, subtask.Title, subtask.AssignedTo))
	}
	subtaskList.Rows = rows

	if len(rows) == 0 {
		subtaskList.Rows = []string{"No subtasks available"}
	}
}

// updateBarChart updates the bar chart with current task completion statistics
func updateBarChart(barChart *widgets.BarChart, tm *task.TaskManager, projectID int) {
	if projectID == -1 {
		barChart.Data = []float64{0, 0}
		return
	}
	tasks := tm.ListTasks(projectID)
	totalTasks := len(tasks)
	if totalTasks == 0 {
		barChart.Data = []float64{0, 0}
		return
	}
	completedTasks := 0
	for _, task := range tasks {
		if task.Complete {
			completedTasks++
		}
	}
	pendingTasks := totalTasks - completedTasks
	barChart.Data = []float64{float64(completedTasks), float64(pendingTasks)}
}

// calculateCompletionPercentage calculates the percentage of completed tasks
func calculateCompletionPercentage(tm *task.TaskManager, projectID int) int {
	totalTasks := len(tm.ListTasks(projectID))
	if totalTasks == 0 {
		return 0
	}
	completedTasks := 0
	for _, task := range tm.ListTasks(projectID) {
		if task.Complete {
			completedTasks++
		}
	}
	percentage := (completedTasks * 100) / totalTasks
	return percentage
}

// updateGauge updates the gauge with the percentage of completed subtasks for the selected task
func updateGauge(gauge *widgets.Gauge, tm *task.TaskManager, projectID int, taskID int) {
	if taskID == -1 {
		// No task selected
		gauge.Percent = 0
		gauge.Label = "No Task Selected"
		gauge.BarColor = termui.ColorYellow
		gauge.LabelStyle = termui.NewStyle(termui.ColorBlack, termui.ColorYellow)
		return
	}

	// Get the subtasks for the selected task
	subtasks := tm.ListSubtasks(projectID, taskID)
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

func updatePieChart(pieChart *widgets.PieChart, tm *task.TaskManager, projectID int) {
	totalTasks := len(tm.ListTasks(projectID))
	if totalTasks == 0 {
		pieChart.Data = []float64{0, 100}
		return
	}
	completedTasks := 0
	for _, task := range tm.ListTasks(projectID) {
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

func updateDescription(description *widgets.Paragraph, tm *task.TaskManager, projectID int, taskIndex int, subtaskIndex int, inSubtaskMode bool) {
	if inSubtaskMode {
		if taskIndex >= 0 && subtaskIndex >= 0 && len(tm.ListTasks(projectID)) > taskIndex {
			task := tm.ListTasks(projectID)[taskIndex]
			if len(task.Subtasks) > subtaskIndex {
				subtask := task.Subtasks[subtaskIndex]
				description.Text = subtask.Description
			}
		} else {
			description.Text = "No Subtask Selected"
		}
	} else {
		if taskIndex >= 0 && len(tm.ListTasks(projectID)) > taskIndex {
			task := tm.ListTasks(projectID)[taskIndex]
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
		"Ctrl+t: Toggle task or subtask completion\n" +
		"Ctrl+b: Show task tree\n"
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

func buildTaskTreeRepresentation(tm *task.TaskManager) string {
	var sb strings.Builder

	projects := tm.ListProjects()
	for _, project := range projects {
		sb.WriteString(fmt.Sprintf("%d. %s\n", project.ID, project.Name))
		for _, task := range project.Tasks {
			taskStatus := "[ ]"
			if task.Complete {
				taskStatus = "[x]"
			}
			sb.WriteString(fmt.Sprintf("    %s %d. %s\n", taskStatus, task.ID, task.Title))
			for _, subtask := range task.Subtasks {
				subStatus := "[ ]"
				if subtask.Complete {
					subStatus = "[x]"
				}
				sb.WriteString(fmt.Sprintf("        %s %d. %s\n", subStatus, subtask.ID, subtask.Title))
			}
		}
	}

	if sb.Len() == 0 {
		sb.WriteString("No projects available")
	}

	return sb.String()
}

func showTaskTreeModal(tm *task.TaskManager) {
	// Create a new Paragraph widget to display the tree
	treeText := widgets.NewParagraph()
	treeText.Title = "Task Tree"
	treeText.WrapText = true

	// Build the tree representation
	treeRepresentation := buildTaskTreeRepresentation(tm)

	treeText.Text = treeRepresentation

	// Set the size and position of the modal
	termWidth, termHeight := termui.TerminalDimensions()
	treeText.SetRect(termWidth/4, termHeight/4, 3*termWidth/4, 3*termHeight/4)
	termui.Render(treeText)

	// Wait for a key event to close the modal
	for e := range termui.PollEvents() {
		if e.Type == termui.KeyboardEvent {
			break
		}
	}
}
