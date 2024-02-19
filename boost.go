package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

var options = huh.NewOptions("Start site", "Stop Site", "Create Site", "Delete Site & Files", "Restart Site", "Fix Permissions", "Add SSH Key", "Container Shell", "Fail2ban Status", "Unban IP", "Whitelist IP", "Prune Docker Images", "MariaDB Upgrade", "Change Site Domain", "DB Search Replace")

var siteChooseOptions = []string{"Stop Site", "Restart Site", "Delete Site & Files", "Fix Permissions", "Container Shell", "Change Site Domain", "DB Search Replace"}

func checkForUpdate() {
	spinner.New().Title("Checking for update...").Action(func() {
		time.Sleep(500_000_000)
	}).Run()
}

func boost() {
	checkForUpdate()

	var chosenOption string
	chosenSite := ""

	form := huh.NewForm(

		// Ask the user what they want to do.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What do you want to do?").
				Options(
					options...,
				).
				Value(&chosenOption), // store the chosen option in the "burger" variable
		),

		// Ask the user for a site.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which site?").
				Options(
					huh.NewOptions(GetDirectoriesInPath("/home/hank")...)...,
				).
				Value(&chosenSite),
		).WithHideFunc(func() bool {
			for _, option := range siteChooseOptions {
				if chosenOption == option {
					return false
				}
			}
			return true
		}),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	runSelection(chosenOption, chosenSite)
}

func deleteSite(chosenSite string) {
	var confirm bool
	huh.NewConfirm().
		Title("Are you sure?").
		Description("Seriously, this will completely delete " + chosenSite + ".").
		Affirmative("Yes").
		Negative("No!").
		Value(&confirm).
		Run()

	if confirm {
		exec.Command("docker", "compose", "-f", "/home/"+os.Getenv("USER")+"/sites/"+chosenSite+"/docker-compose.yml", "stop").Run()
		exec.Command("docker", "compose", "-f", "/home/"+os.Getenv("USER")+"/sites/"+chosenSite+"/docker-compose.yml", "rm").Run()
		exec.Command("sudo", "rm", "-r", "/home/"+os.Getenv("USER")+"/sites/"+chosenSite).Run()
	}

	fmt.Println(
		lipgloss.NewStyle().
			Width(40).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Render("lakjdflakjdlkfj"),
	)

}

func runSelection(selection string, chosenSite string) {
	switch selection {
	case "Start site":
		fmt.Println("Starting site")
		// docker compose -f "/home/$USER/sites/$sitename/docker-compose.yml" up -d
		cmd := exec.Command("docker", "compose", "-f", "/home/"+os.Getenv("USER")+"/sites/"+chosenSite+"/docker-compose.yml", "up", "-d")
		output, _ := cmd.CombinedOutput()
		fmt.Println(string(output))

	case "Stop Site":
		fmt.Println("Stopping site")

	case "Create Site":
		fmt.Println("Creating site")

	case "Delete Site & Files":
		deleteSite(chosenSite)

	case "Restart Site":
		fmt.Println("Restarting site")

	case "Fix Permissions":
		fmt.Println("Fixing permissions")

	case "Add SSH Key":
		addSSHKey()

	case "Container Shell":
		fmt.Println("Container shell")

	case "Fail2ban Status":
		fmt.Println("Fail2ban status")

	case "Unban IP":
		fmt.Println("Unbanning IP")

	case "Whitelist IP":
		fmt.Println("Whitelisting IP")

	case "Prune Docker Images":
		fmt.Println("Pruning Docker images")

	case "MariaDB Upgrade":
		fmt.Println("Upgrading MariaDB")

	case "Change Site Domain":
		fmt.Println("Changing site domain")

	case "DB Search Replace":
		fmt.Println("DB search and replace")

	default:
		fmt.Println("Invalid selection")
	}
}

func main() {
	boost()
}

func addSSHKey() {
	var key string
	huh.NewText().
		Title("Enter public key").
		Description("Look in ~/.ssh - file ends in .pub").
		Value(&key).
		Run()

	fmt.Println(
		lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 3).
			Render("Added SSH key. Have a nice day!"),
	)
}
