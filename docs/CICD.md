# CI/CD 설정 가이드

## 개요

이 프로젝트는 GitHub Actions와 Self-hosted Runner를 사용하여 **경로 기반 자동 배포**를 수행합니다.

## 아키텍처

```
GitHub Repository (push)
  → GitHub Actions (경로 감지)
    → Self-hosted Runner
      → 변경된 서비스만 Docker Build & Deploy
```

- **모노레포 구조**: 모든 서비스가 하나의 레포지토리에 존재
- **경로 기반 배포**: 변경된 서비스만 자동 배포
- **빌드 서버 = 배포 서버**: 동일한 서버에서 빌드와 배포 수행
- **서비스별 독립 컨테이너**: 각 서비스는 고유한 포트로 실행
- **공유 리소스**: MySQL과 데이터베이스는 모든 서비스가 공유

## 서비스 포트 구조

각 서비스는 독립된 포트에서 실행되며, **MySQL(3306)과 데이터베이스(backend_dev)를 공유**합니다:

| 서비스 | 경로 | API 포트 | 상태 | 설명 |
|--------|------|----------|------|------|
| Auth Service | `services/auth-service/` | 6000 | ✅ 운영중 | 사용자 인증 및 권한 관리 |
| Game Service | `services/game-service/` | 6001 | 📋 예정 | 게임 로직 및 매칭 |
| Payment Service | `services/payment-service/` | 6002 | 📋 예정 | 결제 처리 |

**주요 특징:**
- **공유 MySQL**: 모든 서비스가 3306 포트의 MySQL 인스턴스 공유
- **공유 데이터베이스**: 모든 서비스가 `backend_dev` 데이터베이스 사용
- **독립 API 포트**: 각 서비스는 고유한 포트에서 실행
- **경로 기반 트리거**: 변경된 서비스만 재배포하여 효율적인 CI/CD

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

**경로 기반 자동 감지:**
- `services/auth-service/` 변경 → Auth Service만 배포
- `services/game-service/` 변경 → Game Service만 배포
- `services/payment-service/` 변경 → Payment Service만 배포
- `shared/` 변경 → **모든 서비스** 재배포 (공통 코드 변경)
- `scripts/` 또는 `.github/workflows/` 변경 → 영향받는 서비스 배포

**브랜치:**
- `main` 브랜치에 push → 프로덕션 배포
- `develop` 브랜치에 push → 스테이징 배포

**수동 실행:**
- GitHub Actions UI에서 workflow_dispatch
- 특정 서비스 선택 배포 가능 (auth-service, game-service, payment-service, all)

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

**MySQL 컨테이너 재사용**:
- 3306 포트를 사용 중인 MySQL이 있으면 재사용 (컨테이너 이름 무관)
- 없으면 새로 `joker_mysql` 컨테이너 생성
- 각 배포 시 API 컨테이너만 재빌드하여 빠른 배포 (`--no-deps` 플래그 사용)
- 기존 MySQL의 `backend_dev` 데이터베이스 접근
- API 컨테이너는 자동으로 감지된 MySQL 컨테이너 이름으로 연결 (DB_HOST)

## 수동 배포

### 통합 배포 스크립트 사용

```bash
# Auth Service 배포
./scripts/deploy-service.sh auth-service 6000

# Game Service 배포
./scripts/deploy-service.sh game-service 6001

# Payment Service 배포
./scripts/deploy-service.sh payment-service 6002

# 매개변수: [서비스명] [API포트]
# 모든 서비스가 backend_dev 데이터베이스 사용
# 서비스명은 services/ 디렉토리 이름과 일치해야 함
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
DB_NAME=backend_dev  # 모든 서비스가 동일한 DB 사용
DB_PORT=3306  # 모든 서비스가 동일한 MySQL 사용
# ... 기타 설정
```

### 2. GitHub Actions 워크플로우 수정

`.github/workflows/deploy.yml`에서 환경 변수 수정:

```yaml
env:
  SERVICE_NAME: game-server
  SERVICE_PORT: 6001
  # DB는 backend_dev로 모든 서비스 공유
```

### 3. 배포

```bash
# 수동 배포
./scripts/deploy.sh game-server 6001

# 또는 GitHub에 push하여 자동 배포
git push origin main
```

**참고**: 모든 서비스가 `backend_dev` 데이터베이스를 공유하므로 별도의 데이터베이스 생성이 필요하지 않습니다.

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
