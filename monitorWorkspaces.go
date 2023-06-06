// change workspace number to number and icon
// according to the active app running on the workspace
// pass icon file path as parameter
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"i3status/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	appsName = map[string]string{}
)

func getFocusedWorkspace() (int, string) {
	for _, workspace := range utils.SwayMsgWorkspaces() {
		if workspace.Focused {
			return workspace.Num, workspace.Name
		}
	}

	return 1, "1"
}

func getFocusIDWorkspace(id int) (int, string) {
	workspaces := utils.SwayMsgWorkspaces()
	for _, workspace := range workspaces {
		for _, wid := range workspace.Focus {
			if wid == id {
				return workspace.Num, workspace.Name
			}
		}
	}

	return getFocusedWorkspace()
}

func iconAppName(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), "_", " ")

	bname := appsName[name]
	if bname == "" {
		before, _, found := strings.Cut(name, " ")
		if !found || appsName[before] == "" {
			return appsName["generic"]
		}

		return appsName[before]
	}
	return bname
}

func renameWorkspace(num int, oldName string, currName string) {
	if currName != "" {
		currName = iconAppName(currName)
	}
	currName = strconv.Itoa(num) + " " + currName
	exec.Command("swaymsg", "rename", "workspace", oldName, "to", currName).Run()
}

func main() {

	type Container struct {
		AppID string `json:"app_id"`
		Name  string `json:"name"`
		ID    int    `json:"id"`
	}

	type Window struct {
		Change string    `json:"change"`
		Con    Container `json:"container"`
	}

	// read apps icon file from argument
	if len(os.Args) < 2 {
		panic("Path for icon file was not passsed as argument")
	}

	appsIconFile, err := os.ReadFile(os.Args[1])
	if err != nil {
		// error reading file icon, set generic icon only
		appsName["generic"] = "\uf22d"
	}
	json.Unmarshal(appsIconFile, &appsName)

	cmd := exec.Command("swaymsg", "-rt", "subscribe", "-m", "[\"window\"]")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic("ERROR: Swaymsg was unable to subscribe to window event.")
	}

	scanner := bufio.NewScanner(stdout)
	cmd.Start()

	for scanner.Scan() {
		event := scanner.Bytes()

		var window Window

		json.Unmarshal(event, &window)

		change := window.Change

		// TODO: check if I need to add the "floating" event
		fmt.Printf("Change event: %s\n", change)
		// if change == "move" || change == "new" || change == "focus" || change == "title" || change == "floating" {
		if change == "focus" {
			num, name := getFocusIDWorkspace(window.Con.ID)
			renameWorkspace(num, name, window.Con.AppID)
		} else if change == "close" {
			num, name := getFocusedWorkspace()
			renameWorkspace(num, name, "")
		}
	}
}
