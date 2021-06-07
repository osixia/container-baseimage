package helpers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

	previousRelease, _, err := client.Repositories.GetReleaseByTag(ctx, repo.Organization, repo.Project, release.GetTagName())
	if err != nil {
		var githubErr *github.ErrorResponse
		if !errors.As(err, &githubErr) || githubErr.Response == nil || githubErr.Response.StatusCode != http.StatusNotFound {
			return nil, err
		}
	}

	if previousRelease == nil {
		fmt.Printf("create release %v\n", release.GetTagName())
		if !dryRun {
			release, _, err = client.Repositories.CreateRelease(ctx, repo.Organization, repo.Project, release)
			if err != nil {
				return nil, err
			}
		}
	} else {
		fmt.Printf("updating existing %v release\n", previousRelease.GetTagName())
		if !dryRun {
			release, _, err = client.Repositories.EditRelease(ctx, repo.Organization, repo.Project, previousRelease.GetID(), release)
			if err != nil {
				return nil, err
			}
		}
	}

	return release, nil
}

func GithubDeleteReleaseAssetsByName(ctx context.Context, client *github.Client, repo *config.GithubRepo, releaseID int64, name string, dryRun bool) error {

	opts := &github.ListOptions{PerPage: 100}

	for {
		assets, resp, err := client.Repositories.ListReleaseAssets(ctx, repo.Organization, repo.Project, releaseID, opts)
		if err != nil {
			return err
		}

		for _, asset := range assets {
			if asset.GetName() != name {
				continue
			}

			if dryRun {
				fmt.Printf("would have deleted existing asset %v\n", name)
				continue
			}

			fmt.Printf("deleting existing asset %v\n", name)
			if _, err := client.Repositories.DeleteReleaseAsset(ctx, repo.Organization, repo.Project, asset.GetID()); err != nil {
				return err
			}
		}

		if resp.NextPage == 0 {
			return nil
		}

		opts.Page = resp.NextPage
	}
}

func GithubUploadReleaseFiles(ctx context.Context, client *github.Client, repo *config.GithubRepo, release *github.RepositoryRelease, brs []*job.BuildResult, dryRun bool) error {

	for _, br := range brs {
		for _, f := range br.Files {

			fn := filepath.Base(f)

			if dryRun {
				fmt.Printf("would have uploaded %v\n", fn)
				continue
			}

			file, err := os.Open(f)
			if err != nil {
				return fmt.Errorf("error opening %v: %w", f, err)
			}

			if err := GithubDeleteReleaseAssetsByName(ctx, client, repo, release.GetID(), fn, dryRun); err != nil {
				_ = file.Close()
				return fmt.Errorf("error deleting existing release asset %v: %w", fn, err)
			}

			opts := &github.UploadOptions{Name: fn}
			asset, _, err := client.Repositories.UploadReleaseAsset(ctx, repo.Organization, repo.Project, release.GetID(), opts, file)
			_ = file.Close()
			if err != nil {
				return fmt.Errorf("error uploading release asset: %w", err)
			}

			fmt.Printf("asset uploaded: %v\n", asset.GetBrowserDownloadURL())
		}
	}

	return nil
}
