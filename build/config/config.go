package config

import (
	"fmt"

	"dagger.io/dagger"
)

// distributions

var Debian = "debian"
var Alpine = "alpine"
var Ubuntu = "ubuntu"

var Distributions = []string{
	Debian,
	Alpine,
	Ubuntu,
}

// image

type Image struct {
	BaseImage    string
	Distribution string

	Name        string
	Description string
	TagPrefixes []string

	Url           string
	Documentation string
	Source        string

	Authors string
	Vendor  string

	Licences string
}

func (i *Image) Validate() (bool, error) {

	if i.BaseImage == "" {
		return false, fmt.Errorf("image %v: BaseImage required", i)
	}

	if i.Distribution == "" {
		return false, fmt.Errorf("image %v: Distribution required", i)
	}

	if i.Name == "" {
		return false, fmt.Errorf("image %v: Name required", i)
	}

	return true, nil
}

var DefaultVersion = "develop"
var DefaultImage *Image
var Images []*Image

// platforms

type Platform struct {
	Name   dagger.Platform
	GoArch string
}

var Amd64Platform = &Platform{
	Name:   "linux/amd64",
	GoArch: "amd64",
}

var Arm64Platform = &Platform{
	Name:   "linux/arm64",
	GoArch: "arm64",
}

var Platforms = []*Platform{
	Amd64Platform,
	Arm64Platform,
}

// nonroot

type Entity struct {
	ID   int
	Name string
}

var NonrootUser = &Entity{
	ID:   65532,
	Name: "nonroot",
}

var NonrootGroup = &Entity{
	ID:   65532,
	Name: "nonroot",
}

// build

var ExcludedBuildPaths = []string{".git/", ".github/", "bin/", "build/", "docs/", ".dockerignore", ".gitignore", "Dockerfile", "**/.gitkeep", "**/*.md"}

// github repo

type GithubRepo struct {
	Organization string
	Project      string
}

var ProjectGithubRepo *GithubRepo

var TagRegex = `^[0-9]+\.[0-9]+\.[0-9]+-?[^.]*$`                 // x.y.z or x.y.z-a with x, y and z numbers and a any char except '.'
var ReleaseTagRegex = `^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9]+)?$`     //  x.y.z or x.y.z-a with x, y, z and a numbers
var PrereleaseTagRegex = `^[0-9]+\.[0-9]+\.[0-9]+-[^.\[0-9\]]*$` // x.y.z-a with x,y and z numbers and a word character
