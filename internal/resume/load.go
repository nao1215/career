package resume

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// ErrEmptyName is returned when the profile name is missing. Both documents
// print the name prominently, so it is the one universally required field.
var ErrEmptyName = errors.New("profile.name is required")

// Load reads and parses a resume YAML file from path.
func Load(path string) (*Resume, error) {
	f, err := os.Open(path) //nolint:gosec // path is a user-supplied input file by design
	if err != nil {
		return nil, fmt.Errorf("open resume file: %w", err)
	}
	defer f.Close()
	return Parse(f)
}

// Parse decodes a resume document from r. Unknown YAML keys are rejected so
// that typos in field names surface as errors instead of being silently
// ignored.
func Parse(r io.Reader) (*Resume, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)

	var res Resume
	if err := dec.Decode(&res); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("resume document is empty")
		}
		return nil, fmt.Errorf("parse resume yaml: %w", err)
	}
	return &res, nil
}

// ValidateRireki checks that the document has the minimum data needed to render
// a 履歴書.
func (r *Resume) ValidateRireki() error {
	if !r.Profile.Name.Has() {
		return ErrEmptyName
	}
	return nil
}

// ValidateCareer checks that the document has the minimum data needed to render
// a 職務経歴書 or a CV.
func (r *Resume) ValidateCareer() error {
	if !r.Profile.Name.Has() {
		return ErrEmptyName
	}
	if !r.Career.Summary.Has() && len(r.Career.Histories) == 0 {
		return errors.New("career.summary or career.histories is required for a 職務経歴書 or CV")
	}
	return nil
}
