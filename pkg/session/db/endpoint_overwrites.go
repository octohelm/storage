package db

import "net/url"

type EndpointOverrides struct {
	// Overwrite dbname when not empty
	NameOverwrite string `flag:",omitzero"`
	// Overwrite username when not empty
	UsernameOverwrite string `flag:",omitzero"`
	// Overwrite password when not empty
	PasswordOverwrite string `flag:",omitzero,secret"`
	// Overwrite extra when not empty
	ExtraOverwrite string `flag:",omitzero"`
}

func (d *EndpointOverrides) PatchEndpoint(endpoint *Endpoint) error {
	if name := d.NameOverwrite; name != "" {
		if endpoint.Scheme != "sqlite" {
			endpoint.Path = "/" + name
		}
	}

	if username := d.UsernameOverwrite; username != "" {
		endpoint.Username = username
	}

	if password := d.PasswordOverwrite; password != "" {
		endpoint.Password = password
	}

	if extra := d.ExtraOverwrite; extra != "" {
		q, err := url.ParseQuery(extra)
		if err != nil {
			return err
		}
		endpoint.Extra = q
	}

	return nil
}
