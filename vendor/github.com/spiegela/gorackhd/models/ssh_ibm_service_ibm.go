package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// SSHIbmServiceIbm SSH settings
// swagger:model ssh-ibm-service_Ibm
type SSHIbmServiceIbm struct {

	// config
	// Required: true
	Config *SSHIbmServiceIbmConfig `json:"config"`

	// service
	// Required: true
	Service *string `json:"service"`
}

// Validate validates this ssh ibm service ibm
func (m *SSHIbmServiceIbm) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateConfig(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateService(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SSHIbmServiceIbm) validateConfig(formats strfmt.Registry) error {

	if err := validate.Required("config", "body", m.Config); err != nil {
		return err
	}

	if m.Config != nil {

		if err := m.Config.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("config")
			}
			return err
		}
	}

	return nil
}

func (m *SSHIbmServiceIbm) validateService(formats strfmt.Registry) error {

	if err := validate.Required("service", "body", m.Service); err != nil {
		return err
	}

	return nil
}

// SSHIbmServiceIbmConfig SSH ibm service ibm config
// swagger:model SSHIbmServiceIbmConfig
type SSHIbmServiceIbmConfig struct {

	// IP address
	Host string `json:"host,omitempty"`

	// Password
	Password string `json:"password,omitempty"`

	// Username
	User string `json:"user,omitempty"`
}

// Validate validates this SSH ibm service ibm config
func (m *SSHIbmServiceIbmConfig) Validate(formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
