# CI/CD 설정 가이드

## 개요

이 프로젝트는 GitHub Actions와 Self-hosted Runner를 사용하여 자동 빌드 및 배포를 수행합니다.

## 아키텍처

```
GitHub Repository (push) → GitHub Actions → Self-hosted Runner → Docker Build → Deploy
```

- **빌드 서버 = 배포 서버**: 동일한 서버에서 빌드와 배포 수행
- **서비스별 독립 컨테이너**: 각 서비스는 고유한 포트로 실행
- **Docker Compose 기반**: 서비스 오케스트레이션

## 서비스 포트 구조

각 서비스는 독립된 포트에서 실행되며, **MySQL은 3306 포트로 공유**됩니다:

| 서비스 | API 포트 | DB 이름 | 설명 |
|--------|----------|---------|------|
| 인증 서비스 (joker-backend) | 6000 | joker_backend | 기본 백엔드 |
| 게임 서버 | 6001 | game_server | 추가 예정 |
| 결제 서비스 | 6002 | payment_service | 추가 예정 |

**주요 특징:**
- **공유 MySQL**: 모든 서비스가 3306 포트의 MySQL 인스턴스 공유
- **독립 데이터베이스**: 각 서비스는 고유한 데이터베이스 이름 사용
- **독립 API 포트**: 각 서비스는 고유한 포트에서 실행

## Self-hosted Runner 설정

### 1. Runner 설치

GitHub Repository → Settings → Actions → Runners → New self-hosted runner

```bash
# Runner 다운로드 및 설치
mkdir actions-runner && cd actions-runner
curl -o actions-runner-linux-x64-2.311.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz
tar xzf ./actions-runner-linux-x64-2.311.0.tar.gz

# Runner 구성
./config.sh --url https://github.com/JokerTrickster/joker_backend --token YOUR_TOKEN

# Runner 서비스로 실행
sudo ./svc.sh install
sudo ./svc.sh start
```

### 2. 서버 환경 요구사항

- Docker & Docker Compose 설치
- rsync 설치
- curl 설치
- 충분한 디스크 공간 (최소 20GB)

```bash
# 필수 패키지 설치
sudo apt update
sudo apt install -y docker.io docker-compose rsync curl

# Docker 권한 설정
sudo usermod -aG docker $USER
newgrp docker
```

## GitHub Secrets 설정

Repository → Settings → Secrets and variables → Actions → New repository secret

필수 Secrets:

| Secret 이름 | 설명 | 예시 |
|-------------|------|------|
| `DB_USER` | 데이터베이스 사용자명 | joker_user |
| `DB_PASSWORD` | 데이터베이스 비밀번호 | secure_password_123 |
| `DB_NAME` | 데이터베이스 이름 | joker_backend |
| `MYSQL_ROOT_PASSWORD` | MySQL root 비밀번호 | root_password_456 |

## 배포 워크플로우

### 자동 배포 트리거

- `main` 브랜치에 push
- `develop` 브랜치에 push
- 수동 실행 (workflow_dispatch)

### 배포 프로세스

1. **체크아웃**: 코드 가져오기
2. **디스크 정리**: 빌드 전 Docker 리소스 정리 (디스크 공간 확보)
3. **환경 설정**: 환경 변수 및 디렉토리 생성 (`$HOME/services/[service-name]`)
4. **파일 복사**: 프로젝트 파일을 배포 디렉토리로 복사
5. **환경 파일 생성**: `.env` 파일 생성
6. **컨테이너 중지**: 기존 컨테이너 중지
7. **빌드 및 시작**: 새 컨테이너 빌드 및 시작
8. **헬스체크**: 서비스 상태 확인
9. **검증**: 배포 성공 확인
10. **정리**: 오래된 이미지 및 컨테이너 삭제

**배포 경로**: `$HOME/services/[service-name]`
예: `~/services/joker-backend`, `~/services/game-server`

**디스크 공간 관리**: 각 배포마다 자동으로 사용하지 않는 Docker 리소스를 정리하여 "no space left on device" 에러를 방지합니다.

## 수동 배포

### 스크립트 사용

```bash
# 기본 서비스 배포 (포트 6000, DB: joker_backend)
./scripts/deploy.sh

# 특정 서비스 배포
./scripts/deploy.sh game-server 6001 game_server

# 매개변수: [서비스명] [API포트] [DB이름]
```

