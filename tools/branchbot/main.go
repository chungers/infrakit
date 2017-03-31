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

const (
	githubUser       = "docker"
	githubRepo       = "editions"
	editionsRepoDir  = "editions"
	originRemote     = "git@github.com:docker/editions"
	botUser          = "editionsbot"
	badObjectMessage = "bad object"

	// TODO(nathanleclaire): Should include all editions maintainers, just didn't want to
	// spam before the bot was fully ready.
	maintainers = "@nathanleclaire @FrenchBen @ddebroy @kencochrane "
)

var (
	flVerbose        bool
	syncWait         = time.Second * 30
	labelPrefix      = "process/cherrypick-"
	cherrypickLabels = []string{
		labelPrefix + "17.03",
		labelPrefix + "17.04",
	}
	cherrypickCompleteLabel   = labelPrefix + "complete"
	cherrypickImpossibleLabel = labelPrefix + "impossible"
	botRemote                 = "git@github.com:" + botUser + "/editions"
)

func git(args ...string) (string, error) {
	// Run 'git' command within a specific work context.
	cmd := exec.Command(
		"git",
		append([]string{
			"--no-pager",
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
	if _, err := git("fetch", "--all"); err != nil {
		return err
	}
	return nil
}

func createFreshRepo() error {
	// Creates a "fresh" copy of the editions repo locally for PR
	// processing.
	if _, err := os.Stat(editionsRepoDir); err == nil {
		if err := os.RemoveAll(editionsRepoDir); err != nil {
			return err
		}
	}
	log.Print("No repo found, cloning...")
	if _, err := git("clone", originRemote); err != nil {
		return err
	}
	return nil
}

func rebase(upstream, branch string) error {
	// When doing a rebase we will sometimes be rebasing an upstream onto a
	// cherrypick branch.
	//
	// We want to use 'ours' strategy because we want to favor the commits
	// of the upstream branch.
	if _, err := git("rebase", "--strategy=recursive", "-X", "ours", upstream, branch); err != nil {
		return err
	}
	return nil
}

func push(branch string) error {
	// Defaults to force push. This allows us to rebase the local branches
	// however we like.
	if _, err := git("push", "-f", botRemote, branch); err != nil {
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

func cherrypick(sha string) (string, error) {
	// -x:
	//   Add "Cherry-picked from <SHA>" to commit message
	// --stategy=recursive -Xtheirs:
	//   favor the cherry-picked commits over existing code in merge
	// -Xpatience:
	//   spend a little bit more time trying to find the best diff
	return git("cherry-pick", "-x", "--strategy=recursive", "-Xtheirs", "-Xpatience", sha)
}

func markImpossibleToCherrypick(ctx context.Context, client *github.Client, originalLabel string, number int) error {
	// This PR is impossible to cherry-pick automatically, so add the
	// "process/cherrypick-impossible" label and remove the
	// "process/cherrypick-version" label
	if _, _, err := client.Issues.AddLabelsToIssue(ctx, githubUser, githubRepo, number, []string{cherrypickImpossibleLabel}); err != nil {
		return err
	}

	if _, err := client.Issues.RemoveLabelForIssue(ctx, githubUser, githubRepo, number, originalLabel); err != nil {
		return err
	}

	return nil
}

func checkCherrypickLabels(client *github.Client) error {
	ctx := context.Background()

	// First step of cherrypick bot: If there are open cherrypick bot PRs,
	// attempt to rebase them to the latest remote of the target branch.
	openBotPRs, _, err := client.Issues.ListByRepo(ctx, githubUser, githubRepo, &github.IssueListByRepoOptions{
		State:   "open",
		Creator: botUser,
	})
	if err != nil {
		log.Print("Error trying to get open bot PRs: ", err)
		return err
	}

	for _, issue := range openBotPRs {
		pr, _, err := client.PullRequests.Get(ctx, githubUser, githubRepo, issue.GetNumber())
		if err != nil {
			log.Print("Error getting pull request: ", err)
			return err
		}

		if err := checkout(*pr.Head.Ref, false); err != nil {
			log.Print("Error checking out supposedly existing PR branch: ", err)
			return err
		}

		// TODO(nathanleclaire): Any better way to calculate the
		// upstream for rebase?
		if err := rebase("origin/"+*pr.Base.Ref, *pr.Head.Ref); err != nil {
			log.Print("Rebase problem: ", err)
			return err
		}

		if err := push(*pr.Head.Ref); err != nil {
			log.Print("Rebase problem: ", err)
			return err
		}
	}

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

		// Second step of cherry pick bot: Submit new PRs where
		// requested.
	ProcessingPRs:
		// TODO: Possibly validate that this is a PR, not a normal
		// issue (unlikely to be accidentally tagged, but possible)
		for _, issue := range issues {
			if _, err := git("reset", "--hard", "origin/"+targetBranch); err != nil {
				log.Print("Error resetting to origin/"+targetBranch+": ", err)

				// Return on error since this reset MUST succeed
				return err
			}

			cherrypickBranch := fmt.Sprintf("cherrypick-%d-%s", issue.GetNumber(), targetBranch)

			if err := checkout("origin/"+targetBranch, false); err != nil {
				log.Print("Error checking out target branch: ", err)

				// Return on error since this checkout MUST succeed
				return err
			}

			// -D: force delete
			// -r: delete ref (otherwise would need to 'checkout' first)
			if _, err := git("branch", "-D", "-r", botUser+"/"+cherrypickBranch); err != nil {
				log.Print("Couldn't delete local cherrypick branch ref: ", err)
				// OK to keep moving, since sometimes we won't
				// have an existing local ref, which is fine
			}

			if err := checkout(cherrypickBranch, true); err != nil {
				log.Print("Error creating cherry pick branch: ", err)

				// Return on error since this checkout MUST succeed
				return err
			}

			commits, _, err := client.PullRequests.ListCommits(ctx, githubUser, githubRepo, issue.GetNumber(), &github.ListOptions{})
			if err != nil {
				log.Print(err)
				continue ProcessingPRs
			}

			// TODO(nathanleclaire): Break this block into smaller routines
			for _, commit := range commits {
				log.Print("Cherry picking ", *commit.SHA)
				if cherrypickOut, err := cherrypick(*commit.SHA); err != nil {
					log.Print("Error with git cherry-pick: ", err)

					// In most cases if the cherry-pick
					// fails we will just skip to
					// processing the next PR (it will need
					// manual intervention and/or updates
					// to the bot).
					//
					// However in some cases the commits
					// won't be present in the repo because
					// it might have been brought in due to
					// a 'squash merge' (GitHub-specific
					// feature).
					//
					// In those cases we'll check for the
					// familiar error message of this type
					// of situation and grep the repo for
					// the created squash commit to see if
					// we can simply cherry-pick that
					// instead.
					if strings.Contains(cherrypickOut, badObjectMessage) {
						// It might have been a
						// squash/rebase merge, so
						// let's try to cherry-pick
						// that if that's so.
						shas, err := git("log", "--all", fmt.Sprintf("--grep=#%d", issue.GetNumber()), "--oneline", "--pretty=format:%H")
						if err != nil {
							// Just skip if there
							// was error here.
							log.Print("Error attempting to look up squash merge commit: ", err)
							continue ProcessingPRs
						}
						// Might return multiple
						// commits due to multiple
						// cherry-picks, so get the
						// oldest one.
						allSHAs := strings.Split(shas, "\n")
						sha := allSHAs[len(allSHAs)-1]
						if _, err := cherrypick(sha); err != nil {
							log.Print("Error attempting to cherry-pick rebase/squash merge: ", err)
							if err := markImpossibleToCherrypick(ctx, client, cherrypickLabel, issue.GetNumber()); err != nil {
								log.Print("Error attempting to label as non-cherrypickable: ", err)
							}

							continue ProcessingPRs
						}

						// Cool, we cherrypicked the
						// squash commit as expected,
						// so carry on with business as
						// usual.
						break
					}

					// We got an error but it wasn't any of
					// the expected ones. Mark this PR as
					// problematic.
					if err := markImpossibleToCherrypick(ctx, client, cherrypickLabel, issue.GetNumber()); err != nil {
						log.Print("Error attempting to label as non-cherrypickable: ", err)
					}

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
				continue
			}

			log.Print("Created PR #", pr.GetNumber())

			if err := checkout("master", false); err != nil {
				log.Print("Error checking out master: ", err)
				continue ProcessingPRs
			}

			_, err = client.Issues.RemoveLabelForIssue(ctx, githubUser, githubRepo, issue.GetNumber(), cherrypickLabel)
			if err != nil {
				log.Print("Error deleting label: ", err)
				continue ProcessingPRs
			}

			_, _, err = client.Issues.AddLabelsToIssue(ctx, githubUser, githubRepo, issue.GetNumber(), []string{cherrypickCompleteLabel})
			if err != nil {
				log.Print("Error adding cherrypick-complete label: ", err)
				continue ProcessingPRs
			}
		}
	}

	return nil
}

func main() {
	var flAccessToken string

	flag.BoolVar(&flVerbose, "verbose", false, "print verbose output")
	flag.StringVar(&flAccessToken, "token", "", "github API token")
	flag.Parse()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: flAccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Loop until canceled by user.
	//
	// TODO(nathanleclaire): Gracefully exit (clean up / finish tag
	// modifications) when receiving SIGTERM.
	for {
		if err := createFreshRepo(); err != nil {
			log.Fatal("Error cloning repo: ", err)
		}

		// TODO(nathanleclaire): Originally I used --work-tree and
		// related flags but it seems to mess with the usual git
		// command flow, esp. regarding rebase
		if err := os.Chdir(editionsRepoDir); err != nil {
			log.Fatal("Couldn't chdir to source code repo: ", err)
		}

		if _, err := git("remote", "add", botUser, botRemote); err != nil {
			log.Fatal("Error creating remote: ", err)
		}

		if err := fetch(); err != nil {
			log.Fatal("Error fetching latest changes: ", err)
		}

		// TODO(nathanleclaire): This doesn't quite work as intended
		// yet, but the basic idea (to prune merged remote branches)
		// seems sound.
		if _, err := git("remote", "prune", botRemote); err != nil {
			log.Fatal("Error pruning remote: ", err)
		}

		if err := checkCherrypickLabels(client); err != nil {
			log.Print("Error running cherrypick check: ", err)
		}

		// TODO(nathanleclaire): This is kind of a weird UNIX-ism,
		// although Windows FS might work the same way. Investigate
		// whether this is truly the best way to handle this mechanic
		if err := os.Chdir(".."); err != nil {
			log.Fatal("Couldn't switch back to parent directory")
		}

		time.Sleep(syncWait)
	}

}
