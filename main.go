package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath" // ✅ added

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	vaultDir    string
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	docStyle    = lipgloss.NewStyle().Margin(1, 2)
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory", err)
	}

	vaultDir = fmt.Sprintf("%s/.tuigote", homeDir)
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	newFileInput           textinput.Model
	createFileInputVisible bool
	currentFile            *os.File
	noteTextArea           textarea.Model
	list                   list.Model
	showingList            bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+n":
			m.createFileInputVisible = true
			return m, nil

		case "ctrl+l":
			noteList := listFiles()
			m.list.SetItems(noteList)
			m.showingList = true
			return m, nil

		case "ctrl+s":

			if m.currentFile == nil {
				break
			}

			if err := m.currentFile.Truncate(0); err != nil {
				fmt.Println("can not save the file x(")
				return m, nil
			}

			if _, err := m.currentFile.Seek(0, 0); err != nil {
				fmt.Println("can not save the file x(")
				return m, nil
			}

			if _, err := m.currentFile.WriteString(m.noteTextArea.Value()); err != nil {
				fmt.Println("can not save the file x(")
				return m, nil
			}

			if err := m.currentFile.Close(); err != nil {
				fmt.Println("can not close the file.")
			}

			m.currentFile = nil
			m.noteTextArea.SetValue("")

			return m, nil

		case "enter":
			if m.currentFile != nil {
				break
			}

			if m.showingList {
				selectedItem, ok := m.list.SelectedItem().(item) // ✅ fixed shadowing

				if ok {
					// ❌ WRONG (kept)
					// filepath := fmt.Sprintf("%s|%s", vaultDir, item.title)

					filePath := filepath.Join(vaultDir, selectedItem.title) // ✅ correct

					content, err := os.ReadFile(filePath)
					if err != nil {
						log.Printf("Error reading file: %v", err)
						return m, nil
					}

					m.noteTextArea.SetValue(string(content))
					m.noteTextArea.CursorEnd() // ✅ UX improvement
					m.noteTextArea.Focus()

					f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
					if err != nil {
						log.Printf("Error opening file: %v", err)
						return m, nil
					}

					m.currentFile = f
					m.showingList = false
				}
				return m, nil
			}

			// create file
			filename := m.newFileInput.Value()
			if filename != "" {

				// ❌ WRONG (kept)
				// filepath := fmt.Sprintf("%s/%s.md", vaultDir, filename)

				filePath := filepath.Join(vaultDir, filename+".md") // ✅ correct

				if _, err := os.Stat(filePath); err == nil {
					return m, nil
				}

				f, err := os.Create(filePath)
				if err != nil {
					log.Fatalf("%v", err)
				}

				m.currentFile = f
				m.createFileInputVisible = false
				m.newFileInput.SetValue("")
			}

			return m, nil
		}
	}

	if m.createFileInputVisible {
		m.newFileInput, cmd = m.newFileInput.Update(msg)
	}

	if m.currentFile != nil {
		m.noteTextArea, cmd = m.noteTextArea.Update(msg)
	}

	if m.showingList {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("16")).
		Background(lipgloss.Color("205")).
		PaddingLeft(2).
		PaddingRight(2)

	welcome := style.Render("welcome to tuigote :x")
	help := "Ctrl+N:new file | Ctrl+S:save | Ctrl+L:list | Esc:back/save | Ctrl+C/q:quit"
	view := " "

	if m.createFileInputVisible {
		view = m.newFileInput.View()
	}

	if m.currentFile != nil {
		view = m.noteTextArea.View()
	}

	if m.showingList {
		view = m.list.View()
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, help)
}