### Docker Compose 직접 사용

```bash
# 서비스 시작
docker-compose -f docker-compose.prod.yml up -d --build

# 서비스 중지
docker-compose -f docker-compose.prod.yml down

# 로그 확인
docker-compose -f docker-compose.prod.yml logs -f api

# 상태 확인
docker-compose -f docker-compose.prod.yml ps
```

**참고**: `docker-compose` (하이픈 포함) 명령어를 사용합니다.

## 새 서비스 추가 방법

### 1. 환경 변수 설정

`.env.production.example`을 참고하여 새 서비스용 환경 파일 생성:

```bash
cp .env.production.example .env.game-server
```

편집:
```env
SERVICE_NAME=game-server
PORT=6001
DB_NAME=game_server
DB_PORT=3306  # 모든 서비스가 동일한 MySQL 사용
# ... 기타 설정
```

### 2. GitHub Actions 워크플로우 수정

`.github/workflows/deploy.yml`에서 환경 변수 수정:

```yaml
env:
  SERVICE_NAME: game-server
  SERVICE_PORT: 6001
  # DB_PORT는 항상 3306으로 고정
```

### 3. 데이터베이스 초기화

새 서비스용 데이터베이스를 생성합니다:

```sql
-- MySQL에 접속
mysql -u root -p -h localhost -P 3306

-- 새 데이터베이스 생성
CREATE DATABASE game_server CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 권한 부여
GRANT ALL PRIVILEGES ON game_server.* TO 'joker_user'@'%';
FLUSH PRIVILEGES;
```

### 4. 배포

```bash
# 수동 배포
./scripts/deploy.sh game-server 6001 game_server

# 또는 GitHub에 push하여 자동 배포
git push origin main
```

## 모니터링 및 로그

### 컨테이너 상태 확인

```bash
docker ps --filter "name=joker"
```

### 로그 확인

```bash
# API 서버 로그
docker logs joker-backend_api -f

# MySQL 로그
docker logs joker-backend_mysql -f

# 전체 서비스 로그
docker compose -f docker-compose.prod.yml logs -f
```

### 헬스체크

```bash
# 기본 서비스
curl http://localhost:6000/health

# 특정 서비스
curl http://localhost:6001/health
```

## 트러블슈팅

### 포트 충돌

```bash
# 포트 사용 확인
sudo lsof -i :6000

# 프로세스 종료
sudo kill -9 <PID>
```

### 컨테이너 재시작

```bash
# 특정 서비스 재시작
docker-compose -f docker-compose.prod.yml restart api

# 전체 재시작
docker-compose -f docker-compose.prod.yml restart
```

### 데이터베이스 초기화

```bash
# 주의: 모든 데이터가 삭제됩니다!
docker-compose -f docker-compose.prod.yml down -v
docker-compose -f docker-compose.prod.yml up -d
```

### 디스크 공간 확보

```bash
# 자동 정리 스크립트 사용 (권장)
./scripts/cleanup.sh

# 수동 정리
docker system prune -af --volumes

# 특정 이미지만 삭제
docker images | grep joker
docker rmi <IMAGE_ID>

# 디스크 사용량 확인
df -h
docker system df
```

**참고**: 배포 시 자동으로 정리가 수행되지만, 디스크 공간이 부족할 경우 수동으로 실행할 수 있습니다.

## 보안 고려사항

1. **Secrets 관리**: GitHub Secrets에 민감한 정보 저장
2. **포트 방화벽**: 필요한 포트만 외부에 개방
3. **정기 업데이트**: 베이스 이미지 및 의존성 정기 업데이트
4. **로그 관리**: 민감한 정보가 로그에 기록되지 않도록 주의

## 롤백 전략

### 이전 버전으로 롤백

```bash
# 1. 이전 커밋으로 체크아웃
git checkout <commit-hash>

# 2. 재배포
./scripts/deploy.sh

# 3. 또는 Docker 이미지 태그 사용
docker tag joker_backend-api:latest joker_backend-api:backup
docker-compose -f docker-compose.prod.yml up -d
```

## 참고 자료

- [GitHub Actions 문서](https://docs.github.com/en/actions)
- [Docker Compose 문서](https://docs.docker.com/compose/)
- [Self-hosted Runners 가이드](https://docs.github.com/en/actions/hosting-your-own-runners)
