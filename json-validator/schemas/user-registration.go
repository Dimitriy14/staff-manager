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
        "role": {
			"type": "string",
            "enum": ["admin", "user"]
		},
		"email": {
			"type": "string",
			"pattern": "^[a-zA-Z0-9]+@[a-z0-9]+.[a-z]{2,4}$"
		}
	},
	"required": ["firstName", "lastName", "position", "email", "role"],
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

var UserUpdate = "user-update"
var UserUpdateSchema = `
{
    "type": "object",
	"properties": {
		"image": {
			"type": "string",
			"minLength": 1
		},
        "mobilePhone": {
			"type": "string",
			"minLength": 8
		},
        "dateOfBirth": {
			"type": "string",
			"minLength": 8
		}
	},
    "additionalProperties": false
}
`

var AdminUserUpdate = "admin-user-update"
var AdminUserUpdateSchema = `
{
    "type": "object",
	"properties": {
        "mobilePhone": {
			"type": "string",
			"minLength": 8
		},
        "dateOfBirth": {
			"type": "string",
			"minLength": 8
		},
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
        "role": {
			"type": "string",
            "enum": ["admin", "user"]
		}
	},
    "required": ["firstName", "lastName", "position", "role"],
    "additionalProperties": false
}
`
