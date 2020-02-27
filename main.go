package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type CreateArguments struct {
	Title       string
	Description string
	BranchOut   bool
}

func parseCreateArguments(args []string) CreateArguments {
	createCommand := flag.NewFlagSet("create", flag.ExitOnError)
	title := createCommand.String("t", "", "Title of the issue")
	description := createCommand.String("d", "", "Description of the issue")
	branchOut := createCommand.Bool("b", false, "Should a branch be created for this issue")
	createCommand.Parse(args)
	if *title == "" {
		createCommand.Usage()
		os.Exit(1)
	}
	if *description == "" {
		log.Println("Assuming description is equal to the title ('-d')")
		description = title
	}
	return CreateArguments{
		Title:       *title,
		Description: *description,
		BranchOut:   *branchOut,
	}
}

func prettyPrint(issue *gitlab.Issue) {
	fmt.Println(fmt.Sprintf(`Issue :
	ID : %d
	Title : %s
	Url : %s
	`, issue.IID, issue.Title, issue.WebURL))
}

func contains(arr []string, s string) bool {
	for _, n := range arr {
		if s == n {
			return true
		}
	}
	return false
}

func main() {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	if gitlabToken == "" {
		log.Fatalln("Set the GITLAB_TOKEN environment variable for this utility to work")
	}
	gitlabClient := gitlab.NewClient(nil, gitlabToken)

	permittedOperations := []string{"create", "info"}

	if len(os.Args) < 2 || !contains(permittedOperations, os.Args[1]) {
		log.Fatalf("Provide an operation from [%s]", strings.Join(permittedOperations, ","))
	}
	operation := os.Args[1]

	//Get dir where it's launched from
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	//Load repo
	repo, err := git.PlainOpen(rootDir)
	if err != nil {
		log.Fatalln("Error loading repo", err)
	}

	repoConfig, err := repo.Config()
	if err != nil {
		log.Fatalln("Error loading repo config", err)
	}

	//Select remote url
	fmt.Println("Assuming 'origin' remote -- nothing else is supported now")
	remoteURL := repoConfig.Remotes["origin"].URLs[0]

	//Extract 'namespace/project' from URL
	r, err := regexp.Compile(`^git@gitlab.com:(.*)\.git$`)
	if err != nil {
		log.Fatalln("Could not compile regex", err)
	}

	urlMatches := r.FindStringSubmatch(remoteURL)
	fullProjectName := urlMatches[1]
	fmt.Println(fmt.Sprintf("Found project : %s", fullProjectName))

	//Get the project
	project, _, err := gitlabClient.Projects.GetProject(fullProjectName, nil)
	if err != nil {
		log.Fatalln("Error getting project", err)
	}

	if operation == "create" {
		config := parseCreateArguments(os.Args[2:])

		createIssueOptions := &gitlab.CreateIssueOptions{
			Title:       gitlab.String(config.Title),
			Description: gitlab.String(config.Description),
		}

		issue, _, err := gitlabClient.Issues.CreateIssue(project.ID, createIssueOptions)
		if err != nil {
			log.Fatalln("Error creating issue", err)
		}
		prettyPrint(issue)

		if config.BranchOut {
			fmt.Println("Checking out")
			w, err := repo.Worktree()
			if err != nil {
				log.Fatalln("Error loading repo worktree", err)
			}
			err = w.Checkout(
				&git.CheckoutOptions{
					Create: true,
					Keep:   true,
					Branch: plumbing.NewBranchReferenceName(fmt.Sprintf("%d-%s", issue.IID, strings.ReplaceAll(strings.ToLower(issue.Title), " ", "-"))),
				},
			)
			if err != nil {
				log.Fatalln("Error while running 'git checkout'", err)
			}
		}
	}

	if operation == "info" {
		head, err := repo.Head()
		if err != nil {
			log.Fatalln("Error loading repo head", err)
		}
		//Extract 'issue number' from name
		r, err := regexp.Compile(`^(\d+)-.*$`)
		if err != nil {
			log.Fatalln("Could not compile regex", err)
		}
		branchName := head.Name().Short()
		matches := r.FindStringSubmatch(branchName)
		if len(matches) < 2 {
			log.Fatalln("Could not find an issue number")
		}
		issueID, _ := strconv.ParseInt(matches[1], 10, 32)

		issue, _, err := gitlabClient.Issues.GetIssue(project.ID, int(issueID))
		if err != nil {
			log.Fatalln("Error getting issue", err)
		}
		prettyPrint(issue)
	}
}
