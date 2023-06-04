// change workspace number to number and icon
// according to the active app running on the workspace
// get the icon from $HOME/.config/i3status/apps_icon.json
package main

import (
	"bufio"
	"encoding/json"
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
	for _, workspace := range utils.SwayMsgWorkspaces() {
		for _, wid := range workspace.Focus {
			if wid == id {
				return workspace.Num, workspace.Name
			}
		}
	}

	return 1, "1"
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

	// read apps icon file
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic("ERROR: Unable to get Home user directory.")
	}

	appsIconFile, err := os.ReadFile(userHome + "/.config/i3status/apps_icon.json")
	if err != nil {
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
		if change == "move" || change == "new" || change == "title" || change == "focus" {
			num, name := getFocusIDWorkspace(window.Con.ID)
			renameWorkspace(num, name, window.Con.AppID)

		} else if change == "close" {
			num, name := getFocusedWorkspace()
			renameWorkspace(num, name, "")
		}
	}
}
