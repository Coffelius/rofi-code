package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/Wing924/shellwords"
	"github.com/akamensky/argparse"
)

type Workspace struct {
	Folder   string `json:"folder"`
	Name     string
	Basename string
	Time     int64
}

type Storage struct {
	OpenedPathsList struct {
		Workspaces3 []string `json:"workspaces3"`
	} `json:"openedPathsList"`

	LastKnownMenuBarData struct {
		Menus struct {
			File struct {
				Items []struct {
					ID      string `json:"id"`
					Label   string `json:"label"`
					Submenu struct {
						Items []struct {
							ID  string `json:"id"`
							URI struct {
								External string `json:"external"`
								Path     string `json:"path"`
							} `json:"uri"`
						} `json:"items"`
					} `json:"submenu"`
				} `json:"items"`
			} `json:"File"`
		} `json:"menus"`
	} `json:"lastKnownMenuBarData"`
}

func newWorkspaceFromPath(j string) (*Workspace, error) {
	var workspace Workspace
	var err error

	workspace.Folder = j
	if len(workspace.Folder) <= 8 || !strings.HasPrefix(workspace.Folder, "file:///") {
		return nil, errors.New("Wrong folder " + j)
	}

	workspace.Folder = workspace.Folder[7:]
	workspace.Name = workspace.Folder
	workspace.Basename = path.Base(workspace.Name)

	_, err = os.Stat(workspace.Folder)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func getWorkspacesFromStorage(s string) []Workspace {
	var workspaces []Workspace
	var storage Storage
	var err error

	err = loadJSON(path.Join(s, "storage.json"), &storage)
	if err != nil {
		return workspaces
	}

	for _, j := range storage.OpenedPathsList.Workspaces3 {
		var workspace *Workspace
		var err error

		workspace, err = newWorkspaceFromPath(j)
		if err != nil {
			continue
		}
		if !*fullpath {
			workspace.Name = contractTilde(workspace.Name)
		}
		workspaces = append(workspaces, *workspace)
	}

	for _, item := range storage.LastKnownMenuBarData.Menus.File.Items {
		if item.ID != "openRecentFolder" || len(item.Submenu.Items) == 0 {
			continue
		}
		for _, submenu := range item.Submenu.Items {
			var workspace *Workspace
			workspace, err = newWorkspaceFromPath(submenu.URI.External)
			if err != nil {
				continue
			}
			if !*fullpath {
				workspace.Name = contractTilde(workspace.Name)
			}
			workspaces = append(workspaces, *workspace)
		}
	}
	return workspaces
}

func loadJSON(filename string, v interface{}) error {
	var err error = nil

	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, v)
	return err
}

