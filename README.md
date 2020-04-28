# staff-manager 
### BUILD  
To build a binary run `go build` all dependencies will be pulled automaticaly.  
### RUN  
Then run app with command `./staff-manager -config path/to/config.json`  
To recognize if application successfully running run  
`curl -X GET http://localhost:1234/staff/health`  
Response:  
```json
{
    "currentTime": "2020-04-19T15:08:56.758226+03:00",
    "startTime": "2020-04-19T15:08:38.729453+03:00",
    "networkInterfaces": [
        "127.0.0.1:1234"
    ],
    "connections": [
        {
            "serviceName": "Postgres",
            "activeNodes": [
                "localhost:5431"
            ]
        },
        {
            "serviceName": "ElasticSearch",
            "activeNodes": [
                "http://127.0.0.1:9200"
            ]
        }
    ]
}
```
### CONFIGURATION  
Staff manager requires setted up [AWS Credentials](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html)  
Secret configuration should be stored in AWS Secret Manager:  
```json
{
  "Postgres": {
    "Host": "localhost",
    "Port": "5431",
    "User": "app",
    "Password": "1337",
    "DataBaseName": "staff_manager"
  },
  "Cognito": {
    "ClientID": "",
    "UserPoolID": ""
  }
}
```   
Other configuration can be changed in [config.json](./config.json) 
