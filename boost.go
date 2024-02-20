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
	// exec.Command("clear").Run()
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
		fail2banStatus()
	case "Unban IP":
		unbanIp()
	case "Whitelist IP":
		whitelistIp()
	case "Prune Docker Images":
		pruneDockerImages()
	case "MariaDB Upgrade":
		mariadbUpgrade()
	case "Change Site Domain":
		changeSiteDomain(chosenSite)
	case "Database Search Replace":
		databaseSearchReplace(chosenSite)
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

// Grant sudo permissions
func getSudo() {
	err := exec.Command("sudo", "-v").Run()
	checkError(err, "Failed to grant sudo permissions.")
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
	fmt.Println("Site stopped. Have a phenomenal day!")
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
	fmt.Println("Site Restarted. Have a superb day!")
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
	getSudo()

	// spinner
	spinner.New().Title(fmt.Sprintf("Fixing permissions for %s...", chosenSite)).Action(func() {
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
		getSudo()
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

func fail2banStatus() {
	// docker exec fail2ban sh -c "fail2ban-client status | sed -n 's/,//g;s/.*Jail list://p' | xargs -n1 fail2ban-client status"
	cmd := exec.Command("docker", "exec", "fail2ban", "sh", "-c", "fail2ban-client status | sed -n 's/,//g;s/.*Jail list://p' | xargs -n1 fail2ban-client status")
	output, err := cmd.CombinedOutput()
	checkError(err, string(output))
	printInBox(string(output))
}

func unbanIp() {
	var ip string
	huh.NewInput().
		Title("Enter IP to unban").
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("IP address cannot be empty")
			}
			return nil
		}).
		Value(&ip).
		Run()

	script := fmt.Sprintf(`
		JAILS=$(docker exec fail2ban sh -c "fail2ban-client status | grep 'Jail list'" | sed -E 's/^[^:]+:[ \t]+//' | sed 's/,//g');
		for JAIL in $JAILS
		do
			docker exec fail2ban sh -c "fail2ban-client set $JAIL unbanip %s"
		done
	`, ip)
	cmd := exec.Command(script)
	output, err := cmd.CombinedOutput()
	checkError(err, string(output))
	printInBox(fmt.Sprintf("Unbanned %s. Don't forget to whitelist and have a super day!", ip))
}

func whitelistIp() {
	var ip string
	huh.NewInput().
		Title("Enter IP to whitelist").
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("IP address cannot be empty")
			}
			return nil
		}).
		Value(&ip).
		Run()

	getSudo()

	// sudo sed -i "s|ignoreip =.*|& $whitelistip|" ~/server/fail2ban/data/jail.d/jail.local
	cmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("s|ignoreip =.*|& %s|", ip), "/home/"+USER+"/server/fail2ban/data/jail.d/jail.local")
	output, err := cmd.CombinedOutput()
	checkError(err, string(output))
	// docker exec fail2ban sh -c "fail2ban-client reload"
	cmd = exec.Command("docker", "exec", "fail2ban", "sh", "-c", "fail2ban-client reload")
	output, err = cmd.CombinedOutput()
	checkError(err, string(output))
	printInBox(fmt.Sprintf("Whitelisted %s. Have a super day!", ip))
}

func pruneDockerImages() {
	var output []byte
	var err error
	// spinner
	spinner.New().Title("Pruning docker images...").Action(func() {
		// docker image prune -af
		cmd := exec.Command("docker", "image", "prune", "-af")
		output, err = cmd.CombinedOutput()
		checkError(err, string(output))
	}).Run()
	printInBox(fmt.Sprintf("Pruned docker images. Have a super day!\n%s", strings.Split(string(output), "\n")[1]))
}

func mariadbUpgrade() {
	// docker exec mariadb sh -c 'mysql_upgrade -uroot -p"$MYSQL_ROOT_PASSWORD"'
	cmd := exec.Command("docker", "exec", "mariadb", "sh", "-c", `'mysql_upgrade -uroot -p"$MYSQL_ROOT_PASSWORD"'`)
	output, err := cmd.CombinedOutput()
	checkError(err, string(output))
	printInBox(fmt.Sprintf("%s\n\nHave a fabulous day!", string(output)))
}

func databaseSearchReplace(chosenSite string) {
	var search string
	var replace string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter search string").
				Description("This string will be replaced in the database.").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("search string cannot be empty")
					}
					return nil
				}).
				Value(&search),

			huh.NewInput().
				Title("Enter replace string").
				Description("This string will replace the search string.").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("replace string cannot be empty")
					}
					return nil
				}).
				Value(&replace),
		),
	)
	form.Run()

	// docker exec "$sitename" sh -c "cd /usr/src/wordpress && wp search-replace '$searchstring' '$replacestring' --all-tables"
	cmd := exec.Command("docker", "exec", chosenSite, "sh", "-c", fmt.Sprintf("cd /usr/src/wordpress && wp search-replace '%s' '%s' --all-tables", search, replace))
	output, err := cmd.CombinedOutput()
	checkError(err, string(output))
	printInBox(fmt.Sprintf("%s\n\nHave a radical day!", string(output)))
}

func changeSiteDomain(chosenSite string) {
	// get current site
	// yq '.services.wordpress.labels.caddy' "/home/$CUR_USER/sites/$sitename/docker-compose.yml"
	output, err := exec.Command("yq", ".services.wordpress.labels.caddy", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml").CombinedOutput()
	checkError(err, string(output))

	var newDomain string
	var generateSSL bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Change Domain").
				Description(lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(fmt.Sprintf("This will change the domain(s) in Caddy\n\nCurrent domain(s): %s", chosenSite))),

			huh.NewInput().
				Title("Enter new domain").
				Description("Separate domains with a space.").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("domain cannot be empty")
					}
					return nil
				}).
				Value(&newDomain),

			huh.NewConfirm().
				Title("SSL Certificate").
				Affirmative("Let's Encrypt / ZeroSSL").
				Negative("Self-Signed").
				Value(&generateSSL),
		),
	)
	form.Run()

	spinner.New().Title("Changing domain...").Action(func() {
		// update caddy tls option
		if generateSSL {
			// yq -i 'del(.services.wordpress.labels."caddy.tls")' "/home/$CUR_USER/sites/$sitename/docker-compose.yml"
			cmd := exec.Command("yq", "-i", "del(.services.wordpress.labels.\"caddy.tls\")", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml")
			output, err := cmd.CombinedOutput()
			checkError(err, string(output))
		} else {
			// yq -i '.services.wordpress.labels."caddy.tls" = "internal"' "/home/$CUR_USER/sites/$sitename/docker-compose.yml"
			cmd := exec.Command("yq", "-i", ".services.wordpress.labels.\"caddy.tls\" = \"internal\"", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml")
			output, err := cmd.CombinedOutput()
			checkError(err, string(output))
		}

		// update caddy domain
		// yq -i ".services.wordpress.labels.caddy = \"$newdomain\"" "/home/$CUR_USER/sites/$sitename/docker-compose.yml"
		cmd := exec.Command("yq", "-i", fmt.Sprintf(".services.wordpress.labels.caddy = \"%s\"", newDomain), "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml")
		output, err = cmd.CombinedOutput()
		checkError(err, string(output))

		// reload site
		cmd = exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "up", "-d")
		output, err = cmd.CombinedOutput()
		checkError(err, string(output))

	}).Run()
	checkError(err, string(output))

	printInBox("Domain updated. Have a tubular day!")

}
