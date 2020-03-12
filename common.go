package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func CheckoutOnBranch(repository *git.Repository, issue *gitlab.Issue) {
	log.Println("Checking out")
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