func initializeModel() model {

	err := os.MkdirAll(vaultDir, 0750)
	if err != nil {
		log.Fatal(err)
	}

	ti := textinput.New()
	ti.Placeholder = "Name your file?"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.Cursor.Style = cursorStyle
	ti.PromptStyle = cursorStyle

	ta := textarea.New()
	ta.Placeholder = "your note...?"
	ta.ShowLineNumbers = false
	ta.Focus()

	noteList := listFiles()

	finalList := list.New(noteList, list.NewDefaultDelegate(), 0, 0)
	finalList.Title = "All Notes"
	finalList.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("16")).
		Background(lipgloss.Color("254")).
		Padding(0, 1)

	return model{
		newFileInput: ti,
		noteTextArea: ta,
		list:         finalList,
	}
}

func main() {
	p := tea.NewProgram(initializeModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error goon: %v", err)
		os.Exit(1)
	}
}

func listFiles() []list.Item {
	items := make([]list.Item, 0)

	entries, err := os.ReadDir(vaultDir)
	if err != nil {
		log.Fatal("Error reading notes")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime().Format("2006-01-02 15:04")

			items = append(items, item{
				title: entry.Name(),
				desc:  fmt.Sprintf("Modified : %s", modTime),
			})
		}
	}
	return items
}

// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/charmbracelet/bubbles/list"
// 	"github.com/charmbracelet/bubbles/textarea"
// 	"github.com/charmbracelet/bubbles/textinput"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/charmbracelet/lipgloss"
// )

// var (
// 	vaultDir    string
// 	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
// 	docStyle    = lipgloss.NewStyle().Margin(1, 2)
// )

// func init() {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		log.Fatal("Error getting home directory", err)
// 	}

// 	vaultDir = fmt.Sprintf("%s/.tuigote", homeDir)
// }

// type item struct {
// 	title, desc string
// }

// func (i item) Title() string       { return i.title }
// func (i item) Description() string { return i.desc }
// func (i item) FilterValue() string { return i.title }

// type model struct {
// 	newFileInput           textinput.Model
// 	createFileInputVisible bool
// 	currentFile            *os.File
// 	noteTextArea           textarea.Model
// 	list                   list.Model
// 	showingList            bool
// }

// func (m model) Init() tea.Cmd {
// 	return nil
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd

// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		h, v := docStyle.GetFrameSize()
// 		m.list.SetSize(msg.Width-h, msg.Height-v-5)

// 	case tea.KeyMsg:

// 		switch msg.String() {

// 		case "ctrl+c", "q":
// 			return m, tea.Quit

// 		case "ctrl+n":
// 			m.createFileInputVisible = true
// 			return m, nil

// 		case "ctrl+l":
// 			m.showingList = true
// 			return m, nil

// 		case "ctrl+s":

// 			if m.currentFile == nil {
// 				break
// 			}

// 			if err := m.currentFile.Truncate(0); err != nil {
// 				fmt.Println("can not save the file x(")
// 				return m, nil
// 			}

// 			if _, err := m.currentFile.Seek(0, 0); err != nil {
// 				fmt.Println("can not save the file x(")
// 				return m, nil
// 			}

// 			if _, err := m.currentFile.WriteString(m.noteTextArea.Value()); err != nil {
// 				fmt.Println("can not save the file x(")
// 				return m, nil
// 			}

// 			if err := m.currentFile.Close(); err != nil {
// 				fmt.Println("can not close the file.")
// 			}

// 			m.currentFile = nil
// 			m.noteTextArea.SetValue("")

// 			return m, nil

// 		case "enter":
// 			if m.currentFile != nil {
// 				break
// 			}

// 			if m.showingList {
// 				item, ok := m.list.SelectedItem().(item)

// 				if ok {
// 					filepath := fmt.Sprintf("%s|%s", vaultDir, item.title)

// 					content, err := os.ReadFile(filepath)
// 					if err != nil {
// 						log.Printf("Error reading file: %v", err)
// 						return m, nil
// 					}

// 					m.noteTextArea.SetValue(string(content))

// 					f, err := os.OpenFile(filepath, os.O_RDWR, 0644)
// 					if err != nil {
// 						log.Printf("Error reading file: %v", err)
// 						return m, nil
// 					}

