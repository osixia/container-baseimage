package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/build/config"
)

type ExportOptions struct {
	BuildOptions

	ExportTo   string
	ExportFile string

	ExportPerArch bool

	ExportCompress      bool
	ExportCompressImage string
}

func DefaultExport(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {

	o := options.(*ExportOptions)

	builds, err := Build(ctx, client, &o.BuildOptions)
	if err != nil {
		return nil, err
	}

	if o.ExportTo == "" || o.ExportFile == "" {
		return builds, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	exportDir, err := filepath.Abs(filepath.Join(wd, o.ExportTo))
	if err != nil {
		return nil, err
	}

	// create empty directory to put build outputs
	outputs := client.Directory()

	// export file from default image in outputs directory
	for _, b := range builds {

		if b.Image != config.DefaultImage {
			continue
		}

		for _, c := range b.Containers {

			file := c.File(o.ExportFile)
			fileName := filepath.Base(o.ExportFile)

			if o.ExportPerArch {
				p, err := c.Platform(ctx)
				if err != nil {
					return nil, err
				}
				ps := strings.ReplaceAll(string(p), "/", "_")

				name, ext, found := strings.Cut(fileName, ".")
				if found {
					ext = "." + ext
				}

				fileName = fmt.Sprintf("%v_%v%v", name, ps, ext)
			}

			if o.ExportCompress && o.ExportCompressImage != "" {
				tarName := fmt.Sprintf("%v.tar.gz", fileName)

				// container to compress bin
				tar := client.Container().From(o.ExportCompressImage)

				tar, _ = tar.WithFile(fileName, file).Sync(ctx)
				tar = tar.WithExec([]string{"tar", "-czf", tarName, fileName})

				file = tar.File(tarName)
				fileName = tarName
			}

			outputs = outputs.WithFile(fileName, file)

			b.Files = append(b.Files, filepath.Join(exportDir, fileName))
		}

		break
	}

	// export outputs directory on filesystem
	_, err = outputs.Export(ctx, exportDir)
	if err != nil {
		return nil, err
	}

	return builds, nil
}
