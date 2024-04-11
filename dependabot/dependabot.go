package dependabot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/BESTSELLER/dependabot-circleci/circleci"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/google/go-github/v60/github"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type FileUpdate struct {
	SHA     *string
	Content *string
	Node    *yaml.Node
}
type Update struct {
	Type        string
	CurrentName string
	FileUpdates map[string]FileUpdate
}

type RepoInfo struct {
	repoConfig        *config.RepoConfig
	repoOwner         string
	repoDefaultBranch string
	targetBranch      string
	repoName          string
}

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
	// should we then remove the repo from our db ?
	if repo.GetArchived() {
		log.Debug().Msg(fmt.Sprintf("Repo '%s' is archived", repo.GetName()))
		return
	}

	repoInfo, err := getRepoInfo(ctx, client, repo)
	if err != nil {
		log.Debug().Err(err).Msgf("could not get repo info for repo %s", repo.GetName())
		return
	}

	go datadog.IncrementCount("analysed_repos", 1, []string{fmt.Sprintf("organization:%s", repoInfo.repoOwner)})
	updates := map[string]Update{}
	var wg sync.WaitGroup
	wg.Add(1)
	go gatherUpdates(&wg, ctx, client, repoInfo, repoInfo.repoConfig.ConfigPath, &updates)
	wg.Wait()

	for newVerName, updateInfo := range updates {
		oldBranch, prBranch, err := handleBranch(ctx, client, repoInfo, newVerName, updateInfo)
		if err != nil || oldBranch {
			return
		}
		prTitle := generatePRTitle(updateInfo, newVerName)
		exists, oldPRs, err := gh.CheckPR(ctx, client, repoInfo.repoOwner, repoInfo.repoName, prTitle, updateInfo.CurrentName)
		if err != nil {
			log.Error().Err(err).Msgf("could not get old branch in %s", repoInfo.repoName)
			return
		}
		if exists {
			return
		}

		err = handleUpdate(ctx, client, repoInfo, prBranch, &updateInfo)
		if err != nil {
			go datadog.IncrementCount("failed_repos", 1, []string{fmt.Sprintf("organization:%s", repoInfo.repoOwner)})
			return
		}
		prNumber, err := handlePR(ctx, client, repoInfo, prBranch, prTitle)
		if err != nil {
			return
		}
		if oldPRs != nil || len(oldPRs) > 0 {
			gh.CleanUpOldBranch(ctx, client, repoInfo.repoOwner, repoInfo.repoName, oldPRs, prNumber)
			go func() {
				datadog.IncrementCount("superseeded_updates", 1, []string{fmt.Sprintf("organization:%s", repoInfo.repoOwner)})
			}()
		}
	}

}

func handlePR(ctx context.Context, client *github.Client, info *RepoInfo, branchName, prTitle string) (int, error) {
	// create pull req
	newPR, err := gh.CreatePR(ctx, client, info.repoOwner, info.repoName, info.repoConfig.Reviewers, info.repoConfig.Assignees, info.repoConfig.Labels, &github.NewPullRequest{
		Title:               github.String(prTitle),
		Head:                github.String(branchName),
		Base:                github.String(info.targetBranch),
		Body:                github.String(branchName),
		MaintainerCanModify: github.Bool(true),
	})
	if err != nil {
		log.Info().Err(err).Msgf("could not create pr in %s", info.repoName)
		return -1, err
	}

	go func() {
		datadog.IncrementCount("pull_requests", 1, []string{fmt.Sprintf("organization:%s", info.repoOwner)})
	}()
	return newPR.GetNumber(), nil
}

func getRepoInfo(ctx context.Context, client *github.Client, repo *github.Repository) (*RepoInfo, error) {
	// defer wg.Done()
	repoName := repo.GetName()

	log.Debug().Msg(fmt.Sprintf("Checking repo: %s", repoName))

	// Use this to test the application against a single repo
	// if repoName != "bestone-bi4-sales-core-salesorderservice" {
	// 	return
	// }

	repoConfig := getRepoConfig(ctx, client, repo)
	if repoConfig == nil {
		return nil, errors.New(fmt.Sprintf("could not get repo config for repo %s", repoName))
	}

	// determine repo details
	repoOwner := repo.GetOwner().GetLogin()
	repoDefaultBranch := repo.GetDefaultBranch()
	targetBranch := getTargetBranch(ctx, client, repoOwner, repoName, repoDefaultBranch, repoConfig)
	if targetBranch == "" {
		return nil, errors.New(fmt.Sprintf("could not get targetBranch for repo %s", repoName))
	}
	return &RepoInfo{
		repoConfig:        repoConfig,
		repoOwner:         repoOwner,
		repoDefaultBranch: repoDefaultBranch,
		targetBranch:      targetBranch,
		repoName:          repoName,
	}, nil
}

