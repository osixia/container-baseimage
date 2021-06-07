package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"dagger.io/dagger"
	"github.com/google/go-github/v69/github"

	"github.com/osixia/container-baseimage/build/config"
	"github.com/osixia/container-baseimage/build/helpers"
	"github.com/osixia/container-baseimage/build/job"
)

type GithubOptions struct {
	job.PublishOptions

	Ref   string
	Token string

	Changelog string
}

func Github(ctx context.Context, client *dagger.Client, options any) ([]*job.BuildResult, error) {
	o := options.(*GithubOptions)

	// GITHUB_REF values:
	// refs/heads/<branch_name>, refs/pull/<pr_number>/merge, refs/tags/<tag_name>

	// Build and test: branches main, develop, feature/*, bugfix/*, release/*, hotfix/*,  support/*
	// Build, test and deploy: tags

	ref := regexp.MustCompile(`^refs/(heads|tags|pull)/(.*)$`).FindStringSubmatch(o.Ref)
	if ref == nil || (len(ref) != 3) {
		return nil, fmt.Errorf("unable to get github reference type and name %v", ref)
	}

	refType := ref[1]
	refName := ref[2]

	o.Version = refName

	githubClient := github.NewClient(nil).WithAuthToken(o.Token)

	// not a tag run only test job
	if refType != "tags" {
		if refType == "pull" {
			o.Version = strings.TrimSuffix(o.Version, "/merge")
		}

		return job.Test(ctx, client, &o.TestOptions)
	}

	if !regexp.MustCompile(config.TagRegex).MatchString(o.Version) {
		return nil, fmt.Errorf("%v tag must match %v", o.Version, config.TagRegex)
	}

	gtags, err := helpers.GithubGetTags(ctx, githubClient, config.ProjectGithubRepo)
	if err != nil {
		return nil, err
	}

	o.Latest, err = helpers.IsLatestTag(o.Version, gtags)
	if err != nil {
		return nil, err
	}

	body := ""
	if o.Changelog != "" {
		var err error
		body, err = helpers.ChangelogGetVersionSection(o.Changelog, o.Version)
		if err != nil {
			return nil, fmt.Errorf("error extracting changelog section: %w", err)
		}
	}

	brs, err := job.Deploy(ctx, client, &o.DeployOptions)
	if err != nil {
		return nil, err
	}

	release, err := helpers.GithubCreateOrUpdateRelease(ctx, githubClient, config.ProjectGithubRepo, o.Version, body, o.DryRun)
	if err != nil {
		return nil, err
	}

	if err := helpers.GithubUploadReleaseFiles(ctx, githubClient, config.ProjectGithubRepo, release, brs, o.DryRun); err != nil {
		return nil, err
	}

	return brs, nil
}
