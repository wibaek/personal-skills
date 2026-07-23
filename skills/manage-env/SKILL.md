---
name: manage-env
description: "AWS Systems Manager Parameter Store를 원본으로 삼아 프로젝트의 로컬 .env를 초기 등록, 갱신, 복원할 때 사용한다. 사용자가 .env 관리, env 동기화, SSM에 env 올리기, 새 PC에서 .env 복구, .env.example 정리를 요청하면 사용한다."
---

# Manage Env

AWS SSM Parameter Store의 Standard `SecureString`과 로컬 `.env`를 AWS CLI로 관리한다. `.env`를 바로 읽거나 변경하지 말고 예정 작업과 대상을 먼저 설명한다.

## 기본값

- AWS SSM을 canonical source, 로컬 `.env`를 개발용 working copy로 사용한다.
- `.env` 원문을 raw dotenv 문자열 하나로 저장한다.
- parameter 경로로 `/<project-slug>/<env>/env`를 제안한다.
- environment는 프로젝트가 사용하는 `dev`, `stg`, `prod` 체계를 일관되게 적용한다. 로컬 `.env`에는 기본적으로 `dev`를 제안하되 확정하지 않는다.
- project slug에는 안정적인 lowercase kebab-case 이름을 사용한다.
- 값은 4096 bytes 이하여야 한다. 넘으면 자동 저장하거나 자르지 말고 큰 변수나 논리적 묶음의 분리를 사용자와 결정한다.
- SSH key나 인증서라는 이유만으로 분리하지 않는다. 크기, 접근 주체, rotation 주기가 실제로 다를 때만 제안한다.
- 별도 KMS key 요구가 없으면 `--key-id`를 생략한다.

## 실행 전 확인

사용자의 “관리해줘”, “동기화해줘”를 즉시 실행 승인으로 해석하지 않는다.

1. secret이 아닌 repository 파일만 조사한다.
2. 다음 내용을 먼저 보여준다.
   - init, push/update, restore 중 수행 방향
   - source와 destination
   - AWS profile, Region, account와 parameter 경로
   - 읽거나 수정할 local file
   - SSM create 또는 update 여부
   - 기존 `.env` overwrite 여부
3. `.env` 내용 읽기, 복호화 조회, SSM write, 기존 `.env` 교체를 구분해서 알린다.
4. 사용자가 명시적으로 확인한 범위만 수행한다.
5. 양쪽이 모두 존재하면 push와 restore 중 어느 방향인지 확인한다. 자동 merge하지 않는다.

## 보안 규칙

- 승인 전에는 `.env`, `.env.*`, `.env.production`, key, certificate, credential file의 내용을 읽지 않는다. 존재 여부와 Git 추적 여부만 확인한다.
- secret 값을 응답, diff, terminal output, log에 표시하지 않는다.
- 업로드할 secret 문자열을 shell argument에 직접 쓰지 말고 `--value file://.env`를 사용한다.
- `eval`, `set -x`, `env`, `printenv`를 사용하지 않는다.
- 기존 SSM parameter와 `.env`를 기본적으로 덮어쓰지 않는다.
- 복원할 때 같은 directory에 임시 파일을 만들고 mode `0600`으로 설정한 뒤 `.env`로 이동한다.
- `.env`가 symlink이면 자동으로 읽거나 교체하지 않는다.
- `.env`가 Git에 추적됐거나 history에 들어갔다면 credential rotation을 먼저 권고한다. SSM 업로드만으로 해결됐다고 말하지 않는다.
- parameter 삭제, credential rotation, Git history rewrite는 별도 요청 없이 수행하지 않는다.

## Init

1. `.env`를 읽지 않고 다음을 확인한다.
   - repository root와 project slug 후보
   - `.env` 존재, symlink, Git 추적 여부
   - `.gitignore`, `.env.example`, README, CONTRIBUTING, `docs/`
   - project의 environment naming
2. `local .env → AWS account/profile/Region → /<project>/<env>/env` 계획을 보여주고 확인받는다.
3. 확인 후 다음 metadata만 검사한다.
   - `wc -c < .env`로 byte 크기
   - key 이름과 중복 여부. 값은 출력하지 않는다.
   - `aws sts get-caller-identity`의 account와 ARN
   - 기존 parameter의 type과 version
4. 4096 bytes를 넘으면 중단하고 분리를 논의한다.
5. `.gitignore`, `.env.example`, 복원 문서의 proposed change를 보여준다.
6. AWS create/update와 local file 변경 목록을 다시 보여주고 최종 확인받는다.
7. [references/aws-cli.md](references/aws-cli.md)의 create 또는 update 명령을 실행한다.
8. parameter가 `SecureString`이고 version이 생성됐는지 확인한다.
9. `.env`가 Git에 무시되고 `.env.example`은 추적 가능한지 확인한다.

### Project 권장사항

- `.gitignore`에는 실제 env 파일을 제외하고 example 파일은 허용한다.
- `.env.example`은 source의 key 순서와 기존 설명을 가능한 한 유지한다.
- secret 값은 비우고 안전하다고 확실한 non-secret default만 남긴다.
- 기존 example을 무조건 재생성하지 말고 proposed diff를 보여준다.
- 복원 절차는 기존 local-development 문서, README/CONTRIBUTING, 새 `docs/environment.md` 순서로 기록한다.
- 반복 가능한 project command가 필요하면 pull-only command를 제안한다. upload target은 기본 생성하지 않는다.

## Restore

1. project 문서에서 profile, Region, parameter 경로와 output 파일을 찾는다.
2. mapping이 없거나 여러 후보가 있으면 제안값을 확인받는다.
3. `AWS account/profile/Region + parameter → .env` 계획과 복호화 조회 사실을 알리고 확인받는다.
4. `aws sts get-caller-identity`로 account와 caller를 확인한다.
5. parameter가 `SecureString`인지 확인한다.
6. 기존 `.env`가 있으면 중단하고 overwrite 여부를 별도로 확인받는다.
7. [references/aws-cli.md](references/aws-cli.md)의 restore 명령으로 임시 파일에 내려받은 뒤 `.env`로 이동한다.
8. mode가 `0600`이고 `.env`가 Git에 무시되는지 확인한다.
9. secret 값을 출력하지 않고 parameter version, destination, byte 크기만 보고한다.
10. 적절한 project 문서가 없으면 같은 restore 명령의 문서화를 제안한다.

## 완료 보고

- init, update, restore 중 수행한 방향
- AWS account, profile, Region과 parameter 이름
- create, update, unchanged와 parameter version
- 변경한 project 파일
- `.env` 크기, mode, Git ignore 검증
- 수동 복원 명령을 기록한 문서 위치
- rotation이나 분리처럼 남은 작업

을 짧게 보고한다.
