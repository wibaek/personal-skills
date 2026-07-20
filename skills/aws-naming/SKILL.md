---
name: aws-naming
description: "AWS 리소스(IAM role/policy, Lambda, S3, SQS 등), Terraform/CDK 식별자, 태그, 환경명 네이밍 컨벤션을 설계하거나 리뷰할 때 사용한다."
---

# AWS 네이밍

AWS 이름을 일관되고 검색 가능하며 IaC에 적합하게 만든다. 사용자 지정 physical name, `Name` 태그, 서비스 고유 식별자, IaC 식별자를 구분해서 다룬다.

## 작업 흐름

1. 기존 repository와 조직 convention을 먼저 확인한다.
2. 이름 유형과 scope를 구분한다: global, account, region, VPC, cluster, service, module-local.
3. account와 environment 관계를 확인하고 필요한 token만 고른다.
4. 서비스별 필수 문법과 길이 제한을 확인한다.
5. live 리소스의 rename이 replacement 또는 migration을 요구하는지 표시한다.

## 이름 유형과 우선순위

- **사용자 지정 physical name**: IAM Role, Lambda function, SQS queue처럼 API나 IaC에서 직접 지정하는 이름이다. 기본 naming convention을 적용한다.
- **`Name` 태그**: VPC, subnet, route table 등에서 검색과 표시를 위해 사용하는 metadata다. physical name으로 취급하지 않는다.
- **서비스 고유 식별자**: KMS alias, SSM parameter path, CloudWatch log group, DNS name처럼 별도 문법이 있는 이름이다. 서비스 문법을 우선한다.
- **AWS-generated ID와 ARN**: `vpc-*`, `subnet-*`, KMS key ID, ARN에는 convention을 적용하지 않는다.
- **IaC 식별자**: Terraform label, CloudFormation logical ID, CDK construct ID에는 도구의 native 형식을 사용한다.

다음 우선순위를 적용한다.

1. AWS 서비스의 필수 문법과 제한
2. 기존 repository 또는 조직 convention
3. 이 skill의 기본 convention

## 기본 형식

사용자가 지정하는 physical name과 `Name` 태그 값은 기본적으로 `lowercase-kebab-case`를 사용한다.

```text
[<env>-]<system>-<component-or-purpose>[-<scope>]-<resource-kind>
```

- `env`는 `dev`, `stg`, `prod` 중 하나를 일관되게 사용한다.
- account-per-env 구조이고 account context만으로 환경이 명확하면 `env`를 생략할 수 있다. cross-account 검색이나 export에서 모호해지면 유지한다.
- `system`은 짧고 안정적인 이름을 사용한다.
- `scope`에는 꼭 필요한 Region, AZ, tenant discriminator, variant만 넣는다.
- `resource-kind`는 기본적으로 마지막 token으로 둔다. AWS가 prefix, suffix, path, DNS 형식을 요구하면 서비스 문법을 우선한다.
- 사람 이름, 이메일, ticket ID, 임시 project명, secret, PII, 중복되거나 쉽게 변하는 정보는 넣지 않는다.
- 긴 설명과 ownership, cost, compliance metadata는 tag, description, 문서로 분리한다.

예시:

```text
prod-ninework-api-lambda
prod-ninework-event-queue
ninework-order-table
```

마지막 예시는 environment별 account가 분리된 경우다.

### 길이 줄이기

서비스 제한을 넘거나 가까워지면 다음 순서로 줄인다.

1. account나 stack context에 이미 있는 중복 token과 선택 token을 제거한다.
2. `system`을 조직에서 문서화한 안정적인 약어로 축약한다.
3. 자주 쓰는 용어를 registry에 등록된 약어로 축약한다.
4. `purpose`와 `resource-kind`는 마지막까지 보존하고 임의 절단은 피한다.

## Location Token

- Region은 `us-east-1` 같은 공식 Region code를 사용한다.
- 단일 account 안에서 AZ를 구분하면 `us-east-1a` 같은 AZ name을 사용한다.
- multi-account, shared VPC처럼 같은 물리 AZ를 맞춰야 하면 `use1-az1` 같은 AZ ID를 사용한다.
- `use1`, `use1a` 같은 자체 shorthand는 공식 값처럼 암묵적으로 쓰지 않는다. 필요하면 별도 registry에 정의한다.
- Region이나 AZ가 account, stack, resource context에서 이미 명확하면 생략한다.
- route table은 여러 subnet과 연결될 수 있으므로 전용 설계가 아니면 AZ token을 넣지 않는다.

