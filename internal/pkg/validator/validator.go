package validator

import "github.com/go-playground/validator/v10"

// V is global validator
var V = validator.New(validator.WithRequiredStructEnabled())
