package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-github/v69/github"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"go.uber.org/thriftrw/ptr"

	"github.com/osixia/container-baseimage/build/config"
	"github.com/osixia/container-baseimage/build/job"
)

type githubFlags struct {
	deployFlags
}

type githubData struct {
	contributors []string
	tags         []string
}

var githubCmdFlags = &githubFlags{}

var githubCmd = &cobra.Command{
	Use:   "github dockerfile github_ref github_token",
	Short: "Run jobs based on GitHub Actions CI/CD environment",

	GroupID: pipelineGroupID,

	Args: cobra.ExactArgs(4),

	Run: func(cmd *cobra.Command, args []string) {

		dockerfile := args[0]
		changelog := args[1]
		githubRef := args[2]
		githubToken := args[3]

		fmt.Printf("Dockerfile: %v, changelog: %v, githubRef: %v\n", dockerfile, changelog, githubRef)

		testFlags := testFlags{
			buildFlags: buildFlags{
				dockerfile: dockerfile,

				distributions: distributions(),
				arches:        arches(),

				withNonroot: true,
			},
		}

		jb := job.Test
		var jFlags jobFlags = &testFlags

		// GITHUB_REF values:
		// refs/heads/<branch_name>, refs/pull/<pr_number>/merge, refs/tags/<tag_name>

		// Build and test: branches main, develop, feature/*, bugfix/*, release/*, hotfix/*,  support/*
		// Build, test and deploy: tags

		ref := regexp.MustCompile(`^refs/(heads|tags|pull)/(.*)$`).FindStringSubmatch(githubRef)
		if ref == nil || (len(ref) != 3) {
			fatal(fmt.Errorf("unable to get github reference type and name %v", ref))
		}

		refType := ref[1]
		refName := ref[2]

		testFlags.version = refName

		githubClient := github.NewClient(nil).WithAuthToken(githubToken)

		switch refType {
		case "pull":

			testFlags.version = strings.TrimSuffix(refName, "/merge")

		case "tags": // deploy

			if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+-?[^.]*$`).MatchString(refName) {
				fatal(fmt.Errorf("%v tag must be formated like x.y.z or x.y.z-a with x, y and z numbers and a any char except '.'", refName))
			}

			gi, err := githubDataRequest(cmd.Context(), githubClient, config.ProjectGithubRepo)
			if err != nil {
				fatal(err)
			}

			testFlags.latest, err = isLatestTag(testFlags.version, gi.tags)
			if err != nil {
				fatal(err)
			}

			deployFlags := githubCmdFlags.deployFlags
			deployFlags.testFlags = testFlags
			deployFlags.contributors = strings.Join(gi.contributors, ", ")

			jb = job.Deploy
			jFlags = &deployFlags

		}

		brs, err := job.Run(cmd.Context(), jb, jFlags.toJobOptions())
		if err != nil {
			fatal(err)
		}

		if jb == job.Deploy {
			if err := githubReleaseRequest(cmd.Context(), githubClient, config.ProjectGithubRepo, jFlags.(*deployFlags), brs, changelog); err != nil {
				fatal(err)
			}
		}

	},
}

func init() {
	// flags
	githubCmd.Flags().SortFlags = false
	addDeployFlags(githubCmd.Flags(), &githubCmdFlags.deployFlags)
}

func isLatestTag(tag string, existingTags []string) (bool, error) {

	// test latest only on x.y.z tags (not x.y.z-a)
	if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(tag) {
		return false, nil
	}

	tVersion, err := version.NewVersion(tag)
	if err != nil {
		return false, err
	}

	// remove prefixes from github tags
	etVersions := make([]*version.Version, 0, len(existingTags))
	for _, gtag := range existingTags {
		m := regexp.MustCompile(`^.*([0-9]+\.[0-9]+\.[0-9]+)-?.*$`).FindStringSubmatch(gtag)
		if len(m) < 2 {
			return false, fmt.Errorf("error: parsing tag: %v", gtag)
		}

		v, err := version.NewVersion(m[1])
		if err != nil {
			return false, err
		}
		etVersions = append(etVersions, v)
	}

	// sort github tags
	sort.Sort(version.Collection(etVersions))

	// compare new tag version and greated github tag version
	return tVersion.GreaterThanOrEqual(etVersions[len(etVersions)-1]), nil
}

func githubDataRequest(ctx context.Context, client *github.Client, repo *config.GithubRepo) (*githubData, error) {

	// get all contributors
	copt := &github.ListContributorsOptions{}

	var contributors []string
	for {
		cbs, r, err := client.Repositories.ListContributors(ctx, repo.Organization, repo.Project, copt)
		if err != nil {
			return nil, err
		}

		for _, c := range cbs {
			contributors = append(contributors, *c.Login)
		}

		if r.NextPage == 0 {
			break
		}

		copt.Page = r.NextPage
	}

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

	return &githubData{
		contributors: contributors,
		tags:         tags,
	}, nil
}

func extractChangelogSection(changelog, version string) (string, error) {
	f, err := os.Open(changelog)
	if err != nil {
		return "", err
	}
	defer f.Close()

	startPrefix := "## [" + version + "]"
	sectionHeaderPrefix := "## ["

	var b strings.Builder
	inSection := false

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()

		if !inSection && strings.HasPrefix(line, startPrefix) {
			inSection = true
			b.WriteString(line)
			b.WriteByte('\n')
			continue
		}

		if inSection {
			if strings.HasPrefix(line, sectionHeaderPrefix) && !strings.HasPrefix(line, startPrefix) {
				break
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}

	if !inSection {
		return "", nil
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func githubReleaseRequest(ctx context.Context, client *github.Client, repo *config.GithubRepo, df *deployFlags, brs []*job.BuildResult, changelog string) error {

	fmt.Printf("working on release %v\n", df.version)

	body, err := extractChangelogSection(changelog, df.version)
	if err != nil {
		return fmt.Errorf("error extracting changelog section: %w", err)
	}

	release := &github.RepositoryRelease{
		TagName:    ptr.String(df.version),
		Name:       ptr.String(df.version),
		Body:       ptr.String(body),
		Prerelease: ptr.Bool(regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+-[^.]*$`).MatchString(df.version)),
	}

	var previousRelease *github.RepositoryRelease

	// check if the release already exists
	releases, _, err := client.Repositories.ListReleases(ctx, repo.Organization, repo.Project, nil)
	if err != nil {
		return err
	}
	for _, r := range releases {
		if r.GetTagName() == release.GetTagName() {
			previousRelease = r
		}
	}

	if previousRelease != nil {
		fmt.Printf("updating existing %v release\n", previousRelease.GetTagName())
		if !df.dryRun {
			release, _, err = client.Repositories.EditRelease(ctx, repo.Organization, repo.Project, previousRelease.GetID(), release)
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("create release %v\n", release.GetTagName())
		if !df.dryRun {
			release, _, err = client.Repositories.CreateRelease(ctx, repo.Organization, repo.Project, release)
			if err != nil {
				return err
			}
		}
	}

	for _, br := range brs {
		for _, f := range br.Files {
			file, err := os.Open(f)
			if err != nil {
				return fmt.Errorf("error opening %s: %w", f, err)
			}
			defer file.Close()

			fn := filepath.Base(file.Name())

			if df.dryRun {
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
