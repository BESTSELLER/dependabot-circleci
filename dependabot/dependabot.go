package dependabot

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/BESTSELLER/dependabot-circleci/circleci"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/google/go-github/v60/github"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Start will run through all repos it has access to and check for updates and make pull requests if needed.
func Start(ctx context.Context, client *github.Client, org string, repositories []string) {
	// If we are running in Bestseller specific mode, we need to set the running variable to true
	// To be able to query private orbs and docker images
	config.AppConfig.BestsellerSpecific.Running = org == "BESTSELLER"

	// get repos
	// TODO: Get only repos in the list, but in a single API Call
	repos, err := gh.GetRepos(ctx, client, 1)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to repos")
	}

	// Loop through all repos
	for _, repoName := range repositories {
		for _, repository := range repos {
			// only check repo if repo exists in our db and is scheduled to trigger today
			if repository.GetName() == repoName {
				checkRepo(ctx, client, repository)
				break
			}
		}
	}
}

func checkRepo(ctx context.Context, client *github.Client, repo *github.Repository) {
	// defer wg.Done()

	repoName := repo.GetName()

	log.Debug().Msg(fmt.Sprintf("Checking repo: %s", repoName))

	// should we then remove the repo from our db ?
	if repo.GetArchived() {
		log.Debug().Msg(fmt.Sprintf("Repo '%s' is archived", repoName))
		return
	}

	// Use this to test the application against a single repo
	// if repoName != "bestone-bi4-sales-core-salesorderservice" {
	// 	return
	// }

	repoConfig := getRepoConfig(ctx, client, repo)
	if repoConfig == nil {
		return
	}

	// determine repo details
	repoOwner := repo.GetOwner().GetLogin()
	repoDefaultBranch := repo.GetDefaultBranch()

	targetBranch := getTargetBranch(ctx, client, repoOwner, repoName, repoDefaultBranch, repoConfig)
	if targetBranch == "" {
		return
	}

	go datadog.IncrementCount("analysed_repos", 1, []string{fmt.Sprintf("organization:%s", repoOwner)})
	parseRepoContent(ctx, client, repoConfig, repoOwner, repoName, targetBranch, repoConfig.ConfigPath)
}

func parseRepoContent(ctx context.Context, client *github.Client, repoConfig *config.RepoConfig, owner, repo, branch, pathInRepo string) {
	log.Info().Msgf("Processing: %s", pathInRepo)
	// 1. Get directory contents
	options := &github.RepositoryContentGetOptions{Ref: branch}
	fileContent, directoryContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, pathInRepo, options)
	if err != nil {
		log.Error().Err(err).Msgf("could not parseRepoContent %s", repo)
		return
	}
	if fileContent == nil {
		for _, dir := range directoryContent {
			if getRelativeDirDepth(repoConfig.ConfigPath, dir.GetPath()) > repoConfig.ScanDepth {
				return
			}
			go parseRepoContent(ctx, client, repoConfig, owner, repo, branch, dir.GetPath())
		}
		return
	}
	// 3. Check if file is .yml/.yaml - if not, skip - if yes, process
	if filename := fileContent.GetName(); !strings.HasSuffix(filename, ".yml") && !strings.HasSuffix(filename, ".yaml") {
		log.Info().Msgf("Skipping %s, not yml/yaml", filename)
		return
	}
	content, err := fileContent.GetContent()
	if err != nil {
		log.Error().Err(err).Msgf("could not fileContent.GetContent() %s", repo)
	}
	// unmarshal
	var cciconfig yaml.Node
	err = yaml.Unmarshal([]byte(content), &cciconfig)
	if err != nil {
		log.Error().Err(err).Msgf("could not unmarshal yaml in %s", repo)
		return
	}

	// check for updates
	orbUpdates, dockerUpdates := circleci.GetUpdates(&cciconfig)
	for old, update := range orbUpdates {
		// wg.Add(1)
		err = handleUpdate(ctx, client, update, "orb", old, content, owner, repo, branch, pathInRepo, fileContent.SHA, repoConfig)
		if err != nil {
			go datadog.IncrementCount("failed_repos", 1, []string{fmt.Sprintf("organization:%s", owner)})
			return
		}
	}
	for old, update := range dockerUpdates {
		// wg.Add(1)
		err = handleUpdate(ctx, client, update, "docker", old, content, owner, repo, branch, pathInRepo, fileContent.SHA, repoConfig)
		if err != nil {
			go datadog.IncrementCount("failed_repos", 1, []string{fmt.Sprintf("organization:%s", owner)})
			return
		}
	}
}

