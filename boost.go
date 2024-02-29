package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

const VERSION = "0.0.8"

var USER = os.Getenv("USER")
var chosenOption string
var chosenSite string

type Option struct {
	name       string
	chooseSite bool
	action     func()
}

var options = []Option{
	{"Create Site", false, createSite},
	{"Start Site", true, startSite},
	{"Stop Site", true, stopSite},
	{"Restart Site", true, restartSite},
	{"Delete Site & Files", true, deleteSite},
	{"Change Domain / SSL", true, changeSiteDomain},
	{"Container Shell", true, containerShell},
	{"Fix Permissions", true, fixPermissions},
	{"Migrate Files", true, migrateFiles},
	{"Database Search Replace", true, databaseSearchReplace},
	{"Import WP Database", true, importWPDatabase},
	{"Update WP Database Config", true, changeDatabaseInfo},
	{"Toggle WP Maintenance Mode", true, maintenanceMode},
	{"Add SSH Key", false, addSSHKey},
	{"Generate / View SSH Key", false, generateSshKey},
	{"Prune Docker Images", false, pruneDockerImages},
	{"MariaDB Upgrade", false, mariadbUpgrade},
	{"Fail2ban Status", false, fail2banStatus},
	{"Unban IP", false, unbanIp},
	{"Whitelist IP", false, whitelistIp},
}

func main() {
	CheckForUpdate()
	introForm()
}

func introForm() {
	// reset cursor to beginning of line
	fmt.Print("\033[0G")

	// create list of options to use for intro form
	allOptions := make([]string, 0, len(options))
	for _, option := range options {
		allOptions = append(allOptions, option.name)
	}

	form := huh.NewForm(
		// Ask the user what they want to do.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What do you want to do?").
				Options(
					huh.NewOptions(allOptions...)...,
				).
				Value(&chosenOption),
		),

		// Ask the user for a site.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which site?").
				Options(
					huh.NewOptions(GetDirectoriesInPath("/home/"+USER+"/sites")...)...,
				).
				Value(&chosenSite),
		).WithHideFunc(func() bool {
			for _, option := range options {
				if chosenOption == option.name {
					return !option.chooseSite
				}
			}
			return true
		}),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Run the chosen action
	for _, option := range options {
		if option.name == chosenOption {
			option.action()
			return
		}
	}

	buhBye()
}

func printInBox(content string) {
	fmt.Println(
		lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 3).
			Render(content),
	)

	fmt.Println()
}

// Exit script
func buhBye() {
	printInBox("Buh bye!")
	os.Exit(0)
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

func startSite() {
	var err error
	var output []byte
	//spinner
	spinner.New().Title("Starting " + chosenSite + "...").Action(func() {
		cmd := exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "up", "-d")
		output, err = cmd.CombinedOutput()
	}).Run()
	checkError(err, string(output))
	fmt.Println("Site started. Have a wonderful day!")
}

func stopSite() {
	var err error
	var output []byte

	spinner.New().Title("Stopping " + chosenSite + "...").Action(func() {
		// docker compose -f "/home/$CUR_USER/sites/$sitename/docker-compose.yml" stop
		cmd := exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "stop")
		output, err = cmd.CombinedOutput()
	}).Run()

	checkError(err, string(output))
	fmt.Println("Site stopped. Have a phenomenal day!")
}

