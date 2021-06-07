package config

import "dagger.io/dagger"

type Image struct {
	RootImage    string
	Distribution string

	ImageName   string
	TagPrefixes []string
}

type Platform struct {
	Name   dagger.Platform
	GoArch string
}

type Entity struct {
	ID   int
	Name string
}

type GithubRepo struct {
	Organization string
	Project      string
}

// images

var DefaultVersion = "develop"
var DefaultImage = DebianTrixieImage

var DebianTrixieImage = &Image{
	RootImage:    "debian:trixie-slim",
	Distribution: "debian",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"debian-trixie", "debian-13", "debian"},
}

var DebianBookwormImage = &Image{
	RootImage:    "debian:bookworm-slim",
	Distribution: "debian",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"debian-bookworm", "debian-12"},
}

var DebianBullseyeImage = &Image{
	RootImage:    "debian:bullseye-slim",
	Distribution: "debian",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"debian-bullseye", "debian-11"},
}

var Ubuntu2404Image = &Image{
	RootImage:    "ubuntu:24.04",
	Distribution: "ubuntu",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"ubuntu-24.04", "ubuntu-noble", "ubuntu"},
}

var Ubuntu2204Image = &Image{
	RootImage:    "ubuntu:22.04",
	Distribution: "ubuntu",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"ubuntu-22.04", "ubuntu-jammy"},
}

var Alpine323Image = &Image{
	RootImage:    "alpine:3.23.3",
	Distribution: "alpine",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"alpine-3.23", "alpine-3", "alpine"},
}

var Alpine322Image = &Image{
	RootImage:    "alpine:3.22.3",
	Distribution: "alpine",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"alpine-3.22"},
}

var Alpine321Image = &Image{
	RootImage:    "alpine:3.21.6",
	Distribution: "alpine",

	ImageName:   "osixia/baseimage",
	TagPrefixes: []string{"alpine-3.21"},
}

var Images = []*Image{
	DebianTrixieImage,
	DebianBookwormImage,
	DebianBullseyeImage,

	Ubuntu2404Image,
	Ubuntu2204Image,

	Alpine323Image,
	Alpine322Image,
	Alpine321Image,
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

// nonroot

var NonrootUser = &Entity{
	ID:   65532,
	Name: "nonroot",
}

var NonrootGroup = &Entity{
	ID:   65532,
	Name: "nonroot",
}

// github repo

var ProjectGithubRepo = &GithubRepo{
	Organization: "osixia",
	Project:      "container-baseimage",
}
