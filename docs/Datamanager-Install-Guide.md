# Data Manager Docker install 가이드


## 목차

- [Data Manager Docker install 가이드](#data-manager-docker-install-가이드)
  - [목차](#목차)
  - [사전 준비](#사전-준비)
  - [Linux에서 Docker 설치 및 실행](#linux에서-docker-설치-및-실행)


## 사전 준비
- linux (ubuntu 22.04) 대상
- [Docker 공식 ](https://docs.docker.com/engine/install/ubuntu) 설치


## Linux에서 Docker 설치 및 실행

1. apt-repo를 이용한 docker 설치

    - apt 패키지 업데이트
        ```bash
        sudo apt-get update
        ```
    - 필수 패키지 설치
        ```shell
        sudo apt-get install \
        ca-certificates curl
        ```
    - Docker의 공식 GPG 키 추가
        ```shell
        sudo install -m 0755 -d /etc/apt/keyrings

        sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
        
        sudo chmod a+r /etc/apt/keyrings/docker.asc
        ```

    - Docker apt repository 설정
        ```shell
        echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
        sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        ```

    - Docker Engine 설치
        ```shell
        sudo apt-get update
        sudo apt-get install \ docker-ce docker-ce-cli \ containerd.io \ docker-compose-plugin
        ```
    - Docker 설치 확인
        ```shell
        sudo docker --version
        ```

    - docker-compose command alias
        ```shell
        echo "alias docker-compose='docker compose'" >> ~/.$bashrc 
        ```

    - Add Group
        ```shell
        sudo usermod -aG docker $USER && exec sudo su - $USER
        ```

2. Linux에서 Docker 커맨드 실행
    - Docker 환경변수 설정

        - .env 환경변수 설정
            ```shell
            // 인증 관련 데이터베이스 설정
            MC_DATA_MANAGER_DATABASE_HOST=
            MC_DATA_MANAGER_DATABASE_PORT=
            MC_DATA_MANAGER_DATABASE_USER=
            MC_DATA_MANAGER_DATABASE_PASSWORD=
            MC_DATA_MANAGER_DATABASE_NAME=

            // 암/복호화 관련 키 설정
            ENCODING_SECRET_KEY=

            // mc-data-manager 서버 설정
            MC_DATA_MANAGER_PORT=
            MC_DATA_MANAGER_ALLOW_IP_RANGE= //허용 CIDR ex. 0.0.0.0/0

            // tumblebug api URL 설정
            TUMBLEBUG_URL=
            ```

    - Docker run

        - 실행
            ```shell
            docker run -d \
                -p 3300:3300 \
                -v data:/app/data \
                --name mc-data-manager \
                cloudbaristaorg/mc-data-manager
            ```

    - Docker compose
        - docker-compose yaml 파일 생성  
            ```yaml
            services:
                mc-data-manager:
                    container_name: mc-data-manager
                    image: cloudbaristaorg/mc-data-manager
                    ports:
                        - "3300:3300"
                    volumes:
                        - ./data:/app/data/
                        - ./scripts:/app/scripts/
                    ...
            ```
        - 실행
            ```shell
            docker compose -f <filename>.yaml up -d
            ```

