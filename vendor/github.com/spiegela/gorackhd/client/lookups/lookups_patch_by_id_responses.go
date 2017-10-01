package lookups

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/spiegela/gorackhd/models"
)

// LookupsPatchByIDReader is a Reader for the LookupsPatchByID structure.
type LookupsPatchByIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *LookupsPatchByIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewLookupsPatchByIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 404:
		result := NewLookupsPatchByIDNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewLookupsPatchByIDDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewLookupsPatchByIDOK creates a LookupsPatchByIDOK with default headers values
func NewLookupsPatchByIDOK() *LookupsPatchByIDOK {
	return &LookupsPatchByIDOK{}
}

/*LookupsPatchByIDOK handles this case with default header values.

Successfully modified the lookup
*/
type LookupsPatchByIDOK struct {
	Payload *models.Lookups20LookupBase
}

func (o *LookupsPatchByIDOK) Error() string {
	return fmt.Sprintf("[PATCH /lookups/{id}][%d] lookupsPatchByIdOK  %+v", 200, o.Payload)
}

func (o *LookupsPatchByIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Lookups20LookupBase)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewLookupsPatchByIDNotFound creates a LookupsPatchByIDNotFound with default headers values
func NewLookupsPatchByIDNotFound() *LookupsPatchByIDNotFound {
	return &LookupsPatchByIDNotFound{}
}

/*LookupsPatchByIDNotFound handles this case with default header values.

The specified lookup was not found
*/
type LookupsPatchByIDNotFound struct {
	Payload *models.Error
}

func (o *LookupsPatchByIDNotFound) Error() string {
	return fmt.Sprintf("[PATCH /lookups/{id}][%d] lookupsPatchByIdNotFound  %+v", 404, o.Payload)
}

func (o *LookupsPatchByIDNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewLookupsPatchByIDDefault creates a LookupsPatchByIDDefault with default headers values
func NewLookupsPatchByIDDefault(code int) *LookupsPatchByIDDefault {
	return &LookupsPatchByIDDefault{
		_statusCode: code,
	}
}

/*LookupsPatchByIDDefault handles this case with default header values.

Unexpected error
*/
type LookupsPatchByIDDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the lookups patch by Id default response
func (o *LookupsPatchByIDDefault) Code() int {
	return o._statusCode
}

func (o *LookupsPatchByIDDefault) Error() string {
	return fmt.Sprintf("[PATCH /lookups/{id}][%d] lookupsPatchById default  %+v", o._statusCode, o.Payload)
}

func (o *LookupsPatchByIDDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
