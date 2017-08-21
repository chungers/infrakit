package types

import (
	"fmt"
	"net/url"
	"strings"
)

// Spec is the specification of the resource / object
type Spec struct {

	// Kind is the category of the resources and kind can have types  -- e.g. instance-aws/ec2-instance
	Kind string `json:"kind"`

	// Version is the name of the interface and version - instance/v0.1.0
	Version string `json:"version"`

	// Metadata is metadata / useful information about object
	Metadata Metadata `json:"metadata"`

	// Template is a template of the resource's properties
	Template *URL `json:"template,omitempty" yaml:",omitempty"`

	// Properties is the desired properties of the resource, if template is specified,
	// then the values of properties override the same fields in template.
	Properties *Any `json:"properties"`

	// Options is additional data for handling the object that is not intrinsic to the object itself
	// but is necessary for the controllers working with it.
	Options *Any `json:"options,omitempty" yaml:",omitempty"`

	// Depends is a list of dependencies that this spec needs to have resolved before instances can
	// be instantiated.
	Depends []Dependency `json:"depends,omitempty" yaml:",omitempty"`
}

// Validate checks the spec for validity
func (s Spec) Validate() error {
	if s.Kind == "" {
		return errMissingAttribute("kind")
	}
	if s.Version == "" {
		return errMissingAttribute("version")
	}
	if s.Metadata.Name == "" {
		return errMissingAttribute("metadata.name")
	}
	return nil
}

// Dependency models the reference and usage of another spec, by spec's Kind and Name, and a way
// to extract its properties, and how it's referenced via the alias in the Properties section of the dependent Spec.
type Dependency struct {

	// Kind is the Kind of the spec this spec depends on
	Kind string `json:"kind"`

	// Name is the Name of the spec this spec dependes on
	Name string `json:"name"`

	// Bind is an associative array of pointer to the fields in the object to a variable name that will be referenced
	// in the properties or template of the owning spec.
	Bind map[string]*Pointer `json:"bind"`
}

// Identity uniquely identifies an instance
type Identity struct {

	// ID is a unique identifier for the object instance.
	ID string `json:"id" yaml:"id"`
}

// Metadata captures label and descriptive information about the object
type Metadata struct {

	// Identity is an optional component that exists only in the case of a real object instance.
	*Identity `json:",inline,omitempty" yaml:",inline,omitempty"`

	// Name is a user-friendly name.  It may or may not be unique.
	Name string `json:"name"`

	// Tags are a collection of labels, in key-value form, about the object
	Tags map[string]string `json:"tags"`
}

// AddTagsFromStringSlice will parse any '=' delimited strings and set the Tags map. It overwrites on duplicate keys.
func (m Metadata) AddTagsFromStringSlice(v []string) Metadata {
	other := m
	if other.Tags == nil {
		other.Tags = map[string]string{}
	}
	for _, vv := range v {
		p := strings.Split(vv, "=")
		other.Tags[p[0]] = p[1]
	}
	return other
}

// URL is an alias of url
type URL url.URL

// NewURL creates a new url from string
func NewURL(s string) (*URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	v := URL(*u)
	return &v, nil
}

// Absolute returns true if the url is absolute (not relative)
func (u URL) Absolute() bool {
	return url.URL(u).Scheme != ""
}

// Value returns the aliased struct
func (u URL) Value() *url.URL {
	copy := url.URL(u)
	return &copy
}

// String returns the string representation of the URL
func (u URL) String() string {
	return u.Value().String()
}

// MarshalJSON returns the json representation
func (u URL) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, u.String())), nil
}

// UnmarshalJSON unmarshals the buffer to this struct
func (u *URL) UnmarshalJSON(buff []byte) error {
	str := strings.Trim(string(buff), " \"\\'\t\n")
	uu, err := url.Parse(str)
	if err != nil {
		return err
	}
	copy := URL(*uu)
	*u = copy
	return nil
}
