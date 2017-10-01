package tasks

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/spiegela/gorackhd/models"
)

// GetTasksByIDReader is a Reader for the GetTasksByID structure.
type GetTasksByIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetTasksByIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetTasksByIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 404:
		result := NewGetTasksByIDNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewGetTasksByIDDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetTasksByIDOK creates a GetTasksByIDOK with default headers values
func NewGetTasksByIDOK() *GetTasksByIDOK {
	return &GetTasksByIDOK{}
}

/*GetTasksByIDOK handles this case with default header values.

Successfully retrieved the specified task
*/
type GetTasksByIDOK struct {
	Payload GetTasksByIDOKBody
}

func (o *GetTasksByIDOK) Error() string {
	return fmt.Sprintf("[GET /tasks/{identifier}][%d] getTasksByIdOK  %+v", 200, o.Payload)
}

func (o *GetTasksByIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetTasksByIDNotFound creates a GetTasksByIDNotFound with default headers values
func NewGetTasksByIDNotFound() *GetTasksByIDNotFound {
	return &GetTasksByIDNotFound{}
}

/*GetTasksByIDNotFound handles this case with default header values.

The specified task was not found
*/
type GetTasksByIDNotFound struct {
	Payload *models.Error
}

func (o *GetTasksByIDNotFound) Error() string {
	return fmt.Sprintf("[GET /tasks/{identifier}][%d] getTasksByIdNotFound  %+v", 404, o.Payload)
}

func (o *GetTasksByIDNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetTasksByIDDefault creates a GetTasksByIDDefault with default headers values
func NewGetTasksByIDDefault(code int) *GetTasksByIDDefault {
	return &GetTasksByIDDefault{
		_statusCode: code,
	}
}

/*GetTasksByIDDefault handles this case with default header values.

Unexpected error
*/
type GetTasksByIDDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get tasks by Id default response
func (o *GetTasksByIDDefault) Code() int {
	return o._statusCode
}

func (o *GetTasksByIDDefault) Error() string {
	return fmt.Sprintf("[GET /tasks/{identifier}][%d] getTasksById default  %+v", o._statusCode, o.Payload)
}

func (o *GetTasksByIDDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*GetTasksByIDOKBody get tasks by ID o k body
swagger:model GetTasksByIDOKBody
*/
type GetTasksByIDOKBody interface{}
