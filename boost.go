package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

var USER = os.Getenv("USER")

var options = huh.NewOptions("Start Site", "Stop Site", "Create Site", "Delete Site & Files", "Restart Site", "Fix Permissions", "Add SSH Key", "Container Shell", "Prune Docker Images", "MariaDB Upgrade", "Change Site Domain", "Database Search Replace", "Fail2ban Status", "Unban IP", "Whitelist IP")

var siteChooseOptions = []string{"Start Site", "Stop Site", "Restart Site", "Delete Site & Files", "Fix Permissions", "Container Shell", "Change Site Domain", "Database Search Replace"}

func checkForUpdate() {
	spinner.New().Title("Checking for update...").Action(func() {
		time.Sleep(500_000_000)
	}).Run()
}

func main() {
	exec.Command("clear").Run()
	// checkForUpdate()

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

func runSelection(selection string, chosenSite string) {
	switch selection {
	case "Start Site":
		startSite(chosenSite)
	case "Stop Site":
		stopSite(chosenSite)
	case "Create Site":
		createSite()
	case "Delete Site & Files":
		deleteSite(chosenSite)
	case "Restart Site":
		restartSite(chosenSite)
	case "Fix Permissions":
		fixPermissions(chosenSite)
	case "Add SSH Key":
		addSSHKey()
	case "Container Shell":
		containerShell(chosenSite)
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

	case "Database Search Replace":
		fmt.Println("DB search and replace")

	default:
		fmt.Println("Invalid selection")
	}
}

func printInBox(content string) {
	fmt.Println(
		lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 3).
			Render(content),
	)
}

func checkError(err error, msg string) {
	if err != nil {
		printInBox(fmt.Sprintf("Command failed with error:\n\n%s", strings.TrimSpace(msg)))
		os.Exit(1)
	}
}

func startSite(chosenSite string) {
	var err error
	var output []byte
	//spinner
	spinner.New().Title("Starting site...").Action(func() {
		cmd := exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "up", "-d")
		output, err = cmd.CombinedOutput()
	}).Run()
	checkError(err, string(output))
	fmt.Println("Site started. Have a wonderful day!")
}

func stopSite(chosenSite string) {
	var err error
	var output []byte

	spinner.New().Title("Stopping site...").Action(func() {
		// docker compose -f "/home/$CUR_USER/sites/$sitename/docker-compose.yml" stop
		cmd := exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "stop")
		output, err = cmd.CombinedOutput()
	}).Run()

	checkError(err, string(output))
	fmt.Println("Site stopped. Have a wonderful day!")
}

func restartSite(chosenSite string) {
	var err error
	var output []byte

	spinner.New().Title("Restarting site...").Action(func() {
		// docker compose -f "/home/$CUR_USER/sites/$sitename/docker-compose.yml" stop
		cmd := exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "restart")
		output, err = cmd.CombinedOutput()
	}).Run()

	checkError(err, string(output))
	fmt.Println("Site Restarted. Have a wonderful day!")
}

func createSite() {
	var sitename string
	huh.NewInput().
		Title("Enter site name").
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("site name cannot be empty")
			}
			return nil
		}).
		Value(&sitename).
		Run()

	sitename = ReplaceSpacesWithDashes(sitename)

	// spinner
	spinner.New().Title("Creating site...").Action(func() {
		time.Sleep(1_000_000_000)
	}).Run()

	db_name := ReplaceDashWithUnderscore(sitename)
	db_user := "u_" + ReplaceDashWithUnderscore(sitename)
	db_pass, err := GeneratePassword(14)
	if err != nil {
		log.Fatal(err)
	}

	var sb strings.Builder
	keyword := func(s string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(s)
	}
	fmt.Fprintf(&sb,
		"%s\n\nDatabase: %s\nUsername: %s\nPassword: %s\nServer:   %s",
		lipgloss.NewStyle().Bold(true).Render("Created "+sitename+"!"),
		keyword(db_name),
		keyword(db_user),
		keyword(db_pass),
		keyword("mariadb"),
	)

	printInBox(sb.String())
}

func fixPermissions(chosenSite string) {
	// spinner
	spinner.New().Title("Fixing permissions...").Action(func() {
		// sudo chown -R nobody: "/home/$CUR_USER/sites/$sitename/wordpress"
		cmd := exec.Command("sudo", "chown", "-R", "nobody:", "/home/"+USER+"/sites/"+chosenSite+"/wordpress")
		output, err := cmd.CombinedOutput()
		checkError(err, string(output))
		// sudo find "/home/$CUR_USER/sites/$sitename" -type d -exec chmod 755 {} +
		cmd = exec.Command("sudo", "find", "/home/"+USER+"/sites/"+chosenSite, "-type", "d", "-exec", "chmod", "755", "{}", "+")
		output, err = cmd.CombinedOutput()
		checkError(err, string(output))
		// sudo find "/home/$CUR_USER/sites/$sitename/wordpress" -type f -exec chmod 644 {} +
		cmd = exec.Command("sudo", "find", "/home/"+USER+"/sites/"+chosenSite+"/wordpress", "-type", "f", "-exec", "chmod", "644", "{}", "+")
		output, err = cmd.CombinedOutput()
		checkError(err, string(output))
	}).Run()

	printInBox("Permissions fixed. Have a fantastic day!")
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
		spinner.New().Title("Deleting site...").Action(func() {
			output, err := exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "stop").CombinedOutput()
			checkError(err, string(output))
			output, err = exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "rm").CombinedOutput()
			checkError(err, string(output))
			output, err = exec.Command("sudo", "rm", "-r", "/home/"+USER+"/sites/"+chosenSite).CombinedOutput()
			checkError(err, string(output))
		}).Run()

		printInBox("Deleted " + chosenSite)
	}

}

func addSSHKey() {
	var key string
	huh.NewText().
		Title("Enter public key").
		Description("Look in ~/.ssh - file ends in .pub").
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("key cannot be empty")
			}
			return nil
		}).
		Value(&key).
		Run()

	// cmd := exec.Command("echo", key, ">>", "/home/"+USER+"/.ssh/authorized_keys")
	// output, err := cmd.CombinedOutput()
	err := AppendToFile("/home/"+USER+"/.ssh/authorized_keys", key)
	if err != nil {
		log.Fatal(err)
	}

	printInBox("Added SSH key. Have a nice day!")
}

func containerShell(chosenSite string) {
	notice := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(fmt.Sprintf("Connecting shell for %s...", chosenSite))
	fmt.Println(notice)
	// docker exec -it "$sitename" ash
	cmd := exec.Command("docker", "exec", "-it", chosenSite, "ash")
	output, err := cmd.CombinedOutput()

	checkError(err, string(output))
}
