package main

import (
	"log"
	"regexp"
	"strconv"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/src-d/go-git.v4"
)

//InfoIssue :
func InfoIssue(gitlabClient *gitlab.Client, gitlabProject *gitlab.Project, repository *git.Repository) {
	head, err := repository.Head()
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

	issue, _, err := gitlabClient.Issues.GetIssue(gitlabProject.ID, int(issueID))
	if err != nil {
		log.Fatalln("Error getting issue", err)
	}
	prettyPrint(issue)
}
