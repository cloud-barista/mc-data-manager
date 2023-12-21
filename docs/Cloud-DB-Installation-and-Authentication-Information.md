# Cloud DB 생성 및 인증정보

## AWS RDS MySQL 생성 및 인증정보
1. AWS 콘솔에 로그인합니다.
2. RDS 서비스 선택 -> 대시보드 -> "데이터베이스 생성" 메뉴를 선택합니다.

<p align="center"><img src="/docs/image/pre-check/rds0.png" ></p>

3. 데이터베이스 생성방식 선택은 "표준 생성"을 선택하고, 엔진 옵션은 "MySQL"을 선택합니다.

<p align="center"><img src="/docs/image/pre-check/rds1.png" ></p>

4. 템플릿은 비용에 맞게 적절히 선택이 필요합니다. (현 가이드에선 "프리 티어" 선택)
DB인스턴스 식별자는 RDS 접속을 위해 자동으로 생성할 인스턴스 이름을 입력합니다.
자격증명은 DB 접속을 위한 계정/패스워드를 입력합니다.

<p align="center"><img src="/docs/image/pre-check/rds2.png" ></p>

5. DB인스턴스 생성을 위한 스펙을 선택합니다.

<p align="center"><img src="/docs/image/pre-check/rds3.png" ></p>

6. 네트워크 설정 후 퍼블릭 엑세스는 "예"로 선택합니다.

<p align="center"><img src="/docs/image/pre-check/rds4.png" ></p>

7. 세부사항은 필요시 추가한 후 "데이터베이스 생성" 버튼을 클릭합니다.

<p align="center"><img src="/docs/image/pre-check/rds5.png" ></p>

8. 데이터베이스 생성화면입니다.

<p align="center"><img src="/docs/image/pre-check/rds6.png" ></p>

9.  정상적으로 생성 후 엔드포인트 및 포트정보를 확인할 수 있습니다.
RDS(MySQL) 연결을 위해 보안그룹을 클릭합니다.

<p align="center"><img src="/docs/image/pre-check/rds7.png" ></p>

10.	인바운드 규칙을 수정해서 MySQL 접속하는 Client(PC)의 IP와 포트 규칙을 추가합니다.
    ```
    프로토콜 : TCP
    포트 범위 : 3306
    소스 : xxx.xxx.xxx.xxx/32
    ```

<p align="center"><img src="/docs/image/pre-check/rds8.png" ></p>

11.	Client(PC)에서 MySQL를 통해 정상적으로 접속되는지 확인합니다.

<p align="center"><img src="/docs/image/pre-check/rds9.png" ></p>

## GCP SQL MySQL 생성 및 인증정보
1.	GCP 콘솔에 로그인합니다.
2.	메뉴->SQL으로 들어가서 "인스턴스 만들기"를 클릭합니다.

<p align="center"><img src="/docs/image/pre-check/sql0.png" ></p>

3.	MYSQL 인스턴스 만들기에서 인스턴스 이름과 root 계정의 비밀번호를 입력합니다.

<p align="center"><img src="/docs/image/pre-check/sql1.png" ></p>

4.	Cloud SQL 버전은 특성에 따라 선택합니다.

<p align="center"><img src="/docs/image/pre-check/sql2.png" ></p>

5.	리전은 "asia-northeast3 (서울)"을 선택합니다.

<p align="center"><img src="/docs/image/pre-check/sql3.png" ></p>

6.	접근을 위해 "공개 IP" 체크와 "네트워크 추가"를 클릭해서 접속할 Client(PC)의 IP를 입력합니다.

<p align="center"><img src="/docs/image/pre-check/sql4.png" ></p>

7.	설정이 마치면 Cloud SQL서비스를 생성합니다.

<p align="center"><img src="/docs/image/pre-check/sql5.png" ></p>

8.	생성이 완료되면 왼쪽 메뉴의 연결 클릭 시 접속할 수 있는 공개IP를 확인할 수 있습니다.

<p align="center"><img src="/docs/image/pre-check/sql6.png" ></p>

9.	공개IP를 이용하여 SQL에 접속합니다.

<p align="center"><img src="/docs/image/pre-check/sql7.png" ></p>

## NCP Cloud DB for MySQL 생성 및 인증정보
1.	NCP 콘솔에 로그인합니다.
2.	메뉴->Cloud DB for MySQL에 선택한 후 “DB Server 생성”을 클릭합니다.

<p align="center"><img src="/docs/image/pre-check/mysql0.png" ></p>

3.	DB Server 생성에 필요한 정보를 입력합니다.

<p align="center"><img src="/docs/image/pre-check/mysql1.png" ></p>

4.	접속ID, 암호, 포트, 접속 허용할 Host IP 와 기본 DB 명을 입력합니다.

<p align="center"><img src="/docs/image/pre-check/mysql2.png" ></p>

5.	생성 된 DB Server를 클릭한 후 Public 도메인 관리를 선택합니다.

<p align="center"><img src="/docs/image/pre-check/mysql3.png" ></p>

6.	Public 도메인 신청에서 “예”를 클릭하여 도메인은 신청합니다.

<p align="center"><img src="/docs/image/pre-check/mysql4.png" ></p>

7.	접근 보안을 위해 Client(PC)의 IP를 확인한 후 ACG 설정으로 들어갑니다.

<p align="center"><img src="/docs/image/pre-check/mysql5.png" ></p>

8.	DB서버의 ACG에 접근 소스와 허용포트 ”3306” 값을 추가한 후 적용을 합니다.

<p align="center"><img src="/docs/image/pre-check/mysql6.png" ></p>

9.	DB에 접근이 되는지 확인합니다.

<p align="center"><img src="/docs/image/pre-check/mysql7.png" ></p>

## NCP Cloud DB for MongoDB 생성 및 인증정보
1.	NCP 콘솔에 로그인합니다.
2.	메뉴->Cloud DB for MySQL에 선택한 후 “DB Server 생성”을 클릭합니다.
    ![mongodb0](/docs/image/pre-check/mongodb0.png)
3.	DB Server 생성에 필요한 정보를 입력합니다.
    ![mongodb1](/docs/image/pre-check/mongodb1.png)
    ![mongodb2](/docs/image/pre-check/mongodb2.png)
4.	접속ID,암호를 입력합니다.
    ![mongodb3](/docs/image/pre-check/mongodb3.png)
5.	생성 된 DB Server를 클릭한 후 Public 도메인 관리를 선택합니다.
    ![mongodb4](/docs/image/pre-check/mongodb4.png)
6.	접근 보안을 위해 Client(PC)의 IP를 확인한 후 ACG 설정으로 들어갑니다.
    ![mongodb5](/docs/image/pre-check/mongodb5.png)
7.	DB서버의 ACG에 접근 소스와 허용포트 ”17017” 값을 추가한 후 적용을 합니다.
    ![mongodb6](/docs/image/pre-check/mongodb6.png)
8.	DB에 접근이 되는지 확인합니다.
    ![mongodb7](/docs/image/pre-check/mongodb7.png)
    ![mongodb8](/docs/image/pre-check/mongodb8.png)
