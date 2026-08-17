# Security and Configuration Naming Convention

- [Secrets Manager Secret](#1-secrets-manager-secret-이름-짓는-법)
- [SSM Parameter](#2-ssm-parameter-이름-짓는-법)
- [KMS Alias](#3-kms-alias-이름-짓는-법)

## 1. Secrets Manager Secret 이름 짓는 법

기본 템플릿:

```text
<system>/<environment>/<purpose>
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- purpose: `app-env`, `db-credentials`, `jwt-signing-key`, `webhook-key`, `github-app`
- 이름은 1자 이상 512자 이하로 제한한다.
- `/`를 계층 구분자로 사용하고 각 경로 조각은 `lowercase-kebab-case`로 작성한다.
- 선행 `/`는 사용하지 않는다.
- Secret 값은 환경마다 다르므로 `environment`를 유지한다.
- 환경별 AWS 계정이 완전히 분리됐을 때만 `environment`를 생략할 수 있다.
- Secrets Manager 문맥에서 이미 리소스 종류가 드러나므로 `secret` suffix를 넣지 않는다.
- 실제 비밀번호, 토큰, 이메일 등 민감한 값은 이름에 넣지 않는다.
- ARN의 자동 suffix와 혼동될 수 있으므로 이름을 하이픈과 임의의 6자로 끝내지 않는다.
- Secret version이나 rotation 상태는 이름에 넣지 않는다.

예시:

```text
ninework/production/app-env
ninework/production/db-credentials
ninework/production/jwt-signing-key
payments/staging/webhook-key
auth/production/github-app
```

## 2. SSM Parameter 이름 짓는 법

전체 `.env`를 하나의 Parameter로 저장할 때의 권장 템플릿:

```text
/<system>/<environment>/env
```

컴포넌트 구분이 필요하면:

```text
/<system>/<environment>/<component>/env
```

환경변수를 개별 Parameter로 저장할 때의 템플릿:

```text
/<system>/<environment>/<ENVIRONMENT_VARIABLE>
```

컴포넌트 구분이 필요하면:

```text
/<system>/<environment>/<component>/<ENVIRONMENT_VARIABLE>
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- component: `api`, `web`, `worker`, `batch`
- 전체 `.env` 원문을 하나의 `SecureString`으로 저장하는 방식을 권장한다.
- 변수별 접근 권한이나 관리 주기를 분리해야 하면 환경변수를 개별 Parameter로 저장할 수 있다.
- `/`를 계층 구분자로 사용하고 경로는 선행 `/`로 시작한다.
- `system`, `environment`, `component`는 `lowercase-kebab-case`로 작성한다.
- 개별 환경변수 이름은 예외적으로 `.env`의 키와 동일한 `UPPER_SNAKE_CASE`를 사용한다.
- Parameter 이름은 AWS가 ARN에 사용하는 선행 문자 수를 포함하여 최대 1,011자로 제한하며, 계층 깊이는 최대 15단계로 제한한다.
- 이름은 대소문자, 숫자, 마침표(`.`), 하이픈(`-`), 밑줄(`_`), 슬래시(`/`)만 사용한다.
- 이름을 대소문자와 관계없이 `aws` 또는 `ssm`으로 시작하지 않는다.
- Systems Manager 문맥에서 이미 리소스 종류가 드러나므로 `ssm` 또는 `parameter` suffix를 넣지 않는다.
- 전체 `.env`를 Standard Parameter 하나에 저장할 때 값은 4,096바이트 이하여야 한다.

예시:

```text
/ninework/development/env
/ninework/production/env
/ninework/production/worker/env
/ninework/production/DATABASE_URL
/ninework/production/JWT_SECRET
/ninework/production/worker/QUEUE_URL
/payments/staging/STRIPE_SECRET_KEY
```

## 3. KMS Alias 이름 짓는 법

기본 템플릿:

```text
alias/<system>/<environment>/<purpose>
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- purpose: `app`, `rds`, `backup`, `secrets`, `signing`
- AWS KMS API에서 요구하는 `alias/` prefix를 사용한다.
- `alias/` 뒤의 각 경로 조각은 `lowercase-kebab-case`로 작성한다.
- Key의 용도가 환경마다 다르면 `environment`를 유지한다.
- 환경별 AWS 계정이 완전히 분리됐을 때만 `environment`를 생략할 수 있다.
- AWS KMS 문맥에서 이미 리소스 종류가 드러나므로 `kms` 또는 `key` suffix를 넣지 않는다.
- 이름은 `alias/`를 포함하여 1자 이상 256자 이하로 제한한다.
- 이름은 계정과 Region 안에서 중복할 수 없다.
- `alias/aws/` prefix는 AWS 관리형 키에 예약되어 있으므로 사용하지 않는다.
- 실제 키 값이나 민감한 정보는 이름에 넣지 않는다.

예시:

```text
alias/ninework/production/app
alias/ninework/production/rds
alias/ninework/production/backup
alias/payments/staging/secrets
alias/auth/production/signing
```
