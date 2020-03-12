package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/src-d/go-git.v4"
)

func check(e error) {
	if e != nil {
		panic(e)
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

func findGitRootDir(absolutePath string) (*git.Repository, error) {
	//Load repo
	repo, err := git.PlainOpen(absolutePath)
	if err != nil {
		parent, _ := path.Split(absolutePath)
		if parent != "" {
			return findGitRootDir(parent)
		}
	}
	return repo, err
}

func main() {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	if gitlabToken == "" {
		log.Fatalln("Set the GITLAB_TOKEN environment variable for this utility to work")
	}
	gitlabClient := gitlab.NewClient(nil, gitlabToken)

	permittedOperations := []string{"create", "info"}

	if len(os.Args) < 2 || !contains(permittedOperations, os.Args[1]) {
		log.Fatalf("Provide an operation from [ %s ]", strings.Join(permittedOperations, " | "))
	}
	operation := os.Args[1]

	//Get dir where it's launched from
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	//Load repo
	repo, err := findGitRootDir(rootDir)
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
	if len(urlMatches) < 2 {
		log.Fatalln("Did not recognize gitlab repository")
	}
	fullProjectName := urlMatches[1]
	fmt.Println(fmt.Sprintf("Found project : %s", fullProjectName))

	//Get the project
	project, _, err := gitlabClient.Projects.GetProject(fullProjectName, nil)
	if err != nil {
		log.Fatalln("Error getting project", err)
	}

	switch operation {
	case "create":
		config := ParseCreateArgs(os.Args[2:])
		CreateGitlabIssue(config, gitlabClient, project, repo)
	case "info":
		InfoIssue(gitlabClient, project, repo)
	default:
		log.Fatalf("Provide an operation from [%s]", strings.Join(permittedOperations, ","))
	}
}
