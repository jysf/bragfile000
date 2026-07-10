package story

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed profiles/*.yaml directives/*.md
var assetsFS embed.FS

// bundledProfile returns the raw bytes of the embedded profile asset
// profiles/<name>.yaml and whether it exists. The audience set is thus
// the embedded FS + the override dir — NOT a Go enum (DEC-029 choice 2).
func bundledProfile(name string) ([]byte, bool) {
	b, err := assetsFS.ReadFile("profiles/" + name + ".yaml")
	if err != nil {
		return nil, false
	}
	return b, true
}

// directiveAsset returns the exact byte content of the embedded framing
// directive directives/<basename>. Errors (wrapped with context) if the
// asset is missing.
func directiveAsset(basename string) ([]byte, error) {
	b, err := assetsFS.ReadFile("directives/" + basename)
	if err != nil {
		return nil, fmt.Errorf("read directive asset %q: %w", basename, err)
	}
	return b, nil
}

// bundledProfileNames lists the audience names present as embedded
// profile assets (basename without the .yaml suffix). Used to prove the
// set is data-driven — no Go type enumerates it.
func bundledProfileNames() ([]string, error) {
	ents, err := fs.ReadDir(assetsFS, "profiles")
	if err != nil {
		return nil, fmt.Errorf("read embedded profiles dir: %w", err)
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		n := strings.TrimSuffix(e.Name(), ".yaml")
		names = append(names, n)
	}
	return names, nil
}
