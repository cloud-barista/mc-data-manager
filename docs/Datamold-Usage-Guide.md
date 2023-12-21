# Datamold 사용 가이드

## 사전 준비 사항
### 서비스 신청 및 권한 부여
* GCP, NCP같은 경우 서비스 이용 시 이용 신청을 해야한다.
* 로그인 후 사용할 서비스 페이지 접속 후 이용 신청을 하면 된다.
    * 예시

        **GCP**
        ![gcp_pre_auth](/docs/image/pre-check/gcp_pre_auth.png)
        **NCP**
        
<p align="center"><img src="/docs/image/pre-check/ncp_pre_auth.png" ></p>

* GCP와 NCP는 사전에 인증정보에 권한을 부여해야합니다.
  
    사용하고자 하는 서비스계정에 Storage Admin 권한 추가
    * GCP : https://cloud.google.com/storage/docs/access-control/iam-roles?hl=ko
    
    서브 계정 사용 시 Object Storage 권한 추가
    * NCP : https://guide.ncloud-docs.com/docs/storage-objectstorage-subaccount

### CSP 인증정보
1. AWS 인증정보
    * [AWS S3, DynamoDB 인증정보](https://docs.aws.amazon.com/ko_kr/IAM/latest/UserGuide/id_credentials_access-keys.html#Using_CreateAccessKey)
2. GCP 인증정보
    * [GCP Cloud Storage, FirestoreDB 인증정보](https://developers.google.com/workspace/guides/create-credentials?hl=ko)
3. NCP 인증정보
    * [NCP Object Storage 인증정보](https://medium.com/naver-cloud-platform/%EC%9D%B4%EB%A0%87%EA%B2%8C-%EC%82%AC%EC%9A%A9%ED%95%98%EC%84%B8%EC%9A%94-%EB%84%A4%EC%9D%B4%EB%B2%84-%ED%81%B4%EB%9D%BC%EC%9A%B0%EB%93%9C-%ED%94%8C%EB%9E%AB%ED%8F%BC-%EC%9C%A0%EC%A0%80-api-%ED%99%9C%EC%9A%A9-%EB%B0%A9%EB%B2%95-1%ED%8E%B8-494f7d8dbcc3)

### AWS, GCP, NCP Cloud DB 설치 및 인증정보
* [DB 설치 및 인증정보](/docs/Cloud-DB-Installation-and-Authentication-Information.md)

## 1. 정형데이터 생성 및 마이그레이션
### 온프레미스(리눅스서버)에서 AWS S3
1. data-mold server 접속 후 좌측 메뉴에서 데이터 생성 -> Object Storage -> AWS S3 순으로 클릭
    ![main](/docs/image/web/main.png)
2. 사용자의 AWS 인증정보와 사용하고자 하는 리전 선택, 버킷 명을 입력한 다음 생성 할 데이터를 선택 및 용량 입력 후 생성 버튼 클릭
    ![main](/docs/image/web/creates3.png)
3. 성공 및 실패는 아래 로그에서 확인이 가능합니다.
    ![main](/docs/image/web/s3sql.png)

## 2. 비정형데이터 생성 및 마이그레이션
### 온프레미스(리눅스서버)에서 GCP Cloud Storage
1. data-mold server 접속 후 좌측 메뉴에서 데이터 생성 -> Object Storage -> Google Cloud Storage 순으로 클릭
    ![main](/docs/image/web/main.png)
2. 사용자의 GCP 인증정보와 사용하고자 하는 리전 선택, 버킷 명을 입력한 다음 생성 할 데이터를 선택 및 용량 입력 후 생성 버튼 클릭
    ![main](/docs/image/web/creategcp.png)
3. 성공 및 실패는 아래 로그에서 확인이 가능합니다.
    ![main](/docs/image/web/creategcpresult.png)

## 3. 반정형데이터 생성 및 마이그레이션
### 온프레미스(리눅스서버)에서 NCP Object Storage
1. data-mold server 접속 후 좌측 메뉴에서 데이터 생성 -> Object Storage -> Google Cloud Storage 순으로 클릭
    ![main](/docs/image/web/main.png)
2. 사용자의 NCP 인증정보와 사용하고자 하는 리전 선택, 버킷 명을 입력한 다음 생성 할 데이터를 선택 및 용량 입력 후 생성 버튼 클릭
    ![main](/docs/image/web/createncp.png)
3. 성공 및 실패는 아래 로그에서 확인이 가능합니다.
    ![main](/docs/image/web/ncpjson.png)

## 4. 클라우드 관계형데이터베이스 생성 및 마이그레이션
### AWS RDS(MySQL)에서 GCP Cloud SQL(MySQL) 환경 시연
1. data-mold server 접속 후 좌측 메뉴에서 Migration -> SQL Database -> MySQl 순으로 클릭
    ![main](/docs/image/web/main.png)
2. 사용자의 AWS RDS 접속정보를 소스 MySQL, 사용자의 GCP SQL 접속정보를 목표 MySQL에 입력 후 생성 버튼 클릭
    ![main](/docs/image/web/migmysql.png)
3. 성공 및 실패는 아래 로그에서 확인이 가능합니다.
    ![main](/docs/image/web/rdstosql.png)

## 5. 클라우드 비관계형데이터베이스 생성 및 마이그레이션
### AWS DynamoDB에서 NCP Cloud DB for MongoDB
1. data-mold server 접속 후 좌측 메뉴에서 Migration -> NoSQL -> AWS DynamoDB to -> MongoDB 순으로 클릭
    ![main](/docs/image/web/main.png)
2. 사용자의 AWS RDS 접속정보를 소스 MySQL, 사용자의 GCP SQL 접속정보를 목표 MySQL에 입력 후 생성 버튼 클릭
    ![main](/docs/image/web/migmysql.png)
3. 성공 및 실패는 아래 로그에서 확인이 가능합니다.
    ![main](/docs/image/web/rdstosql.png)
