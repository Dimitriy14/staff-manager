package schemas

var TaskCreation = "task-creation"
var TaskCreationSchema = `
{
    "type": "object",
	"properties": {
		"title": {
			"type": "string",
			"minLength": 1
		},
        "description": {
            "type": "string",
			"minLength": 1
        },
		"assignedID": {
			"type": "string",
            "pattern": "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
        }
	},
	"required": ["title", "description"],
    "additionalProperties": false
}
`

var TaskUpdate = "task-update"
var TaskUpdateSchema = `
{
    "type": "object",
	"properties": {
		"title": {
			"type": "string",
			"minLength": 1
		},
        "description": {
            "type": "string",
			"minLength": 1
        },
        "status": {
            "type": "string",
            "enum": ["Ready", "InProgress", "Done", "Blocked"]
        },
		"assignedID": {
			"type": "string",
            "pattern": "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
        }
	},
	"required": ["title", "description"],
    "additionalProperties": false
}
`