// 					m.currentFile = f
// 					m.showingList = false
// 				}
// 				return m, nil
// 			}
// 			//todo: create file
// 			filename := m.newFileInput.Value()
// 			if filename != "" {
// 				filepath := fmt.Sprintf("%s/%s.md", vaultDir, filename)

// 				if _, err := os.Stat(filepath); err == nil {
// 					return m, nil
// 				}

// 				f, err := os.Create(filepath)
// 				if err != nil {
// 					log.Fatalf("%v", err)
// 				}

// 				m.currentFile = f
// 				m.createFileInputVisible = false
// 				m.newFileInput.SetValue("")
// 			}

// 			return m, nil
// 		}
// 	}

// 	if m.createFileInputVisible {
// 		m.newFileInput, cmd = m.newFileInput.Update(msg)
// 	}

// 	if m.currentFile != nil {
// 		m.noteTextArea, cmd = m.noteTextArea.Update(msg)
// 	}

// 	if m.showingList {
// 		m.list, cmd = m.list.Update(msg)
// 	}

// 	return m, cmd
// }

// func (m model) View() string {

// 	var style = lipgloss.NewStyle().
// 		Bold(true).
// 		Foreground(lipgloss.Color("16")).
// 		Background(lipgloss.Color("205")).
// 		PaddingLeft(2).
// 		PaddingRight(2)

// 	welcome := style.Render("welcome to tuigote :x")
// 	help := "Ctrl+N:new file | Ctrl+S:save | Ctrl+L:list | Esc:back/save | Ctrl+C/q:quit"
// 	view := " "
// 	if m.createFileInputVisible {
// 		view = m.newFileInput.View()
// 	}

// 	if m.currentFile != nil {
// 		view = m.noteTextArea.View()
// 	}

// 	if m.showingList {
// 		view = m.list.View()
// 	}

// 	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, help)
// }

// func initializeModel() model {

// 	err := os.MkdirAll(vaultDir, 0750)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	//initialize new file input
// 	ti := textinput.New()
// 	ti.Placeholder = "Name your file?"
// 	ti.Focus()
// 	ti.CharLimit = 156
// 	ti.Width = 20
// 	ti.Cursor.Style = cursorStyle
// 	ti.PromptStyle = cursorStyle

// 	//textarea
// 	ta := textarea.New()
// 	ta.Placeholder = "your note...?"
// 	ta.ShowLineNumbers = false
// 	ta.Focus()

// 	//list
// 	noteList := listFiles()
// 	// fmt.Println(noteList)
// 	finalList := list.New(noteList, list.NewDefaultDelegate(), 0, 0)
// 	finalList.Title = "All Notes"
// 	finalList.Styles.Title = lipgloss.NewStyle().
// 		Foreground(lipgloss.Color("16")).
// 		Background(lipgloss.Color("254")).
// 		Padding(0, 1)

// 	return model{
// 		newFileInput:           ti,
// 		createFileInputVisible: false,
// 		noteTextArea:           ta,
// 		list:                   finalList,
// 		// list:                   list.New(noteList, list.NewDefaultDelegate(), 0, 0),
// 	}
// }

// func main() {
// 	p := tea.NewProgram(initializeModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("there's been an error goon: %v", err)
// 		os.Exit(1)
// 	}
// }

// func listFiles() []list.Item {
// 	items := make([]list.Item, 0)

// 	entries, err := os.ReadDir(vaultDir)
// 	if err != nil {
// 		log.Fatal("Error reading notes")
// 	}

// 	for _, entry := range entries {
// 		if !entry.IsDir() {
// 			info, err := entry.Info()
// 			if err != nil {
// 				continue
// 			}

// 			modTime := info.ModTime().Format("2006-01-02 15:04")

// 			items = append(items, item{
// 				title: entry.Name(),
// 				desc:  fmt.Sprintf("Modified : %s", modTime),
// 			})
// 		}
// 	}
// 	return items
// }
