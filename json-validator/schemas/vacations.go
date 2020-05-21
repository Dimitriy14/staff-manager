package schemas

var VacationCreate = "VacationCreate"
var VacationCreateSchema = `
{
    "type": "object",
	"properties": {
		"startDate": {
			"type": "string",
			"format": "date"
		},
        "endDate": {
            "type": "string",
			"format": "date"
        }
	},
	"required": ["startDate", "endDate"],
    "additionalProperties": false
}
`

var VacationStatusUpdate = "VacationStatusUpdate"
var VacationStatusUpdateSchema = `
{
    "type": "object",
	"properties": {
        "status": {
			"type": "string",
            "enum": ["Approved", "Rejected"]
		}
	},
	"required": ["status"],
    "additionalProperties": false
}
`