func createSite() {
	var sitename string
	var domain string
	php7 := false
	cloudflare := false
	createDb := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter site name").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("site name cannot be empty")
					}
					return nil
				}).
				Value(&sitename),

			huh.NewInput().
				Title("Enter domain(s)").
				Description("Separate domains with a space.").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("domain cannot be empty")
					}
					return nil
				}).
				Value(&domain),

			huh.NewConfirm().
				Title("Requires PHP 7").
				Value(&php7),

			huh.NewConfirm().
				Title("Proxied through Cloudflare").
				Value(&cloudflare),

			huh.NewConfirm().
				Title("Create database").
				Value(&createDb),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	sitename = ReplaceSpacesWithDashes(sitename)

	var db_name string
	var db_user string
	var db_pass string

	const repo_base = "https://raw.githubusercontent.com/BOOST-Creative/docker-server-setup-caddy/main"

	type Download struct {
		source string
		target string
	}

	wordpressCompose := Download{
		source: repo_base + "/wordpress/docker-compose.yml",
		target: "/home/" + USER + "/sites/" + sitename + "/docker-compose.yml",
	}
	htNinja := Download{
		source: repo_base + "/wordpress/.htninja",
		target: "/home/" + USER + "/sites/" + sitename + "/.htninja",
	}
	redisConf := Download{
		source: repo_base + "/wordpress/redis.conf",
		target: "/home/" + USER + "/sites/" + sitename + "/redis.conf",
	}

	getSudo()

	// spinner
	spinner.New().Title("Creating site...").Action(func() {
		// create directory
		err := os.MkdirAll("/home/"+USER+"/sites/"+sitename+"/wordpress", os.ModePerm)
		checkError(err, "Failed to create directory.")

		// download files
		DownloadFile(wordpressCompose.source, wordpressCompose.target)
		DownloadFile(htNinja.source, htNinja.target)
		DownloadFile(redisConf.source, redisConf.target)

		// replace stuff in wordpress docker compose
		if php7 {
			ReplaceTextInFile(wordpressCompose.target, "docker-wordpress-8", "docker-wordpress-7")
		}
		if cloudflare {
			// swap plugins if proxying through cloudflare
			cmd := exec.Command("yq", "-i", ".services.wordpress.environment.ADDITIONAL_PLUGINS = \"w3-total-cache better-wp-security\"", wordpressCompose.target)
			output, err := cmd.CombinedOutput()
			checkError(err, string(output))
			// comment out fail2ban volume mount
			f2bMount := "- /home/CHANGE_TO_USERNAME/server/wp-fail2ban"
			ReplaceTextInFile(wordpressCompose.target, f2bMount, "# "+f2bMount)
		}
		ReplaceTextInFile(wordpressCompose.target, "CHANGE_TO_SITE_NAME", sitename)
		ReplaceTextInFile(wordpressCompose.target, "CHANGE_TO_USERNAME", USER)

		// update domain
		cmd := exec.Command("yq", "-i", fmt.Sprintf(".services.wordpress.labels.caddy = \"%s\"", domain), wordpressCompose.target)
		output, err := cmd.CombinedOutput()
		checkError(err, string(output))

		// create container
		// docker compose -f "/home/$CUR_USER/sites/$sitename/docker-compose.yml" create
		cmd = exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+sitename+"/docker-compose.yml", "create")
		_, err = cmd.CombinedOutput()
		checkError(err, "Failed to create site.")

		// fix permissions
		// sudo chown nobody: "/home/$CUR_USER/sites/$sitename/wordpress"
		cmd = exec.Command("sudo", "chown", "nobody:", "/home/"+USER+"/sites/"+sitename+"/wordpress")
		_, err = cmd.CombinedOutput()
		checkError(err, "Failed to set permissions")

		// create database
		if createDb {
			db_name = ReplaceDashWithUnderscore(sitename)
			db_user = "u_" + ReplaceDashWithUnderscore(sitename)
			db_pass, err = GeneratePassword(14)
			checkError(err, "Failed to generate password.")
			output, err := exec.Command("docker", "exec", "-e", "DB_NAME="+db_name, "mariadb", "bash", "-c", "mysql -uroot -p\"$MYSQL_ROOT_PASSWORD\" -e \"CREATE DATABASE $DB_NAME;\"").CombinedOutput()
			checkError(err, string(output))
			// create user
			output, err = exec.Command("docker", "exec", "-e", "DB_USER="+db_user, "-e", "DB_PASSWORD="+db_pass, "mariadb", "bash", "-c", "mysql -uroot -p\"$MYSQL_ROOT_PASSWORD\" -e \"CREATE USER '$DB_USER'@'%' IDENTIFIED BY '$DB_PASSWORD';\"").CombinedOutput()
			checkError(err, string(output))
			// grant user privileges to database
			output, err = exec.Command("docker", "exec", "-e", "DB_NAME="+db_name, "-e", "DB_USER="+db_user, "mariadb", "bash", "-c", "mysql -uroot -p\"$MYSQL_ROOT_PASSWORD\" -e \"GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'%';\"").CombinedOutput()
			checkError(err, string(output))
		}

	}).Run()

	var sb strings.Builder
	msg := lipgloss.NewStyle().Bold(true).Render("Created " + sitename + "!")

	fmt.Fprint(&sb, msg)

	if createDb {
		keyword := func(s string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(s)
		}
		fmt.Fprintf(&sb,
			"\n\nDatabase: %s\nUsername: %s\nPassword: %s\nServer:   %s",
			keyword(db_name),
			keyword(db_user),
			keyword(db_pass),
			keyword("mariadb"),
		)
		clipboard.WriteAll(fmt.Sprintf("Database: %s\nUsername: %s\nPassword: %s\nServer:   %s", db_name, db_user, db_pass, "mariadb"))
	}

	printInBox(sb.String())
}

