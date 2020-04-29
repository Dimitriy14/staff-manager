package schemas

var UserRegistration = "user-registration"
var UserRegistrationSchema = `
{
    "type": "object",
	"properties": {
		"firstName": {
			"type": "string",
			"minLength": 1
		},
        "lastName": {
            "type": "string",
			"minLength": 1
        },
		"position": {
			"type": "string",
			"minLength": 1
		},
		"email": {
			"type": "string",
			"pattern": "^[a-zA-Z0-9]+@[a-z0-9]+.[a-z]{2,4}$"
		}
	},
	"required": ["firstName", "lastName", "position", "email"],
    "additionalProperties": false
}
`

var SignIn = "sign-in"
var SignInSchema = `
{
    "type": "object",
	"properties": {
		"password": {
			"type": "string",
			"minLength": 8
		},
		"email": {
			"type": "string",
			"pattern": "^[a-zA-Z0-9]+@[a-z0-9]+.[a-z]{2,4}$"
		}
	},
	"required": ["password", "email"],
    "additionalProperties": false
}
`
