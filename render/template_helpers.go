package render

import (
	"encoding/json"
	"html/template"
	"path/filepath"
	"sync"

	"github.com/gobuffalo/tags"
	"github.com/pkg/errors"
)

var assetsMutex = &sync.Mutex{}
var assetMap map[string]string

func loadManifest(manifest string) error {
	assetsMutex.Lock()
	defer assetsMutex.Unlock()

	err := json.Unmarshal([]byte(manifest), &assetMap)
	return err
}

func assetPathFor(file string) string {
	assetsMutex.Lock()
	defer assetsMutex.Unlock()

	filePath := assetMap[file]
	if filePath == "" {
		filePath = filepath.Join("/assets", file)
	}
	return filepath.ToSlash(filePath)
}

type helperTag struct {
	name string
	fn   func(string, tags.Options) template.HTML
}

func (s templateRenderer) addAssetsHelpers(helpers Helpers) Helpers {
	helpers["assetPath"] = func(file string) (string, error) {
		return s.assetPath(file)
	}

	ah := []helperTag{
		{"javascriptTag", jsTag},
		{"stylesheetTag", cssTag},
		{"imgTag", imgTag},
	}

	for _, h := range ah {
		func(h helperTag) {
			helpers[h.name] = func(file string, options tags.Options) (template.HTML, error) {
				if options == nil {
					options = tags.Options{}
				}
				f, err := s.assetPath(file)
				if err != nil {
					return "", errors.WithStack(err)
				}
				return h.fn(f, options), nil
			}
		}(h)
	}

	return helpers
}

func jsTag(src string, options tags.Options) template.HTML {
	if options["type"] == nil {
		options["type"] = "text/javascript"
	}

	options["src"] = src
	jsTag := tags.New("script", options)

	return jsTag.HTML()
}

func cssTag(href string, options tags.Options) template.HTML {
	if options["rel"] == nil {
		options["rel"] = "stylesheet"
	}

	if options["media"] == nil {
		options["media"] = "screen"
	}

	options["href"] = href
	cssTag := tags.New("link", options)

	return cssTag.HTML()
}

func imgTag(src string, options tags.Options) template.HTML {
	options["src"] = src
	imgTag := tags.New("img", options)

	return imgTag.HTML()
}