func restartSite() {
	var err error
	var output []byte

	spinner.New().Title("Restarting " + chosenSite + "...").Action(func() {
		// docker compose -f "/home/$CUR_USER/sites/$sitename/docker-compose.yml" stop
		cmd := exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "restart")
		output, err = cmd.CombinedOutput()
	}).Run()

	checkError(err, string(output))
	fmt.Println("Site Restarted. Have a superb day!")
}

func fixPermissions() {
	getSudo()
	// spinner
	spinner.New().Title(fmt.Sprintf("Fixing permissions for %s...", chosenSite)).Action(runFixPermissions).Run()
	printInBox("Permissions fixed. Have a fantastic day!")
}

func runFixPermissions() {
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
}

func deleteSite() {
	// TODO: delete database
	confirm := false
	huh.NewConfirm().
		Title(fmt.Sprintf("Are you sure you want to delete %s?", chosenSite)).
		Description("This will COMPLETELY DELETE " + chosenSite + ".").
		Value(&confirm).
		Run()

	if !confirm {
		buhBye()
	}

	getSudo()
	spinner.New().Title("Deleting site...").Action(func() {
		db_name := strings.ReplaceAll(chosenSite, "-", "_")
		// drop database
		output, err := exec.Command("docker", "exec", "-e", "DB_NAME="+db_name, "mariadb", "bash", "-c", "mysql -uroot -p\"$MYSQL_ROOT_PASSWORD\" -e \"DROP DATABASE $DB_NAME;\"").CombinedOutput()
		checkError(err, string(output))
		// stop and remove containers
		output, err = exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "stop").CombinedOutput()
		checkError(err, string(output))
		output, err = exec.Command("docker", "compose", "-f", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml", "rm").CombinedOutput()
		checkError(err, string(output))
		// remove site folder
		output, err = exec.Command("sudo", "rm", "-r", "/home/"+USER+"/sites/"+chosenSite).CombinedOutput()
		checkError(err, string(output))
	}).Run()

	printInBox("Deleted " + chosenSite)

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

	if key == "" {
		buhBye()
	}

	err := AppendToFile("/home/"+USER+"/.ssh/authorized_keys", key)
	if err != nil {
		log.Fatal(err)
	}

	printInBox("Added SSH key. Have a nice day!")
}

