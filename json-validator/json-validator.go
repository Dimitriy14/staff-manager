package jsonvalidator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Dimitriy14/staff-manager/json-validator/schemas"

	"github.com/xeipuuv/gojsonschema"
)

var (
	schemaNames = map[string]string{
		schemas.UserRegistration:     schemas.UserRegistrationSchema,
		schemas.SignIn:               schemas.SignInSchema,
		schemas.UserUpdate:           schemas.UserUpdateSchema,
		schemas.AdminUserUpdate:      schemas.AdminUserUpdateSchema,
		schemas.TaskCreation:         schemas.TaskCreationSchema,
		schemas.TaskUpdate:           schemas.TaskUpdateSchema,
		schemas.TaskSearch:           schemas.TaskSearchSchema,
		schemas.VacationCreate:       schemas.VacationCreateSchema,
		schemas.VacationStatusUpdate: schemas.VacationStatusUpdateSchema,
	}

	schemasMap = map[string]*gojsonschema.Schema{}
)

// Validate validates data according to json schema
func Validate(schemaName string, data []byte) error {
	var err error
	schema, ok := schemasMap[schemaName]
	if !ok {
		schemaStr, ok := schemaNames[schemaName]
		if !ok {
			return fmt.Errorf("JSON schema validation: unknown schema %q", schemaName)
		}

		schema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaStr))
		if err != nil {
			return fmt.Errorf("JSON schema validation: parsing schema %q: %v", schemaName, err)
		}

		schemasMap[schemaName] = schema
	}

	documentLoader := gojsonschema.NewBytesLoader(data)

	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("JSON schema validation: validating %s: %v", data, err)
	}

	if result.Valid() {
		return nil
	}

	errs := make([]string, 0)
	for _, validationError := range result.Errors() {
		errs = append(errs, validationError.Field()+": "+validationError.Description())
	}

	return errors.New(strings.Join(errs, "; "))
}
