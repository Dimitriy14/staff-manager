# staff-manager 
### BUILD  
To build a binary run `go build` all dependencies will be pulled automaticaly.  
### RUN  
Then run app with command `./staff-manager -config path/to/config.json`  
### CONFIGURATION  
Staff manager requires setted up [AWS Credentials](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html)  
Postgres credentials should be stored in AWS Secret Manager:  
`{
  "Host": "localhost",
  "Port": "5431",
  "User": "app",
  "Password": "1337",
  "DataBaseName": "staff_manager"
}`  
Other configuration can be changed in [config.json](../config.json) 