func getWorkspace(s string) (*Workspace, error) {
	var modifiedtime int64 = 0
	var err error
	var file os.FileInfo

	if *sortOption == "time" {
		file, err = os.Stat(path.Join(s, "state.vscdb"))

		if err != nil {
			return nil, err
		}
		modifiedtime = file.ModTime().Unix()
	}

	var workspace Workspace

	err = loadJSON(path.Join(s, "workspace.json"), &workspace)
	if err != nil {
		return nil, err
	}

	workspace.Time = modifiedtime

	if len(workspace.Folder) <= 8 || !strings.HasPrefix(workspace.Folder, "file:///") {
		return nil, errors.New("Protocol unknown in url " + workspace.Folder)
	}

	workspace.Folder = workspace.Folder[7:]
	workspace.Name = workspace.Folder
	workspace.Basename = path.Base(workspace.Name)

	file, err = os.Stat(workspace.Folder)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func getUniqueWorkspaces(workspaces []Workspace) []Workspace {
	var out []Workspace

	if len(workspaces) < 2 {
		return workspaces
	}

	for i := len(workspaces) - 1; i > 0; i-- {
		duplicated := false
		for j := i - 1; j >= 0; j-- {
			if workspaces[i].Name == workspaces[j].Name {
				duplicated = true
				break
			}
		}
		if !duplicated {
			out = append(workspaces[i:i+1], out...)
		}
	}
	out = append(workspaces[0:1], out...)
	return out
}

type byTime []Workspace

func (a byTime) Len() int           { return len(a) }
func (a byTime) Less(i, j int) bool { return a[i].Time > a[j].Time }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type byName []Workspace

func (a byName) Len() int           { return len(a) }
func (a byName) Less(i, j int) bool { return a[i].Basename < a[j].Basename }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type byPath []Workspace

func (a byPath) Len() int           { return len(a) }
func (a byPath) Less(i, j int) bool { return a[i].Folder < a[j].Folder }
func (a byPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

var homeDir string
var rofiCmd *string
var codeCmd *string
var sortOption *string
var fullpath *bool
var basePath string

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	homeDir = usr.HomeDir
}

func expandTilde(s string) string {
	// replace tilde by homedir
	if strings.HasPrefix(s, "~/") {
		return filepath.Join(homeDir, (s)[2:])
	}
	return s
}

func contractTilde(s string) string {
	if strings.Index(s, homeDir) == 0 {
		return "~" + s[len(homeDir):]
	}
	return s
}

func runRofi(workspaces []Workspace) {
	args, err := shellwords.Split(*rofiCmd)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(args[0], args[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		for _, s := range workspaces {
			io.WriteString(stdin, s.Name+"\n")
		}
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	args, err = shellwords.Split(*codeCmd)
	if err != nil {
		log.Fatal(err)
	}
	selectedItem := strings.TrimSpace(string(out))
	selectedItem = expandTilde(selectedItem)
	args = append(args, selectedItem)

	executable, err := exec.LookPath(args[0])
	if err != nil {
		log.Fatal(err)
	}
	err = syscall.Exec(executable, args, os.Environ())
	if err != nil {
		log.Fatal(err)
	}
}

func sortWorkspaces(workspaces []Workspace) {
	switch {
	case *sortOption == "name":
		sort.Sort(byName(workspaces))
	case *sortOption == "path":
		sort.Sort(byPath(workspaces))
	case *sortOption == "time":
		sort.Sort(byTime(workspaces))
	}
}

func getWorkspacesFromUserWorkspace(basePath string) []Workspace {
	var workspaces []Workspace

	paths, err := filepath.Glob(path.Join(basePath, "User/workspaceStorage/*"))
	if err != nil {
		return workspaces
	}

	for _, s := range paths {
		workspace, err := getWorkspace(s)
		if err == nil {
			// replace home path by tilde
			if !*fullpath {
				workspace.Name = contractTilde(workspace.Name)
			}

			workspaces = append(workspaces, *workspace)
		}
	}

	return workspaces
}

func detectCodeExecutablePath() (*string, error) {
	var binaries = []string{"codium", "code"}
	for _, binary := range binaries {
		path, err := exec.LookPath(binary)
		if err == nil {
			return &path, nil
		}
	}
	return nil, errors.New("No Code or Codium automatically found. You will have to specify an alternative editor executable using the --code option")
}

func main() {
	var err error
	// Create new parser object
	parser := argparse.NewParser("rofi-code", "Use rofi to quickly open a VSCode or Codium workspace")
	// Create string flag
	s := parser.String("d", "dir", &argparse.Options{Required: false, Help: "Comma separated paths to the config directories", Default: "~/.config/VSCodium,~/.config/Code,~/.config/Code\\ -\\ OSS"})
	sortOption = parser.Selector("s", "sort", []string{"time", "path", "name"}, &argparse.Options{Required: false, Help: "How the workspaces should be sorted", Default: "time"})
	fullpath = parser.Flag("f", "full", &argparse.Options{Required: false, Help: "Show the full path instead of collapsing the home directory to a tilde", Default: false})
	var output *bool = parser.Flag("o", "output", &argparse.Options{Required: false, Help: "Just prints the workspaces to stdout and exit", Default: false})
	rofiCmd = parser.String("r", "rofi", &argparse.Options{Required: false, Help: "Command line to launch rofi", Default: "rofi -dmenu -p \"Open workspace\" -no-custom"})
	codeCmd = parser.String("c", "code", &argparse.Options{Required: false, Help: "Command line to launch the editor. It will try to detect codium or code", Default: nil})

	err = parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	if *codeCmd == "" {
		codeCmd, err = detectCodeExecutablePath()
		if err != nil {
			log.Fatal(err)
		}
	}

	var basePaths []string

	basePaths = strings.Split(*s, ",")
	for i := range basePaths {
		basePaths[i] = expandTilde(strings.TrimSpace(basePaths[i]))
	}

	var workspaces []Workspace

	for _, basePath := range basePaths {

		workspaces = append(workspaces, getWorkspacesFromUserWorkspace(basePath)...)
	}

	if *sortOption == "time" {
		sortWorkspaces(workspaces)
	}

	var workspacesFromJSON []Workspace

	for _, basePath := range basePaths {
		workspacesFromJSON = append(workspacesFromJSON, getWorkspacesFromStorage(basePath)...)
	}

	workspaces = append(workspacesFromJSON, workspaces...)
	workspaces = getUniqueWorkspaces(workspaces)

	if len(workspaces) == 0 {
		log.Fatal("Not found any workspace")
	}
	if *sortOption != "time" {
		sortWorkspaces(workspaces)
	}

	if *output {
		for _, s := range workspaces {
			fmt.Println(s.Name)
		}
		return
	}

	runRofi(workspaces)
}
