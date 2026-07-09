---
name: aws-naming
description: "AWS 리소스 이름, IaC 식별자, 태그, 환경명 네이밍 규칙을 설계하거나 리뷰할 때 사용한다."
---

# AWS 네이밍

AWS 이름을 일관되고 검색 가능하며 IaC에 적합하게 만든다. 물리 AWS 리소스 이름, IaC 내부 식별자, 태그는 목적이 다르므로 같은 형식을 무조건 복사하지 않는다.

## 작업 흐름

1. 기존 repository 또는 platform convention을 먼저 확인한다.
2. 리소스 scope를 구분한다: global, account, region, VPC, cluster, service, module-local.
3. 환경명 vocabulary를 하나로 고른다. 기존 convention이 없으면 `dev`, `stg`, `prod`를 사용한다.
4. 필요한 token만 사용해서 이름을 만든다.
5. 최종 확정 전에 service-specific 제약을 확인한다. 특히 S3, IAM, CloudWatch, Route 53, ACM, KMS alias, 길이 제한에 가까운 리소스를 주의한다.
6. 기존 live 리소스를 리뷰하는 경우 rename impact를 표시한다. 많은 AWS 리소스 이름 변경은 replacement 또는 migration을 요구한다.

## 기본 규칙

- 모든 이름은 `lowercase-kebab-case`를 사용한다.
- 단어는 하이픈(`-`)으로 구분한다.
- 리소스 타입은 suffix로 통일한다: `-role`, `-policy`, `-user`, `-group`, `-lambda`, `-sqs`, `-sns`, `-ddb`.
- 이름은 짧되 의미가 드러나야 한다.
- 중복 정보는 제거한다.
- 이름에는 변하지 않는 정보만 넣는다.
- 설명이 길어지면 태그, 설명(description), 문서로 분리한다.
- 이름에 넣지 않는다: 사람 이름, 이메일, 티켓 번호, 임시 프로젝트명, 중복 정보, 없어도 문맥상 충분한 기술명.
- 같은 프로젝트에서 `dev`와 `staging`, `development`와 `prod`처럼 환경명 체계를 섞지 않는다.
- 기존 convention이 없으면 DynamoDB table이나 database table 이름을 plural로 만들지 않는다.

## 기본 Token 모델

사람이 보는 AWS 물리 리소스 이름은 기본적으로 다음 순서를 사용한다.

```text
<env>-<system>-<component-or-purpose>-<resource>[-<qualifier>]
```

예시:

```text
prod-ninework-api-lambda
prod-ninework-events-sqs
prod-ninework-private-subnet-use1a
prod-ninework-orders-ddb
```

규칙:

- `env`: `dev`, `stg`, `prod`.
- `system`: `ninework`, `payments`, `auth`처럼 짧고 안정적인 system 이름.
- multi-region 또는 zonal 리소스 구분이 필요할 때만 region이나 AZ를 넣는다.
- 구현체 이름보다 의미 있는 component 또는 purpose를 선호한다: `checkout`, `events`, `worker`, `api`, `ingest`.
- customer name, secret, 내부 ticket ID, 개인 이름, 임시 농담을 이름에 넣지 않는다.

## IAM

IAM 이름은 대상보다 존재 이유와 권한 경계를 우선해서 드러낸다.

### Role

기본 템플릿:

```text
<env>-<system>-<platform>-<purpose>-role
```

platform 정보가 불필요하면:

```text
<env>-<system>-<purpose>-role
```

Token:

- `env`: `dev`, `stg`, `prod`
- `system`: `ninework`, `payments`, `auth`
- `platform`: `ec2`, `github`
- `purpose`: `runtime`, `deploy`, `ci`, `cd`, `worker`, `batch`, `cron`, `migration`, `backup`, `restore`, `monitoring`, `maintenance`, `bootstrap`, `readonly`, `breakglass`

규칙:

