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

func iconAppName(appID string) string {
	fmt.Printf("AppID: %s\n\n", appID)
	icon := appsName[strings.ToLower(appID)]
	if icon == "" {
		return appsName["generic"]
	}

	return icon
}

func renameWorkspace(num int, oldName string, name string) {
	name = strconv.Itoa(num) + " " + name
	exec.Command("swaymsg", "rename", "workspace", oldName, "to", name).Run()
}

func main() {

	type Container struct {
		AppID string `json:"app_id"`
		// Name  string `json:"name"`
		ID int `json:"id"`
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

		if change == "focus" {
			num, name := getFocusIDWorkspace(window.Con.ID)
			icon := iconAppName(window.Con.AppID)
			renameWorkspace(num, name, icon)
		} else if change == "close" {
			num, name := getFocusedWorkspace()
			renameWorkspace(num, name, "")
		}
	}
}
