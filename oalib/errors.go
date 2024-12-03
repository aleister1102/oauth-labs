package oalib

// Error container for all kinds of OAuth errors.
// https://www.rfc-editor.org/rfc/rfc6749#section-4.2.2.1
type VerboseError struct {
	Err         string `json:"error"`
	Description string `json:"error_description"`
	URI         string `json:"error_uri"`
	State       string `json:"state"`
}

func (v VerboseError) Error() string {
	return v.Err
}
