package cli

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	Name string `json:"name"`
	Url  string `json:"url"`
}

func GenerateCmd() *cobra.Command {
	var lang, dir, tag string
	cmd :=
		&cobra.Command{
			Use:          "generate [resource]",
			Aliases:      []string{"gen"},
			Short:        "Generate scaffolding for a new Extism resource, e.g. 'plugin'",
			Example:      "generate plugin",
			ValidArgs:    []string{"plugin"},
			SilenceUsage: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 || args[0] == "" {
					cmd.Help()
					return fmt.Errorf("not enough arguments, expected resource name (plugin, ...)")
				}
				switch args[0] {
				case "plugin":
					return generatePlugin(lang, dir, tag)
				default:
					cmd.Help()
					return fmt.Errorf("unsupported resource: '%s'", args[0])
				}
			},
		}

	flags := cmd.Flags()
	flags.StringVarP(&lang, "lang", "l", "", "[optional] The name of the PDK language to generate a plugin scaffold, e.g. 'rust'")
	flags.StringVarP(&dir, "output", "o", ".", "The path to an output directory where resource scaffolding will be generated")
	flags.StringVarP(&tag, "tag", "t", "main", "A tag to clone from the template repository")

	return cmd
}

func generatePlugin(lang string, dir, tag string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("missing `git`, please install before executing this command")
	}
	var templates []pdkTemplate
	data := templatesData
	res, err := http.Get("https://raw.githubusercontent.com/extism/cli/main/pdk-templates.json")
	if err == nil && res.StatusCode == 200 {
		t, err := io.ReadAll(res.Body)
		if err == nil {
			data = t
		}
		defer res.Body.Close()
	} else {
		Log("Unable to fetch PDK templates, falling back to local list")
	}
	err = json.Unmarshal(data, &templates)
	if err != nil {
		fmt.Println(string(data))
		return err
	}

	lang = strings.ToLower(lang)
	if lang != "" {
		var match bool
		var pdk pdkTemplate
		for _, tmpl := range templates {
			if strings.ToLower(tmpl.Name) == lang {
				match = true
				pdk = tmpl
				break
			}
		}

		if match {
			if err := cloneTemplate(pdk, dir, tag); err != nil {
				return err
			}
			return nil
		}
	} else {
		pdk := pickPdk(templates)
		return cloneTemplate(pdk, dir, tag)
	}

	var langs []string
	for _, tmpl := range templates {
		langs = append(langs, tmpl.Name)
	}

	return fmt.Errorf("unsupported template: '%s'. Supported templates are: %s", lang, strings.Join(langs, ", "))
}

func runCmdInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.Run()
}

func cloneTemplate(pdk pdkTemplate, dir, tag string) error {
	if err := runCmdInDir("", "git", "clone", "--depth=1", pdk.Url, "--branch", tag, dir); err != nil {
		return err
	}

	// initialize submodules if any
	if _, err := os.Stat(filepath.Join(dir, ".gitmodules")); err == nil {
		if err := runCmdInDir(dir, "git", "submodule", "update", "--init", "--recursive"); err != nil {
			return err
		}
	}

	// recursively check that parents are not a git repository, create an orphan branch & commit, cleanup
	// otherwise, remove the git repository and assume this should be a plain directory within the parent
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if hasGitRepoInParents(absDir, 100) {
		if err := os.RemoveAll(filepath.Join(dir, ".git")); err != nil {
			return err
		}
	} else {
		if err := runCmdInDir(dir, "git", "checkout", "--orphan", "extism-init", "main"); err != nil {
			return err
		}

		if err := runCmdInDir(dir, "git", "commit", "-am", "init: extism"); err != nil {
			return err
		}

		if err := runCmdInDir(dir, "git", "branch", "-M", "extism-init", "main"); err != nil {
			return err
		}

		if err := runCmdInDir(dir, "git", "remote", "remove", "origin"); err != nil {
			return err
		}
	}

	fmt.Println("Generated", pdk.Name, "plugin scaffold at", dir)

	return nil
}

func hasGitRepoInParents(dir string, depth int) bool {
	parent := filepath.Dir(dir)
	if depth == 0 || parent == "" || parent == "." || parent == dir {
		return false
	}
	fi, err := os.Stat(filepath.Join(parent, ".git"))
	if err != nil && os.IsNotExist(err) {
		return hasGitRepoInParents(parent, depth-1)

	}
	if fi.IsDir() {
		// found a git repository
		return true
	}

	return hasGitRepoInParents(parent, depth-1)
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
		return quitTextStyle.Render(fmt.Sprintf("Generating scaffold for plugin using %s...", m.choice))
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
		pdkMap[pdk.Name] = pdk
		items = append(items, item(pdk.Name))
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
