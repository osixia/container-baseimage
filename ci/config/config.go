package config

import "dagger.io/dagger"

type Image struct {
	RootImageName string
	RootImageTag  string

	BuildImageName string
	TagPrefixes    []string
}

type Platform struct {
	Name   dagger.Platform
	GoArch string
}

type GithubRepo struct {
	Organization string
	Project      string
}

// images

var DefaultVersion = "develop"
var DefaultImage = DebianBookwormImage

var DebianBookwormImage = &Image{
	RootImageName: "debian",
	RootImageTag:  "bookworm-slim",

	BuildImageName: "osixia/baseimage",
	TagPrefixes:    []string{"debian-bookworm", "debian"},
}

var DebianBullseyeImage = &Image{
	RootImageName: "debian",
	RootImageTag:  "bullseye-slim",

	BuildImageName: "osixia/baseimage",
	TagPrefixes:    []string{"debian-bullseye"},
}

var Ubuntu2204Image = &Image{
	RootImageName: "ubuntu",
	RootImageTag:  "22.04",

	BuildImageName: "osixia/baseimage",
	TagPrefixes:    []string{"ubuntu-22.04", "ubuntu"},
}

var Alpine319Image = &Image{
	RootImageName: "alpine",
	RootImageTag:  "3.19.0",

	BuildImageName: "osixia/baseimage",
	TagPrefixes:    []string{"alpine-3.19", "alpine-3", "alpine"},
}

var Alpine318Image = &Image{
	RootImageName: "alpine",
	RootImageTag:  "3.18.5",

	BuildImageName: "osixia/baseimage",
	TagPrefixes:    []string{"alpine-3.18"},
}

var Images = []*Image{
	DebianBookwormImage,
	DebianBullseyeImage,
	Ubuntu2204Image,
	Alpine319Image,
	Alpine318Image,
}

// platforms

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

// Github repo
var BaseimageGithubRepo = &GithubRepo{
	Organization: "osixia",
	Project:      "container-baseimage",
}