```text
prod-ninework-private-us-east-1a-subnet
prod-ninework-private-use1-az1-subnet
prod-ninework-public-route-table
```

## 서비스별 예외

| 대상 | 규칙 | 예시 |
| --- | --- | --- |
| KMS alias | API와 CLI에서는 `alias/` prefix를 사용한다. Console은 자동으로 붙인다. | `alias/prod-ninework-order-data` |
| Lambda log group | 기본 log group 형식을 유지한다. | `/aws/lambda/prod-ninework-api-lambda` |
| Custom log group | `/` hierarchy를 사용할 수 있다. | `/prod/ninework/api` |
| SQS/SNS FIFO | resource kind 뒤에 `.fifo` suffix를 붙인다. | `prod-ninework-event-queue.fifo` |
| SSM Parameter | 필요하면 `/` hierarchy를 사용한다. | `/prod/ninework/api/database-url` |
| ECR repository | 필요하면 `/` namespace를 사용한다. | `prod/ninework/api` |
| RDS DB instance | 문자로 시작하고 끝 하이픈과 연속 하이픈을 금지한다. | `prod-ninework-order-db-instance` |
| ACM/Route 53 | resource naming template 대신 FQDN과 DNS 규칙을 적용한다. | `api.example.com` |

### S3 Bucket

bucket 이름은 전역에서 충돌할 수 있으므로 필요한 경우 organization, account ID, Region을 안정적인 discriminator로 넣는다.

```text
<org>-[<env>-]<system>-<purpose>[-<account-id>][-<region>]-bucket
```

```text
acme-prod-ninework-artifact-123456789012-us-east-1-bucket
```

충돌 방지에 필요하지 않은 discriminator와 confidential business detail은 넣지 않는다.

## IAM

IAM 이름은 대상보다 존재 이유와 권한 경계를 우선해서 드러낸다. account-per-env 구조에서는 앞의 규칙에 따라 `env`를 생략할 수 있다.

### Role

```text
[<env>-]<system>[-<component>][-<trust-source>]-<purpose>-role
```

- `component`: `api`, `worker`, `batch`처럼 workload를 구분한다.
- `trust-source`: `ec2`, `ecs-task`, `lambda`, `github`처럼 assume 경계를 구분할 때만 넣는다.
- `purpose`: `runtime`, `deploy`, `migration`, `backup`, `readonly`, `breakglass`처럼 구체적으로 적는다.
- system 안에 workload가 여러 개이거나 권한 경계가 다르면 `component`를 넣는다.
- 같은 목적의 Role을 trust boundary로 구분해야 할 때만 `trust-source`를 넣는다.
- Role name은 최대 64자다. Console Switch Role을 사용하면 path와 RoleName 합계도 64자를 넘기지 않는다.

```text
prod-ninework-api-ec2-runtime-role
prod-ninework-api-github-deploy-role
prod-ninework-batch-migration-role
```

### Policy

```text
[<env>-]<system>-<scope>-<capability>-policy
```

- `scope`는 `artifact`, `event`, `secret`, `observability`처럼 실제 business 또는 resource boundary를 나타낸다.
- `capability`는 `read`, `write`, `publish`, `consume`, `encrypt`, `decrypt`, `manage`를 사용한다.
- 단순 CRUD로 매핑하지 말고 실제 IAM action의 의미를 드러낸다.
- `s3-read`, `cloudwatch-read`처럼 서비스만 적은 scope는 해당 서비스 전체가 의도된 범위일 때만 사용한다.
- scope별 권한과 lifecycle이 다르면 policy를 분리한다. 이름만을 위해 불필요하게 쪼개지는 않는다.
- Policy name은 최대 128자다.

```text
prod-ninework-artifact-read-policy
prod-ninework-event-publish-policy
prod-ninework-secret-read-policy
```

### User와 Group

```text
svc-<system>-<purpose>-user
<team-or-function>-<privilege>-group
```

- IAM User는 가능하면 사람 계정보다 서비스성 계정에 한정하고 `svc-` prefix를 사용한다.
- User 이름에 사람 이름이나 이메일을 넣지 않는다.
- Group은 `readonly`, `deploy`, `poweruser`, `admin`, `auditor`처럼 권한 수준을 드러낸다.
- 사람 분류용 Group과 권한 부여용 Group을 구분한다.
- User name은 최대 64자, Group name은 최대 128자다.

