package profiles

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/spiegela/gorackhd/models"
)

// ProfilesPutLibByNameReader is a Reader for the ProfilesPutLibByName structure.
type ProfilesPutLibByNameReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ProfilesPutLibByNameReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 201:
		result := NewProfilesPutLibByNameCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 500:
		result := NewProfilesPutLibByNameInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewProfilesPutLibByNameDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewProfilesPutLibByNameCreated creates a ProfilesPutLibByNameCreated with default headers values
func NewProfilesPutLibByNameCreated() *ProfilesPutLibByNameCreated {
	return &ProfilesPutLibByNameCreated{}
}

/*ProfilesPutLibByNameCreated handles this case with default header values.

Successfully created or modified the specified profile
*/
type ProfilesPutLibByNameCreated struct {
	Payload ProfilesPutLibByNameCreatedBody
}

func (o *ProfilesPutLibByNameCreated) Error() string {
	return fmt.Sprintf("[PUT /profiles/library/{name}][%d] profilesPutLibByNameCreated  %+v", 201, o.Payload)
}

func (o *ProfilesPutLibByNameCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewProfilesPutLibByNameInternalServerError creates a ProfilesPutLibByNameInternalServerError with default headers values
func NewProfilesPutLibByNameInternalServerError() *ProfilesPutLibByNameInternalServerError {
	return &ProfilesPutLibByNameInternalServerError{}
}

/*ProfilesPutLibByNameInternalServerError handles this case with default header values.

Profile creation failed
*/
type ProfilesPutLibByNameInternalServerError struct {
	Payload *models.Error
}

func (o *ProfilesPutLibByNameInternalServerError) Error() string {
	return fmt.Sprintf("[PUT /profiles/library/{name}][%d] profilesPutLibByNameInternalServerError  %+v", 500, o.Payload)
}

func (o *ProfilesPutLibByNameInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewProfilesPutLibByNameDefault creates a ProfilesPutLibByNameDefault with default headers values
func NewProfilesPutLibByNameDefault(code int) *ProfilesPutLibByNameDefault {
	return &ProfilesPutLibByNameDefault{
		_statusCode: code,
	}
}

/*ProfilesPutLibByNameDefault handles this case with default header values.

Unexpected error
*/
type ProfilesPutLibByNameDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the profiles put lib by name default response
func (o *ProfilesPutLibByNameDefault) Code() int {
	return o._statusCode
}

func (o *ProfilesPutLibByNameDefault) Error() string {
	return fmt.Sprintf("[PUT /profiles/library/{name}][%d] profilesPutLibByName default  %+v", o._statusCode, o.Payload)
}

func (o *ProfilesPutLibByNameDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*ProfilesPutLibByNameCreatedBody profiles put lib by name created body
swagger:model ProfilesPutLibByNameCreatedBody
*/
type ProfilesPutLibByNameCreatedBody interface{}
