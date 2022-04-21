package workspace

import "sort"

type Workspace struct {
	Folder   string `json:"folder"`
	Name     string
	Basename string
	Time     int64
}

type WorkspaceCollection []Workspace

func (workspaces WorkspaceCollection) Unique() WorkspaceCollection {
	var out WorkspaceCollection

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

type ByTime WorkspaceCollection

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Less(i, j int) bool { return a[i].Time > a[j].Time }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByName WorkspaceCollection

func (a ByName) Len() int           { return len(a) }
func (a ByName) Less(i, j int) bool { return a[i].Basename < a[j].Basename }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPath WorkspaceCollection

func (a ByPath) Len() int           { return len(a) }
func (a ByPath) Less(i, j int) bool { return a[i].Folder < a[j].Folder }
func (a ByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (workspaces WorkspaceCollection) Sort(sortOption string) {
	switch sortOption {
	case "name":
		sort.Sort(ByName(workspaces))
	case "path":
		sort.Sort(ByPath(workspaces))
	case "time":
		sort.Sort(ByTime(workspaces))
	}
}
