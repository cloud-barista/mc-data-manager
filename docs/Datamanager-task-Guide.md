



AWS-NRDB-Backup 
```go
{
    "cron": "* * * * *",
    "tz": "Asia/Seoul",

    "operationId": "operationID-XXX",
    "tasks": [
        {
            "meta": {
                "serviceType": "nrdbms",
                "taskType": "backup",
            },
            "targetPoint": {
                "profileName": "admin",
                "provider": "aws",
                "region": "ap-northeast-2"
            },
            "Directory":"./tmp/schedule/dummy/NRDB/aws"
        }
    ]
}
```


GCP-NRDB-BACKUP
```go
{
    "cron": "* * * * *",
    "tz": "Asia/Seoul",
    "operationId": "operationID-Nrdb-backup-gcp",
    "tasks": [
        {
            "meta": {
                "serviceType": "nrdbms",
                "taskType": "backup"
            },
            "targetPoint": {
                "profileName": "admin",
                "provider": "gcp",
                "region": "asia-northeast2"
            },
            "Directory": "./tmp/schedule/dummy/NRDB/gcp"
        }
    ]
}
```