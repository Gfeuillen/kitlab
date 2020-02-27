package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

func createIssue(gitlabClient *gitlab.Client, fullProjectName string, title string, description string) *gitlab.Issue {
	//Get the project
	project, _, err := gitlabClient.Projects.GetProject(fullProjectName, nil)
	if err != nil {
		log.Fatalln("Error getting project", err)
	}

	createIssueOptions := &gitlab.CreateIssueOptions{
		Title:       gitlab.String(title),
		Description: gitlab.String(description),
	}

	issue, _, err := gitlabClient.Issues.CreateIssue(project.ID, createIssueOptions)
	if err != nil {
		log.Fatalln("Error creating issue", err)
	}
	return issue
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

	permittedOperations := []string{"create"}

	if len(os.Args) < 2 || !contains(permittedOperations, os.Args[1]) {
		log.Fatalf("Provide an operation from [%s]", strings.Join(permittedOperations, ","))
	}
	operation := os.Args[1]

	flag.NewFlagSet("count", flag.ExitOnError)

	if operation == "create" {
		config := parseCreateArguments(os.Args[2:])

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

		fmt.Println(repoConfig.Remotes["origin"].URLs[0])

		//Select remote url
		fmt.Println("Assuming 'origin' remote -- nothing else is supported now")
		remoteURL := repoConfig.Remotes["origin"].URLs[0]

		//Extract 'namespace/project' from URL
		r, err := regexp.Compile(`^git@gitlab.com:(.*)\.git$`)
		if err != nil {
			log.Fatalln("Could not compile regex", err)
		}

		urlMatches := r.FindStringSubmatch(remoteURL)
		fmt.Println(fmt.Sprintf("Found project : %s", urlMatches[1]))

		gitlabClient := gitlab.NewClient(nil, gitlabToken)

		issue := createIssue(gitlabClient, urlMatches[1], config.Title, config.Description)
		fmt.Println(fmt.Sprintf(`Issue created :
        ID : %d
        Title : %s
        Url : %s
        `, issue.IID, issue.Title, issue.WebURL))

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
}
