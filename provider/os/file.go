package os

import (
	"context"
	"fmt"
	"os"

	"github.com/juliankoehn/kspec/core"
)

type FileResource struct{}

func (r *FileResource) Name() string {
	return "file"
}

func (r *FileResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Use config from asset
	config := asset.Config
	path, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("missing 'path' in config for file resource")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Should we return an empty list or an error?
			// Usually for "check if file exists", we might want to return a resource with "exists: false"
			// But for now, let's return error or empty.
			// If the user query is 'exists(resource)', we need a resource.
			// Let's return error for now to keep it simple.
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
		"mod_time":   info.ModTime().Format(map[string]string{"iso": "2006-01-02T15:04:05Z"}["iso"]), // simple iso format
		"path":       path,
		"exists":     true,
	}

	return []core.Resource{res}, nil
}