func containerShell() {
	notice := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(fmt.Sprintf("Connecting shell for %s...", chosenSite))
	fmt.Println(notice)
	cmd := exec.Command("docker", "exec", "-it", chosenSite, "ash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	checkError(err, "Could not spawn shell")
	printInBox("Have a magnificent day!")
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

	spinner.New().Title("Unbanning IP...").Action(func() {
		// docker exec fail2ban sh -c "fail2ban-client status | grep 'Jail list'" | sed -E 's/^[^:]+:[ \t]+//' | sed 's/,//g'
		jails, err := exec.Command("docker", "exec", "fail2ban", "sh", "-c", "fail2ban-client status | grep 'Jail list' | sed -E 's/^[^:]+:[ \t]+//' | sed 's/,//g'").Output()
		checkError(err, string(jails))
		jailsSlice := strings.Fields(string(jails))

		for _, part := range jailsSlice {
			err = exec.Command("docker", "exec", "fail2ban", "sh", "-c", fmt.Sprintf("fail2ban-client set %s unbanip %s", part, ip)).Run()
			checkError(err, "Error unbanning IP address")
		}
	}).Run()

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
	printInBox(fmt.Sprintf("%s\nPruned docker images. Have a super day!", string(output)))
}

func mariadbUpgrade() {
	// docker exec mariadb sh -c 'mysql_upgrade -uroot -p"$MYSQL_ROOT_PASSWORD"'
	cmd := exec.Command("docker", "exec", "mariadb", "sh", "-c", "mysql_upgrade -uroot -p\"$MYSQL_ROOT_PASSWORD\"")
	output, err := cmd.CombinedOutput()
	checkError(err, string(output))
	printInBox(fmt.Sprintf("%s\nHave a fabulous day!", string(output)))
}

func databaseSearchReplace() {
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

	if search == "" || replace == "" {
		buhBye()
	}

	var output []byte
	var err error
	spinner.New().Title("Searching and replacing...").Action(func() {
		cmd := exec.Command("docker", "exec", chosenSite, "sh", "-c", fmt.Sprintf("cd /usr/src/wordpress && wp search-replace '%s' '%s' --all-tables", search, replace))
		output, err = cmd.CombinedOutput()
	}).Run()

	checkError(err, string(output))
	printInBox(fmt.Sprintf("%s\n\nHave a radical day!", string(output)))
}

func changeSiteDomain() {
	// get current site
	// yq '.services.wordpress.labels.caddy' "/home/$CUR_USER/sites/$sitename/docker-compose.yml"
	output, err := exec.Command("yq", ".services.wordpress.labels.caddy", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml").CombinedOutput()
	checkError(err, string(output))

	var newDomain string
	useSelfSigned := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Change Domain").
				Description("This will change the domain(s) in Caddy\nCurrent domain(s): "+lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(string(output))),

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
				Description("DNS must point to this server to generate.\nIf proxying through Cloudflare, use self-signed and CF setting for \"Full\" SSL.").
				Affirmative("Self-Signed").
				Negative("Generate SSL").
				Value(&useSelfSigned),
		),
	)
	form.Run()

	if newDomain == "" {
		buhBye()
	}

	spinner.New().Title("Changing domain...").Action(func() {
		// update caddy tls option
		if useSelfSigned {
			// yq -i '.services.wordpress.labels."caddy.tls" = "internal"' "/home/$CUR_USER/sites/$sitename/docker-compose.yml"
			cmd := exec.Command("yq", "-i", ".services.wordpress.labels.\"caddy.tls\" = \"internal\"", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml")
			output, err := cmd.CombinedOutput()
			checkError(err, string(output))
		} else {
			// yq -i 'del(.services.wordpress.labels."caddy.tls")' "/home/$CUR_USER/sites/$sitename/docker-compose.yml"
			cmd := exec.Command("yq", "-i", "del(.services.wordpress.labels.\"caddy.tls\")", "/home/"+USER+"/sites/"+chosenSite+"/docker-compose.yml")
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

func generateSshKey() {
	const file = "/root/.ssh/id_ed25519"

	printKey := func() {
		publicKey, _ := exec.Command("sudo", "cat", file+".pub").Output()
		trimmedKey := strings.TrimSpace(string(publicKey))
		msg := fmt.Sprintf("Public key:\n\n%s", trimmedKey)
		err := clipboard.WriteAll(trimmedKey)
		if err == nil {
			msg += "\n\nCopied to clipboard!"
		}
		printInBox(msg)
	}

	// test if file exists
	if err := exec.Command("sudo", "test", "-s", file).Run(); err == nil {
		printKey()
		return
	}

	var passphrase string
	huh.NewInput().
		Title("Enter passphrase").
		Password(true).
		Value(&passphrase).
		Run()

	err := exec.Command("sudo", "ssh-keygen", "-t", "ed25519", "-N", passphrase, "-f", file).Run()
	checkError(err, "Failed to create SSH key.")

	printKey()
}

func migrateFiles() {
	hosts, err := GetHostsFromSSHConfig("/root/.ssh/config", true)
	if err != nil || len(hosts) == 0 {
		printInBox("No hosts found in SSH config.\n\nPlease add to /root/.ssh/config and try again.")
		os.Exit(0)
	}

	var sourceHost string
	var sourcePath string
	var destination = "/home/" + USER + "/sites/" + chosenSite + "/wordpress/"
	form := huh.NewForm(
		huh.NewGroup(
			// select from hosts
			huh.NewSelect[string]().
				Title("Select source host").
				Options(huh.NewOptions(hosts...)...).
				Value(&sourceHost),

			// enter source path
			huh.NewInput().
				Title("Enter source path or file").
				Validate(func(s string) error {
					if s == "" || !strings.HasPrefix(s, "/") {
						return fmt.Errorf("please enter a full path")
					}
					return nil
				}).
				Value(&sourcePath),
		),
	)
	form.Run()

	// make sure there's a trailing slash if the source path is not a file
	if !strings.HasSuffix(sourcePath, "/") && !strings.Contains(sourcePath, ".") {
		sourcePath += "/"
	}

	// confirm options
	var confirm bool
	huh.NewConfirm().
		Title("Everything look good?").
		Description(fmt.Sprintf("Source host: %s\nSource path: %s\nDestination: %s", sourceHost, sourcePath, destination)).
		Value(&confirm).
		Run()

	if !confirm {
		buhBye()
	}

	// rsync
	cmd := exec.Command("sudo", "rsync", "-r", "--progress", sourceHost+":"+sourcePath, destination)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		checkError(err, "Error migrating files:\n\n"+err.Error())
	}

	// fix permissions
	spinner.New().Title(fmt.Sprintf("Fixing permissions for %s...", chosenSite)).Action(runFixPermissions).Run()

	printInBox("Files migrated. Have a splendid day!")
}

// imports last modified sql file in site directory
func importWPDatabase() {
	file, err := FindLastModifiedFile("/home/"+USER+"/sites/"+chosenSite+"/wordpress", ".sql")
	if err != nil {
		checkError(err, err.Error())
	}

	// ask for confirmation
	confirm := true
	huh.NewConfirm().
		Title("Are you sure you want to import this database?\nThis will overwrite the current database.\n").
		Description(fmt.Sprintf("Site: %s\nFile: %s", chosenSite, file.Name())).
		Value(&confirm).
		Run()

	if !confirm {
		buhBye()
	}

	var output []byte
	spinner.New().Title(fmt.Sprintf("Importing database for %s...", chosenSite)).Action(func() {
		cmd := exec.Command("docker", "exec", chosenSite, "sh", "-c", "cd /usr/src/wordpress && wp db import "+file.Name())
		output, err = cmd.CombinedOutput()
	}).Run()
	checkError(err, string(output))
	printInBox(fmt.Sprintf("%s\nHave a grand day!", string(output)))
}

func changeDatabaseInfo() {
	var db_name string
	var db_user string
	var db_pass string
	var db_host = "mariadb"

	notEmpty := func(s string) error {
		if s == "" {
			return fmt.Errorf("cannot be empty")
		}
		return nil
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter database name").
				Validate(notEmpty).
				Value(&db_name),

			huh.NewInput().
				Title("Enter database user").
				Validate(notEmpty).
				Value(&db_user),

			huh.NewInput().
				Title("Enter database password").
				Validate(notEmpty).
				Password(true).
				Value(&db_pass),

			huh.NewInput().
				Title("Enter database host").
				Validate(notEmpty).
				Value(&db_host),
		),
	)

	form.Run()

	if db_name == "" || db_user == "" || db_pass == "" || db_host == "" {
		buhBye()
	}

	filePath := "/home/" + USER + "/sites/" + chosenSite + "/wordpress/wp-config.php"
	updates := map[string]string{
		"DB_NAME":     db_name,
		"DB_USER":     db_user,
		"DB_PASSWORD": db_pass,
		"DB_HOST":     db_host,
	}

	err := UpdateDefineValues(filePath, updates)
	if err != nil {
		checkError(err, err.Error())
	}

	printInBox("Database config updated. Have a marvelous day!")
}

func maintenanceMode() {
	var enable bool
	huh.NewConfirm().
		Title("Maintenance mode for " + chosenSite).
		Value(&enable).
		Affirmative("Enable").
		Negative("Disable").
		Run()

	action := "deactivate"
	if enable {
		action = "activate"
	}

	var output []byte
	var err error
	spinner.New().Title(fmt.Sprintf("Changing maintenance mode for %s...", chosenSite)).Action(func() {
		cmd := exec.Command("docker", "exec", chosenSite, "sh", "-c", "cd /usr/src/wordpress && wp maintenance-mode "+action)
		output, err = cmd.CombinedOutput()
	}).Run()

	checkError(err, string(output))
	printInBox(fmt.Sprintf("%s\nHave a brilliant day!", string(output)))

}