func handleBranch(ctx context.Context, client *github.Client, repoInfo *RepoInfo, newName string, updateInfo Update) (existed bool, branchName string, err error) {
	commitBranch := fmt.Sprintf("dependabot-circleci/%s/%s", updateInfo.Type, strings.ReplaceAll(newName, ":", "@"))
	notExists := gh.CheckBranch(ctx, client, repoInfo.repoOwner, repoInfo.repoName, github.String(commitBranch))
	if notExists {
		err = gh.CreateBranch(ctx, client, repoInfo.repoOwner, repoInfo.repoName, repoInfo.targetBranch, github.String(commitBranch))
		if err != nil {
			log.Error().Err(err).Msgf("could not create branch %s in %s", commitBranch, repoInfo.repoName)
			return false, "", err
		}
	} else {
		log.Debug().Msgf("branch %s already exists, skipping creation of branch", commitBranch)
		return true, commitBranch, nil
	}
	return false, commitBranch, nil
}

func gatherUpdates(wg *sync.WaitGroup, ctx context.Context, client *github.Client, repoInfo *RepoInfo, pathInRepo string, updates *map[string]Update) {
	defer wg.Done()
	log.Info().Msgf("Processing: %s", pathInRepo)
	// 1. Get directory contents
	options := &github.RepositoryContentGetOptions{Ref: repoInfo.targetBranch}
	fileContent, directoryContent, _, err := client.Repositories.GetContents(context.Background(), repoInfo.repoOwner, repoInfo.repoName, pathInRepo, options)
	if err != nil {
		log.Error().Err(err).Msgf("could not parseRepoContent %s", repoInfo.repoName)
		return
	}
	if fileContent == nil {
		for _, dir := range directoryContent {
			if !repoInfo.repoConfig.IsWithinScanDepth(dir.GetPath()) {
				return
			}
			wg.Add(1)
			go gatherUpdates(wg, ctx, client, repoInfo, dir.GetPath(), updates)
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
		log.Error().Err(err).Msgf("could not fileContent.GetContent() %s", repoInfo.repoName)
	}
	// unmarshal
	var cciconfig yaml.Node
	err = yaml.Unmarshal([]byte(content), &cciconfig)
	if err != nil {
		log.Error().Err(err).Msgf("could not unmarshal yaml in %s", repoInfo.repoName)
		return
	}

	// check for updates
	orbUpdates, dockerUpdates := circleci.GetUpdates(&cciconfig)
	for old, update := range orbUpdates {
		if _, contains := (*updates)[update.Value]; !contains {
			(*updates)[update.Value] = Update{
				Type:        "orb",
				CurrentName: old,
				FileUpdates: make(map[string]FileUpdate),
			}
		}
		(*updates)[update.Value].FileUpdates[fileContent.GetPath()] = FileUpdate{
			SHA:     fileContent.SHA,
			Node:    update,
			Content: &content,
		}
	}
	for old, update := range dockerUpdates {
		if _, contains := (*updates)[update.Value]; !contains {
			(*updates)[update.Value] = Update{
				Type:        "docker",
				CurrentName: old,
				FileUpdates: make(map[string]FileUpdate),
			}
		}
		(*updates)[update.Value].FileUpdates[fileContent.GetPath()] = FileUpdate{
			SHA:     fileContent.SHA,
			Node:    update,
			Content: &content,
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

func handleUpdate(ctx context.Context, client *github.Client, info *RepoInfo, prBranch string, updates *Update) error {
	for path, update := range updates.FileUpdates {
		log.Debug().Msgf("repo: %s, file%s, old: %s, update: %s", info.repoName, path, updates.CurrentName, update.Node.Value)
		newYaml := circleci.ReplaceVersion(update.Node, updates.CurrentName, *update.Content)

		// commit vars
		var oldVersion, newVersion []string
		var separator string
		if updates.Type == "orb" {
			separator = "@"
		} else {
			separator = ":"
		}
		oldVersion = strings.Split(updates.CurrentName, separator)
		newVersion = strings.Split(update.Node.Value, separator)

		if updates.Type == "orb" && len(newVersion) == 1 {
			return fmt.Errorf("could not find orb version for %s in %s", update.Node.Value, info.repoName)
		}

		commitMessage := fmt.Sprintf("Update %s, @%s from %s to %s", path, oldVersion[0], oldVersion[1], newVersion[1])

		// commit file
		err := gh.UpdateFile(ctx, client, info.repoOwner, info.repoName, path, &github.RepositoryContentFileOptions{
			Message: github.String(commitMessage),
			Content: []byte(newYaml),
			Branch:  github.String(prBranch),
			SHA:     update.SHA,
		})
		if err != nil {
			log.Error().Err(err).Msgf("could not update file in %s", info.repoName)
			return err
		}
	}
	return nil
}

func generatePRTitle(update Update, newName string) string {
	var oldVersion, newVersion []string
	var separator string
	if update.Type == "orb" {
		separator = "@"
	} else {
		separator = ":"
	}
	oldVersion = strings.Split(update.CurrentName, separator)
	newVersion = strings.Split(newName, separator)

	return fmt.Sprintf("Bump @%s from %s to %s", oldVersion[0], oldVersion[1], newVersion[1])
}
