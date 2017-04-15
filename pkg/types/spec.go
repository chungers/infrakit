package types

// URL is a url string
type URL string

// Spec is the specification of the resource / object
type Spec struct {

	// Class is the kind/type of the resource -- e.g. instance-aws/ec2-instance
	Class string `json:"class"`

	// SpiVersion is the name of the interface and version - instance/v0.1.0
	SpiVersion string `json:"spiVersion"`

	// Metadata is metadata / useful information about object
	Metadata Metadata `json:"metadata"`

	// Template is a template of the resource's properties
	Template URL `json:"template,omitempty" yaml:",omitempty"`

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
	if s.Class == "" {
		return errMissingAttribute("class")
	}
	if s.SpiVersion == "" {
		return errMissingAttribute("spiVersion")
	}
	if s.Metadata.Name == "" {
		return errMissingAttribute("metadata.name")
	}
	return nil
}

// Dependency models the reference and usage of another spec, by spec's Class and Name, and a way
// to extract its properties, and how it's referenced via the alias in the Properties section of the dependent Spec.
type Dependency struct {

	// Class is the Class of the spec this spec depends on
	Class string

	// Name is the Name of the spec this spec dependes on
	Name string

	// Bind is an associative array of pointer to the fields in the object to a variable name that will be referenced
	// in the properties or template of the owning spec.
	Bind map[string]*Pointer
}

// Identity uniquely identifies an instance
type Identity struct {

	// UID is a unique identifier for the object instance.
	UID string
}

// Metadata captures label and descriptive information about the object
type Metadata struct {

	// Identity is an optional component that exists only in the case of a real object instance.
	*Identity `json:",omitempty" yaml:",omitempty"`

	// Name is a user-friendly name.  It may or may not be unique.
	Name string

	// Tags are a collection of labels, in key-value form, about the object
	Tags map[string]string
}
