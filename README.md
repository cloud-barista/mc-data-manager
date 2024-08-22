# Cloud-Migrator Data Manager

Data Manager 데이터 마이그레이션 기술의 검증을 위한 환경을 구축하고, 데이터 마이그레이션에 필요한 테스트 데이터를 생성하는 도구이다.
이를 위해 아래와 같은 주요 기능을 제공한다.
1. 데이터 저장소(스토리지 또는 데이터베이스)를 목표 및 소스 컴퓨팅 환경에 생성한다. 
2. 생성된 소스 데이터 저장소에 테스트 데이터를 생성 및 저장한다.
3. 소스에서 목표 컴퓨팅 환경으로 데이터 복제/마이그레이션을 수행하며, 이때 데이터 전/후처리 작업을 수행한다.


## Environments:
* OS: Ubuntu 22.04 LTS, Windows 10 Pro
* Go: 1.21.3


## Installation and Testing Guide

해당 가이드는 Ubuntu 22.04 대상으로 설치 및 명령어 사용방법을 작성한 가이드입니다.

* [Data Manager 기능명세서](docs/Data-manager-Function-Specification.md)
* [Data Manager 사용가이드](docs/Data-manager-Usage-Guide.md)