// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// File provides access to file system information.
type File struct{}

// NewFile creates a new File resource.
func NewFile() *File {
	return &File{}
}

// Name returns the resource identifier for files.
func (r *File) Name() string {
	return "file"
}

// Fetch retrieves file metadata for the path specified in asset configuration.
func (r *File) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	path, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("missing 'path' in config for file resource")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, err
	}

	res := core.Resource{
		"name":       info.Name(),
		"size":       info.Size(),
		"mode":       info.Mode().String(),
		"mode_octal": fmt.Sprintf("%04o", info.Mode().Perm()),
		"is_dir":     info.IsDir(),
		"mod_time":   info.ModTime().UTC().Format(time.RFC3339),
		"path":       path,
		"exists":     true,
	}

	return []core.Resource{res}, nil
}
