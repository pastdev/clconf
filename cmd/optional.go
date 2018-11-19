package cmd

type optionalString struct {
	set   bool
	value string
}

// NewOptionalString returns a new optional string
func NewOptionalString(value string, set bool) *optionalString {
	return &optionalString{set: set, value: value}
}

func (s *optionalString) Set(value string) error {
	s.value = value
	s.set = true
	return nil
}

func (s *optionalString) String() string {
	return s.value
}

func (s *optionalString) Type() string {
	return "string"
}