- Role 이름은 무엇에 붙는가보다 왜 존재하는가를 우선한다.
- `purpose`는 반드시 구체적으로 적는다.
- `platform`은 구분이 필요할 때만 넣는다.

예시:

```text
prod-ninework-ec2-runtime-role
prod-ninework-ec2-worker-role
prod-ninework-ec2-batch-role
prod-ninework-github-deploy-role
prod-ninework-github-ci-role
prod-ninework-migration-role
```

### Policy

기본 템플릿:

```text
<env>-<system>-<resource>-<access>-policy
```

세부 범위가 필요하면:

```text
<env>-<system>-<resource>-<detail>-<access>-policy
```

Token:

- `env`: `dev`, `stg`, `prod`
- `system`: `ninework`, `payments`, `auth`
- `resource`: `assets`, `secrets`, `cloudwatch`, `s3`
- `detail`: `public-private`, `sqlite-backup`, `project-env`
- `access`: `read`, `write`, `crud`

규칙:

- Policy 이름은 대상 리소스, 세부 범위, 권한 수준이 드러나야 한다.
- `crud`는 실제로 create/read/update/delete 성격일 때만 사용한다.
- 읽기 전용이면 `read`, 쓰기 중심이면 `write`를 사용한다.

예시:

```text
prod-ninework-assets-public-private-crud-policy
prod-ninework-assets-sqlite-backup-write-policy
prod-ninework-secrets-project-env-read-policy
prod-ninework-cloudwatch-read-policy
prod-ninework-s3-read-policy
```

### User

기본 템플릿:

```text
svc-<system>-<purpose>-user
```

Token:

- `system`: `ninework`, `payments`, `auth`
- `purpose`: `backup`, `deploy`, `batch`, `legacy-integration`

규칙:

- IAM User는 가능하면 사람 계정보다 서비스성 계정에 한정해서 사용한다.
- 서비스용 계정은 `svc-` prefix로 구분한다.
- 사람 이름이나 이메일은 넣지 않는다.

예시:

```text
svc-ninework-backup-user
svc-ninework-deploy-user
svc-ninework-legacy-integration-user
svc-payments-batch-user
```

### Group

기본 템플릿:

```text
<team-or-function>-<privilege>-group
```

직무 분류용이면:

```text
<job-family>-group
```

Token:

- `team-or-function`: `developer`, `data-science`, `platform`, `security`, `billing`
- `privilege`: `readonly`, `deploy`, `poweruser`, `admin`, `auditor`, `s3-readwrite`

규칙:

- Group은 직무명만 쓰기보다 권한 수준이 드러나게 짓는다.
- 사람 분류용 그룹과 권한 부여용 그룹은 구분한다.
- `console-users-group` 같은 그룹은 공통 baseline 용도로만 쓰고, 실제 업무 권한 그룹과 분리한다.

예시:

```text
console-users-group
developer-readonly-group
developer-deploy-group
developer-poweruser-group
data-science-readonly-group
data-science-s3-readwrite-group
platform-admin-group
security-auditor-group
billing-readonly-group
```

## Compute

### Lambda

기본 템플릿:

```text
<env>-<system>-<purpose>-lambda
```

구분 정보가 더 필요하면:

```text
<env>-<system>-<trigger>-<purpose>-lambda
```

Token:

- `env`: `dev`, `stg`, `prod`
- `system`: `ninework`, `payments`, `auth`
- `trigger`: `apigw`, `sqs`, `eventbridge`, `s3`
- `purpose`: `api`, `webhook`, `worker`, `batch`, `cron`, `sync`, `migration`, `cleanup`, `notifier`

규칙:

- Lambda 이름은 구현 방식보다 함수의 역할을 우선한다.
- `handler`, `function`, `lambda-function` 같은 중복 표현은 넣지 않는다.
- trigger가 이름에서 중요한 구분점일 때만 `trigger`를 넣는다.

예시:

```text
prod-ninework-api-lambda
prod-ninework-webhook-lambda
prod-ninework-sqs-worker-lambda
prod-ninework-eventbridge-cron-lambda
prod-payments-batch-lambda
prod-auth-cleanup-lambda
```