func getRepoConfig(ctx context.Context, client *github.Client, repo *github.Repository) *config.RepoConfig {
	// check if a bot config exists
	repoConfigContent, _, err := gh.GetRepoFileBytes(ctx, client, repo.GetOwner().GetLogin(), repo.GetName(), ".github/dependabot-circleci.yml", "")
	if err != nil {
		log.Debug().Err(err).Msgf("could not load dependabot-circleci.yml in repo: %s", repo.GetName())
		return nil
	}

	repoConfig, err := config.ReadRepoConfig(repoConfigContent)
	if err != nil {
		log.Error().Err(err).Msgf("could not read repo config in %s", repo.GetName())
		return nil
	}

	return repoConfig
}

func getTargetBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, defaultBranch string, repoConfig *config.RepoConfig) string {
	targetBranch := defaultBranch
	if repoConfig.TargetBranch != "" {
		_, _, err := client.Repositories.GetBranch(ctx, repoOwner, repoName, repoConfig.TargetBranch, 3)
		if err != nil {
			return ""
		}
		targetBranch = repoConfig.TargetBranch
	}
	return targetBranch
}

func handleUpdate(ctx context.Context, client *github.Client, update *yaml.Node, updateType, old, content, repoOwner, repoName, targetBranch, pathInRepo string, SHA *string, repoConfig *config.RepoConfig) error {
	// defer wg.Done()

	log.Debug().Msgf("repo: %s, old: %s, update: %s", repoName, old, update.Value)
	newYaml := circleci.ReplaceVersion(update, old, content)

	// commit vars
	var oldVersion, newVersion []string
	if updateType == "orb" {
		oldVersion = strings.Split(old, "@")
		newVersion = strings.Split(update.Value, "@")
	} else {
		oldVersion = strings.Split(old, ":")
		newVersion = strings.Split(update.Value, ":")
	}

	if updateType == "orb" && len(newVersion) == 1 {
		return fmt.Errorf("could not find orb version for %s in %s", update.Value, repoName)
	}

	commitMessage := fmt.Sprintf("Bump @%s from %s to %s", oldVersion[0], oldVersion[1], newVersion[1])
	commitBranch := fmt.Sprintf("dependabot-circleci/%s/%s", updateType, strings.ReplaceAll(update.Value, ":", "@"))

	// err := check and create branch
	exists, oldPRs, err := gh.CheckPR(ctx, client, repoOwner, repoName, targetBranch, commitBranch, commitMessage, oldVersion[0])
	if err != nil {
		log.Error().Err(err).Msgf("could not get old branch in %s", repoName)
		return err
	}
	if exists {
		return err
	}

	notExists := gh.CheckBranch(ctx, client, repoOwner, repoName, github.String(commitBranch))
	if notExists {
		err = gh.CreateBranch(ctx, client, repoOwner, repoName, targetBranch, github.String(commitBranch))
		if err != nil {
			log.Error().Err(err).Msgf("could not create branch %s in %s", commitBranch, repoName)
			return err
		}
	} else {
		log.Debug().Msgf("branch %s already exists, skipping creation of branch", commitBranch)
	}

	// commit file
	err = gh.UpdateFile(ctx, client, repoOwner, repoName, pathInRepo, &github.RepositoryContentFileOptions{
		Message: github.String(commitMessage),
		Content: []byte(newYaml),
		Branch:  github.String(commitBranch),
		SHA:     SHA,
	})
	if err != nil {
		log.Error().Err(err).Msgf("could not update file in %s", repoName)
		return err
	}

	// create pull req
	newPR, err := gh.CreatePR(ctx, client, repoOwner, repoName, repoConfig.Reviewers, repoConfig.Assignees, repoConfig.Labels, &github.NewPullRequest{
		Title:               github.String(commitMessage),
		Head:                github.String(commitBranch),
		Base:                github.String(targetBranch),
		Body:                github.String(commitMessage),
		MaintainerCanModify: github.Bool(true),
	})
	if err != nil {
		log.Info().Err(err).Msgf("could not create pr in %s", repoName)
		return err
	}

	go func() {
		datadog.IncrementCount("pull_requests", 1, []string{fmt.Sprintf("organization:%s", repoOwner)})
	}()

	if oldPRs != nil || len(oldPRs) > 0 {
		gh.CleanUpOldBranch(ctx, client, repoOwner, repoName, oldPRs, newPR.GetNumber())

		go func() {
			datadog.IncrementCount("superseeded_updates", 1, []string{fmt.Sprintf("organization:%s", repoOwner)})
		}()
	}
	return nil
}

func getRelativeDirDepth(root, current string) int {
	return strings.Count(root, string(os.PathSeparator)) - strings.Count(current, string(os.PathSeparator))
}
