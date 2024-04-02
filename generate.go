package cli

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

//go:embed pdk-templates.json
var templatesData []byte

type pdkTemplate struct {
	Pdk string `json:"pdk"`
	Url string `json:"url"`
}

func GenerateCmd() *cobra.Command {
	var lang, dir string
	cmd :=
		&cobra.Command{
			Use:       "generate [resource]",
			Aliases:   []string{"gen"},
			Short:     "Generate scaffolding for a new Extism resource, e.g. 'plugin'",
			Example:   "generate plugin",
			ValidArgs: []string{"plugin"},
			Args:      cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				switch args[0] {
				case "plugin":
					return generatePlugin(lang, dir)
				default:
					return fmt.Errorf("unsupported resource: '%s'", args[0])
				}
			},
		}

	flags := cmd.Flags()
	flags.StringVarP(&lang, "lang", "l", "", "[optional] The name of the PDK language to generate a plugin scaffold, e.g. 'rust'")
	flags.StringVarP(&dir, "output", "o", ".", "The path to an output directory where resource scaffolding will be generated")

	return cmd
}

func generatePlugin(lang string, dir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("missing `git`, please install before executing this command")
	}
	var templates []pdkTemplate
	err := json.Unmarshal(templatesData, &templates)
	if err != nil {
		return err
	}

	lang = strings.ToLower(lang)
	if lang != "" {
		var match bool
		var pdk pdkTemplate
		for _, tmpl := range templates {
			if strings.ToLower(tmpl.Pdk) == lang {
				match = true
				pdk = tmpl
				break
			}
		}

		if match {
			if err := cloneTemplate(pdk, dir); err != nil {
				return err
			}
			return nil
		}
	} else {
		pdk := pickPdk(templates)
		return cloneTemplate(pdk, dir)
	}

	return nil
}

func cloneTemplate(pdk pdkTemplate, dir string) error {
	cloneCmd := exec.Command("git", "clone", pdk.Url, dir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Start(); err != nil {
		return err
	}
	if err := cloneCmd.Wait(); err != nil {
		return err
	}

	if err := os.RemoveAll(filepath.Join(dir, ".git")); err != nil {
		return err
	}

	initCmd := exec.Command("git", "init", dir)
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Start(); err != nil {
		return err
	}
	if err := initCmd.Wait(); err != nil {
		return err
	}

	fmt.Println("Generated", pdk.Pdk, "plugin scaffold at", dir)

	return nil
}

const listHeight = 15

var (
	titleStyle        = lipgloss.NewStyle().Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("Generating scaffold for plugin using %s-pdk...", m.choice))
	}
	if m.quitting {
		return "Operation cancelled."
	}
	return "\n" + m.list.View()
}

func pickPdk(pdks []pdkTemplate) pdkTemplate {
	pdkMap := make(map[string]pdkTemplate, len(pdks))
	var items []list.Item
	for _, pdk := range pdks {
		pdkMap[pdk.Pdk] = pdk
		items = append(items, item(pdk.Pdk))
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a PDK language to use for your plugin:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := &model{list: l}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m.choice == "" {
		os.Exit(0)
	}

	return pdkMap[m.choice]
}
