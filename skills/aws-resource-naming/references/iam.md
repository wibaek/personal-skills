# IAM Naming Convention

- [기본 규칙](#1-기본-규칙)
- [Role](#2-role-이름-짓는-법)
- [Policy](#3-policy-이름-짓는-법)
- [User](#4-user-이름-짓는-법)
- [Group](#5-group-이름-짓는-법)

## 1. 기본 규칙

- 모든 이름은 `lowercase-kebab-case`를 사용한다.
- 단어는 하이픈(`-`)으로 구분한다.
- 리소스 타입은 suffix로 통일한다: `-role`, `-policy`, `-user`, `-group`
- IAM Role과 User 이름은 최대 64자, Policy와 Group 이름은 최대 128자로 제한한다.
- AWS Management Console에서 Switch Role을 사용할 경우 path와 Role 이름을 합쳐 최대 64자로 제한한다.
- 이름은 짧되 의미가 드러나야 한다.
- 중복 정보는 제거한다.
- 이름에는 **변하지 않는 정보만** 넣는다.
- 설명이 길어지면 태그, 설명(description), 문서로 분리한다.
- 이름에 넣지 않는 정보: 사람 이름, 이메일, 티켓 번호, 임시 프로젝트명, 중복 정보, 없어도 문맥상 충분한 기술명

---

## 2. Role 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<platform>-<purpose>-role
```

플랫폼 정보가 불필요하면:

```text
<system>-<environment>-<purpose>-role
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- platform: `ec2`, `github`
- purpose: `runtime`, `deploy`, `ci`, `cd`, `worker`, `batch`, `cron`, `migration`, `backup`, `restore`, `monitoring`, `maintenance`, `bootstrap`, `read-only`
- Role 이름은 무엇에 붙는가보다 왜 존재하는가를 우선한다.
- `purpose`는 반드시 구체적으로 적는다.
- `platform`은 구분이 필요할 때만 넣는다.

예시:

```text
ninework-production-ec2-runtime-role
ninework-production-ec2-worker-role
ninework-production-ec2-batch-role
ninework-production-github-deploy-role
ninework-production-github-ci-role
ninework-production-migration-role
```

---

## 3. Policy 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<resource>-<access>-policy
```

세부 범위가 필요하면:

```text
<system>-<environment>-<resource>-<detail>-<access>-policy
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- resource: `assets`, `secrets`, `cloudwatch`, `s3`
- detail: `public-private`, `sqlite-backup`, `project-env`
- access: `read`, `write`, `crud`
- Policy 이름은 대상 리소스, 세부 범위, 권한 수준이 드러나야 한다.
- `crud`는 실제로 create/read/update/delete 성격일 때만 사용한다.
- 읽기 전용이면 `read`, 쓰기 중심이면 `write`를 사용한다.

예시:

```text
ninework-production-assets-public-private-crud-policy
ninework-production-assets-sqlite-backup-write-policy
ninework-production-secrets-project-env-read-policy
ninework-production-cloudwatch-read-policy
ninework-production-s3-read-policy
```

---

## 4. User 이름 짓는 법

기본 템플릿:

```text
svc-<system>-<purpose>-user
```

- system: `ninework`, `payments`, `auth`
- purpose: `backup`, `deploy`, `batch`, `legacy-integration`
- IAM User는 가능하면 사람 계정보다 **서비스성 계정**에 한정해서 사용한다.
- 서비스용 계정은 `svc-` prefix로 구분한다.
- 사람 이름이나 이메일은 넣지 않는다.

예시:

```text
svc-ninework-backup-user
svc-ninework-deploy-user
svc-ninework-legacy-integration-user
svc-payments-batch-user
```

---

## 5. Group 이름 짓는 법

기본 템플릿:

```text
<team-or-function>-<privilege>-group
```

직무 분류용이면:

```text
<job-family>-group
```

- team-or-function: `developer`, `data-science`, `platform`, `security`, `billing`
- privilege: `read-only`, `deploy`, `admin`, `auditor`, `s3-read-write`
- Group은 직무명만 쓰기보다 **권한 수준이 드러나게** 짓는다.
- 사람 분류용 그룹과 권한 부여용 그룹은 구분한다.
- `console-users-group` 같은 그룹은 공통 baseline 용도로만 쓰고, 실제 업무 권한 그룹과 분리한다.

예시:

```text
console-users-group
developer-read-only-group
developer-deploy-group
data-science-read-only-group
data-science-s3-read-write-group
platform-admin-group
security-auditor-group
billing-read-only-group
```
