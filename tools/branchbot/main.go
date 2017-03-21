package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	flVerbose        bool
	syncWait         = time.Second * 5
	labelPrefix      = "process/cherrypick-"
	cherrypickLabels = []string{
		labelPrefix + "17.03",
		labelPrefix + "17.04",
	}
)

const (
	githubUser      = "docker"
	githubRepo      = "editions"
	editionsRepoDir = "editions"
	originRemote    = "git@github.com:docker/editions"
	botRemote       = "git@github.com:nathanleclaire/editions"
	masterBranch    = "master"
	edgeBranch      = "edge"
	stableBranch    = "stable"

	// TODO(nathanleclaire): Change to editionsbot.
	botUser = "nathanleclaire"

	// TODO(nathanleclaire): Should include all editions maintainers, just didn't want to
	// spam before the bot was fully ready.
	maintainers = "@nathanleclaire "
)

func git(args ...string) (string, error) {
	// Run 'git' command within a specific work context.
	cmd := exec.Command(
		"git",
		append([]string{
			"--git-dir",
			editionsRepoDir,
			"--work-tree",
			editionsRepoDir,
		}, args...)...)
	verbose(cmd.Args)
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	verbose(outStr)
	return outStr, err
}

func verbose(out ...interface{}) {
	if flVerbose {
		log.Print(out)
	}
}

func fetch() error {
	if _, err := git("fetch", "origin"); err != nil {
		return err
	}
	return nil
}

func mkRepo() error {
	if _, err := os.Stat(editionsRepoDir); os.IsNotExist(err) {
		log.Print("No repo found, cloning...")
		if _, err := git("clone", originRemote); err != nil {
			return err
		}
	}
	return nil
}

func rebase(upstream, branch string) error {
	if _, err := git("rebase", upstream, branch); err != nil {
		return err
	}
	return nil
}

func push(branch string) error {
	if _, err := git("push", botRemote, branch); err != nil {
		return err
	}
	return nil
}

func checkout(branch string, newBranch bool) error {
	var args []string
	if newBranch {
		args = []string{"checkout", "-b", branch}
	} else {
		args = []string{"checkout", branch}
	}
	if _, err := git(args...); err != nil {
		return err
	}
	return nil
}

func cherrypick(sha string) error {
	// -x:
	//   Add "Cherry-picked from <SHA>" to commit message
	// --stategy=recursive -Xtheirs:
	//   favor the cherry-picked commits over existing code in merge
	// -Xpatience:
	//   spend a little bit more time trying to find the best diff
	if _, err := git("cherry-pick", "-x", "--strategy=recursive", "-Xtheirs", "-Xpatience", sha); err != nil {
		return err
	}
	return nil
}

func main() {
	var flAccessToken string

	flag.BoolVar(&flVerbose, "verbose", false, "print verbose output")
	flag.StringVar(&flAccessToken, "token", "", "github API token")
	flag.Parse()

	if err := mkRepo(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	if err := fetch(); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: flAccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// GH label filter seems to be an 'and' filter, so run the query once
	// per cherrypick label.
	for _, cherrypickLabel := range cherrypickLabels {
		issueListByRepoOpts := &github.IssueListByRepoOptions{
			State:  "all",
			Labels: []string{cherrypickLabel},
		}

		// PRs are issues. So we get a list of issues by label.
		issues, _, err := client.Issues.ListByRepo(ctx, githubUser, githubRepo, issueListByRepoOpts)
		if err != nil {
			log.Print(err)
			continue
		}

		targetBranch := strings.Replace(cherrypickLabel, labelPrefix, "", -1)

	ProcessingPRs:
		// TODO: Possibly validate that this is a PR, not a normal
		// issue (unlikely to be accidentally tagged, but possible)
		for _, issue := range issues {
			cherrypickBranch := fmt.Sprintf("cherrypick-%d-%s", issue.GetNumber(), targetBranch)

			if err := checkout(targetBranch, false); err != nil {
				log.Print("Error checking out target branch: ", err)
				os.Exit(1)
			}
			if err := checkout(cherrypickBranch, true); err != nil {
				log.Print("Error creating cherry pick branch: ", err)
				os.Exit(1)
			}

			commits, _, err := client.PullRequests.ListCommits(ctx, githubUser, githubRepo, issue.GetNumber(), &github.ListOptions{})
			if err != nil {
				log.Print(err)
				continue
			}

			for _, commit := range commits {
				// cherrypick(commit)
				log.Print("Cherry picking ", *commit.SHA)
				if err := cherrypick(*commit.SHA); err != nil {
					// If the cherry pick bombs, just go to
					// the next PR to process. We won't try
					// to salvage it, for now at least.
					log.Print("Error with git cherry-pick: ", err)
					continue ProcessingPRs
				}
			}

			if err := push(cherrypickBranch); err != nil {
				log.Print("Error pushing cherry pick branch to git remote: ", err)
				continue ProcessingPRs
			}

			prHead := botUser + ":" + cherrypickBranch
			prBase := targetBranch
			prTitle := fmt.Sprintf("[automated] Cherry-pick pull request #%d to %s", issue.GetNumber(), targetBranch)
			prBody := fmt.Sprintf(`
_This is a pull request from a robot._

This PR cherry-picks #%d into the `+"`%s`"+` branch. :cherries:

ping %s
`, issue.GetNumber(), targetBranch, maintainers)

			newPR := &github.NewPullRequest{
				Head:  &prHead,
				Base:  &prBase,
				Title: &prTitle,
				Body:  &prBody,
			}

			pr, _, err := client.PullRequests.Create(ctx, githubUser, githubRepo, newPR)
			if err != nil {
				log.Print("Error creating PR: ", err)
				continue ProcessingPRs
			}

			log.Print("Created PR #", pr.GetNumber())

			if err := checkout("master", false); err != nil {
				log.Print("Error checking out master: ", err)
				continue ProcessingPRs
			}

			_, err = client.Issues.RemoveLabelForIssue(ctx, githubUser, githubRepo, issue.GetNumber(), cherrypickLabel)
			if err != nil {
				log.Print("Error deleting label: ", err)
				continue
			}
		}
	}
}
