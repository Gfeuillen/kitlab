package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

//CreateArgs :
type CreateArgs struct {
	Title       string
	Description string
	BranchOut   bool
}

//ParseCreateArgs :
func ParseCreateArgs(args []string) CreateArgs {
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
	return CreateArgs{
		Title:       *title,
		Description: *description,
		BranchOut:   *branchOut,
	}
}

//CreateGitlabIssue :
func CreateGitlabIssue(config CreateArgs, gitlabClient *gitlab.Client, gitlabProject *gitlab.Project, repository *git.Repository) {
	createIssueOptions := &gitlab.CreateIssueOptions{
		Title:       gitlab.String(config.Title),
		Description: gitlab.String(config.Description),
	}
	issue, _, err := gitlabClient.Issues.CreateIssue(gitlabProject.ID, createIssueOptions)
	if err != nil {
		log.Fatalln("Error creating issue", err)
	}
	prettyPrint(issue)
	if config.BranchOut {
		fmt.Println("Checking out")
		w, err := repository.Worktree()
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
