package helpers

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/hashicorp/go-version"

	"github.com/osixia/container-baseimage/build/config"
)

func IsLatestTag(tag string, existingTags []string) (bool, error) {

	r := regexp.MustCompile(config.ReleaseTagRegex)

	if !r.MatchString(tag) {
		return false, nil
	}

	tVersion, err := version.NewVersion(tag)
	if err != nil {
		return false, err
	}

	etVersions := make([]*version.Version, 0, len(existingTags))
	for _, gtag := range existingTags {
		if !r.MatchString(gtag) {
			fmt.Printf("ignoring tag: %v\n", gtag)
			continue
		}

		v, err := version.NewVersion(gtag)
		if err != nil {
			return false, err
		}
		etVersions = append(etVersions, v)
	}

	if len(etVersions) == 0 {
		return true, nil
	}

	// sort existing tags
	sort.Sort(version.Collection(etVersions))

	// compare new tag version and greatest existing tag version
	return tVersion.GreaterThanOrEqual(etVersions[len(etVersions)-1]), nil
}