## 기타 물리 리소스

AWS console, log, cross-service reference에서 모호해질 수 있으면 명시적인 type suffix를 붙인다.

```text
<env>-<system>-<component>-vpc
<env>-<system>-<component>-sg
<env>-<system>-<component>-alb
<env>-<system>-<component>-tg
<env>-<system>-<component>-sqs
<env>-<system>-<component>-sns
<env>-<system>-<component>-ddb
<env>-<system>-<component>-kms
<env>-<system>-<component>-ecr
```

AWS 약어는 문맥상 명확할 때만 사용한다: `vpc`, `sg`, `alb`, `nlb`, `tg`, `iam`, `kms`, `s3`, `sqs`, `sns`, `ecr`, `ecs`, `eks`, `rds`, `ddb`, `waf`, `cdn`.

### S3

S3 bucket 이름은 global이고 user-visible인 경우가 많다. 충돌 가능성이 있으면 organization 또는 account boundary를 포함한다.

```text
<org>-<env>-<system>-<purpose>[-<region>]
```

예시:

```text
acme-prod-ninework-artifacts-use1
```

bucket 이름에 confidential business detail을 넣지 않는다.

### Network

subnet과 route table은 유용할 때만 scope와 zone을 포함한다.

```text
prod-ninework-vpc
prod-ninework-private-subnet-use1a
prod-ninework-public-rt-use1
```

security group 이름은 protocol만 쓰지 말고 보호 대상 component를 나타낸다.

```text
prod-ninework-api-sg
prod-ninework-alb-sg
```

## IaC 식별자

물리 이름을 IaC-local identifier에 그대로 복사하지 않는다.

Terraform:

- snake_case label을 사용한다: `aws_sqs_queue.events`, `aws_iam_role.task_execution`.
- reusable module 안에서는 system과 environment를 local label에서 생략한다. 변수로 이미 전달되기 때문이다.
- 여러 리소스가 token을 공유하면 `name` 또는 `name_prefix` local로 물리 이름 생성을 모은다.

CloudFormation:

- PascalCase logical ID를 사용한다: `EventsQueue`, `TaskExecutionRole`, `PrivateSubnetUse1A`.
- 같은 stack에 여러 environment가 들어있지 않다면 logical ID에 environment 이름을 넣지 않는다.

CDK:

- construct ID는 stable하고 semantic하게 둔다.
- 외부 시스템이 물리 이름에 의존할 때만 physical name을 명시한다. 그렇지 않으면 replacement safety를 위해 generated name을 허용한다.

## 태그

이름에 과하게 넣지 말아야 할 metadata는 tag로 둔다.

```yaml
Name: "<physical-name>"
System: "<system>"
Environment: "dev|stg|prod"
Component: "<component>"
ManagedBy: "terraform|cloudformation|cdk|manual"
Owner: "<team>"
```

`CostCenter`, `DataClassification`, `Repository`, `ServiceTier`는 조직에서 실제로 사용할 때만 추가한다.

## 리뷰 체크리스트

AWS naming을 리뷰할 때:

- 현재 convention을 먼저 식별한다.
- 제안한 rename이 live physical resource 변경인지 확인한다.
- 환경명 vocabulary가 섞여 있으면 지적한다.
- 민감 정보나 customer-specific detail이 이름에 들어가면 지적한다.
- 불안정한 구현 detail을 이름에 encoding하면 지적한다.
- ownership, cost, compliance, classification metadata는 긴 이름보다 tag를 선호한다.
- uppercase, underscore, dot, slash, prefix를 쓰거나 길이 제한에 가까우면 정확한 AWS service constraint를 확인한다.

## 출력 형식

이름을 설계할 때는 필요하면 `resource`, `physical name`, `IaC identifier`, `tags` column을 가진 compact table로 반환한다. 기본 convention에서 벗어난 이유, service constraint, migration risk만 짧게 덧붙인다.
