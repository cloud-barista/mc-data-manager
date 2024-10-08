# Data Manager Docker install 가이드


## 목차

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
            ca-certificates \
            curl \
            gnupg \
            lsb-release
        ```
    - Docker의 공식 GPG 키 추가
        ```shell
        sudo mkdir -p /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        ```

    - Docker apt repository 설정
        ```shell
        echo \
            "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
            $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        ```

    - Docker Engine 설치
        ```bash
        sudo apt-get update
        sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin
        ```
    - Docker 설치 확인
        ```bash
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

    - Docker run
        -실행 
            ```shell
            docker run -d \
                -p 3300:3300 \
                -v data:/app/data \
                --name mc-data-manager \
                cloudbaristaorg/mc-data-manager
            ```
        - 인증 프로필 카피
            ```shell
            docker  cp  profile.json  mc-data-manager:/app/data/var/run/data-manager/profile/profile.json
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
            ```
        - 실행
            ```shell
            docker compose -f <filename>.yaml up -d
            ```
        - 인증 프로필 카피
            ```shell
            docker compose cp  profile.json  mc-data-manager:/app/data/var/run/data-manager/profile/profile.json
            ```