```text
svc-ninework-backup-user
developer-readonly-group
```

## Compute와 기타 리소스

### Lambda

```text
[<env>-]<system>[-<trigger>]-<purpose>-lambda
```

- 구현 방식보다 함수 역할을 우선한다.
- `trigger`는 `sqs`, `eventbridge`, `s3`처럼 중요한 구분점일 때만 넣는다.
- `handler`, `function`, `lambda-function` 같은 중복 표현은 넣지 않는다.

```text
prod-ninework-api-lambda
prod-ninework-sqs-worker-lambda
prod-auth-cleanup-lambda
```

### Resource Kind

- 가능한 경우 AWS service 이름보다 실제 resource kind를 사용한다: `queue`, `topic`, `table`, `repository`.
- `vpc`, `sg`, `alb`, `nlb`, `tg`처럼 널리 쓰이고 명확한 약어는 사용할 수 있다. `rt` 같은 추가 약어는 조직 registry에 정의한다.
- `kms`, `iam`, `rds`, `cdn`처럼 구체 리소스를 특정하지 못하는 suffix는 단독으로 쓰지 않는다.
- `lambda`는 이 convention에서 Lambda function을 나타내는 예외적인 resource kind로 사용한다.

```text
prod-ninework-api-alb
prod-ninework-order-table
prod-ninework-container-repository
```

VPC, subnet, route table의 다음 값은 physical name이 아니라 `Name` 태그 예시다.

```text
prod-ninework-vpc
prod-ninework-private-use1-az1-subnet
prod-ninework-public-route-table
```

## IaC 식별자

- **Terraform**: snake_case label을 사용한다: `aws_sqs_queue.event`, `aws_iam_role.task_execution`. reusable module 안에서는 변수로 받은 system과 environment를 label에서 생략한다.
- **CloudFormation**: PascalCase logical ID를 사용한다: `EventQueue`, `TaskExecutionRole`. stack이 environment별이면 environment를 logical ID에 넣지 않는다.
- **CDK**: construct ID를 stable하고 semantic하게 둔다. 외부 시스템이 physical name에 의존할 때만 이름을 명시하고, 그렇지 않으면 generated name을 허용한다.

## 태그

- 기존 조직의 tag key convention이 있으면 casing을 유지한다.
- greenfield multi-account 환경에서는 `<organization>:<lowercase-hyphen>` 형식을 고려한다.
- EC2/VPC console 관례가 필요한 경우 `Name` key는 대문자를 유지한다.
- `Name` 태그는 검색과 표시 이점이 있을 때만 넣고, 이미 명확한 physical name을 무조건 복제하지 않는다.
- `aws:` prefix는 AWS 예약이므로 사용자 정의 tag key에 사용하지 않는다.
- tag key와 value에 secret, PII, 사람 이름, 이메일, confidential customer detail을 넣지 않는다.
- multi-account 표준화가 필요하면 AWS Organizations tag policy를 고려한다.

```yaml
Name: "prod-ninework-private-use1-az1-subnet"
"acme:system": "ninework"
"acme:environment": "prod"
"acme:component": "api"
"acme:managed-by": "terraform"
"acme:owner": "platform"
```

`environment`의 허용값은 `dev`, `stg`, `prod`다. `CostCenter`, `DataClassification`, `Repository`, `ServiceTier`는 조직에서 실제로 사용할 때만 추가한다.

## 리뷰 체크리스트

- physical name, `Name` 태그, 서비스 고유 식별자, AWS-generated ID, IaC 식별자를 먼저 구분한다.
- account-per-env 여부와 cross-account 검색 context를 확인한다.
- 서비스 필수 문법, 허용 문자, 길이 제한을 확인한다.
- Region과 AZ가 공식 code, AZ name, AZ ID 중 올바른 값인지 확인한다.
- 중복 정보, 불안정한 구현 detail, 민감 정보가 들어갔는지 확인한다.
- ownership, cost, compliance metadata는 긴 이름보다 tag를 우선한다.
- live physical resource rename의 replacement와 migration risk를 표시한다.

## 출력 형식

필요하면 `resource`, `name type`, `physical name or tag`, `IaC identifier`, `tags` column을 가진 compact table로 반환한다. convention 예외, service constraint, migration risk만 짧게 덧붙인다.
