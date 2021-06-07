package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/google/go-github/v69/github"
	"go.uber.org/thriftrw/ptr"

	"github.com/osixia/container-baseimage/build/config"
	"github.com/osixia/container-baseimage/build/job"
)

func GithubGetTags(ctx context.Context, client *github.Client, repo *config.GithubRepo) ([]string, error) {

	// get all tags
	topt := &github.ListOptions{}

	var tags []string
	for {

		ts, r, err := client.Repositories.ListTags(ctx, repo.Organization, repo.Project, topt)
		if err != nil {
			return nil, err
		}

		for _, t := range ts {
			tags = append(tags, *t.Name)
		}

		if r.NextPage == 0 {
			break
		}

		topt.Page = r.NextPage
	}

	// sort results
	sort.Strings(tags)

	return tags, nil
}

func GithubCreateOrUpdateRelease(ctx context.Context, client *github.Client, repo *config.GithubRepo, version string, body string, dryRun bool) (*github.RepositoryRelease, error) {

	fmt.Printf("working on release %v\n", version)

	release := &github.RepositoryRelease{
		TagName:    ptr.String(version),
		Name:       ptr.String(version),
		Body:       ptr.String(body),
		Prerelease: ptr.Bool(regexp.MustCompile(config.PrereleaseTagRegex).MatchString(version)),
	}

	var previousRelease *github.RepositoryRelease

	// check if the release already exists
	releases, _, err := client.Repositories.ListReleases(ctx, repo.Organization, repo.Project, nil)
	if err != nil {
		return nil, err
	}
	for _, r := range releases {
		if r.GetTagName() == release.GetTagName() {
			previousRelease = r
		}
	}

	if previousRelease != nil {
		fmt.Printf("updating existing %v release\n", previousRelease.GetTagName())
		if !dryRun {
			release, _, err = client.Repositories.EditRelease(ctx, repo.Organization, repo.Project, previousRelease.GetID(), release)
			if err != nil {
				return nil, err
			}
		}
	} else {
		fmt.Printf("create release %v\n", release.GetTagName())
		if !dryRun {
			release, _, err = client.Repositories.CreateRelease(ctx, repo.Organization, repo.Project, release)
			if err != nil {
				return nil, err
			}
		}
	}

	return release, nil
}

func GithubUploadReleaseFiles(ctx context.Context, client *github.Client, repo *config.GithubRepo, release *github.RepositoryRelease, brs []*job.BuildResult, dryRun bool) error {

	for _, br := range brs {
		for _, f := range br.Files {
			file, err := os.Open(f)
			if err != nil {
				return fmt.Errorf("error opening %v: %w", f, err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					fmt.Print(err.Error())
				}
			}()

			fn := filepath.Base(file.Name())

			if dryRun {
				fmt.Printf("would have uploaded %v\n", fn)
				continue
			}

			opts := &github.UploadOptions{Name: fn}
			asset, _, err := client.Repositories.UploadReleaseAsset(ctx, repo.Organization, repo.Project, release.GetID(), opts, file)
			if err != nil {
				return fmt.Errorf("error uploading release asset: %w", err)
			}

			fmt.Printf("asset uploaded: %v\n", asset.GetBrowserDownloadURL())
		}
	}

	return nil
}
