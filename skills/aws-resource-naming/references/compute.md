# Compute Naming Convention

- [Lambda](#1-lambda-이름-짓는-법)
- [EC2](#2-ec2-이름-짓는-법)
- [ECS Cluster](#3-ecs-cluster-이름-짓는-법)
- [ECS Service](#4-ecs-service-이름-짓는-법)
- [ECS Task Definition](#5-ecs-task-definition-이름-짓는-법)
- [ECR Repository](#6-ecr-repository-이름-짓는-법)

## 1. Lambda 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<purpose>-lambda
```

구분 정보가 더 필요하면:

```text
<system>-<environment>-<trigger>-<purpose>-lambda
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- trigger: `apigw`, `sqs`, `eventbridge`, `s3`
- purpose: `api`, `webhook`, `worker`, `batch`, `cron`, `sync`, `migration`, `cleanup`, `notifier`
- Lambda 함수 이름은 최대 64자로 제한한다.
- Lambda 이름은 구현 방식보다 함수의 역할을 우선한다.
- `handler`, `function`, `lambda-function` 같은 중복 표현은 넣지 않는다.
- 트리거가 이름에서 중요한 구분점일 때만 `trigger`를 넣는다.

예시:

```text
ninework-production-api-lambda
ninework-production-webhook-lambda
ninework-production-sqs-worker-lambda
ninework-production-eventbridge-cron-lambda
payments-production-batch-lambda
auth-production-cleanup-lambda
```

## 2. EC2 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<purpose>-ec2
```

고정 인스턴스를 구분해야 하면:

```text
<system>-<environment>-<purpose>-<index>-ec2
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- purpose: `web`, `api`, `worker`, `batch`, `bastion`, `migration`
- index: 선택 항목. `01`, `02`
- EC2 이름은 실제 리소스 이름이 아니라 `Name` 태그로 지정한다.
- `Name` 태그 값은 최대 256자로 제한한다.
- Instance ID, IP, Instance type, AZ, AMI처럼 변경되거나 자동 생성되는 정보는 이름에 넣지 않는다.
- Auto Scaling 인스턴스는 개별 구분이 필요하지 않으면 `index`를 생략한다.

예시:

```text
ninework-production-web-ec2
ninework-production-worker-ec2
ninework-production-bastion-ec2
ninework-production-web-01-ec2
payments-staging-batch-ec2
```

## 3. ECS Cluster 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-ecs-cluster
```

클러스터 용도를 구분해야 하면:

```text
<system>-<environment>-<purpose>-ecs-cluster
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- purpose: 선택 항목. `batch`, `shared`
- 이름은 최대 255자로 제한한다.
- AWS는 영문자, 숫자, 하이픈, 밑줄을 허용하지만 공통 규칙에 따라 소문자와 하이픈만 사용한다.
- 하나의 환경에 클러스터가 하나면 `purpose`를 생략한다.
- Capacity Provider나 `fargate`, `ec2` 같은 실행 방식은 이름에 넣지 않는다.
- `default` 클러스터 대신 명명 규칙을 적용한 클러스터를 사용한다.

예시:

```text
ninework-production-ecs-cluster
ninework-production-batch-ecs-cluster
ninework-production-shared-ecs-cluster
payments-staging-ecs-cluster
```

## 4. ECS Service 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<component>-ecs-service
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- component: `web`, `api`, `worker`, `batch`, `scheduler`
- 이름은 최대 255자로 제한한다.
- 이름은 같은 ECS Cluster 안에서 중복할 수 없다.
- AWS는 영문자, 숫자, 하이픈, 밑줄을 허용하지만 공통 규칙에 따라 소문자와 하이픈만 사용한다.
- Desired count, Task Definition revision, Capacity Provider, `fargate`, `ec2` 같은 실행 설정은 이름에 넣지 않는다.
- 서비스가 담당하는 애플리케이션 역할을 `component`로 사용한다.

예시:

```text
ninework-production-web-ecs-service
ninework-production-api-ecs-service
ninework-production-worker-ecs-service
payments-staging-batch-ecs-service
auth-production-scheduler-ecs-service
```

## 5. ECS Task Definition 이름 짓는 법

Task Definition의 `family` 이름에 다음 템플릿을 적용한다.

```text
<system>-<environment>-<component>-task-definition
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- component: `web`, `api`, `worker`, `batch`, `scheduler`
- `family` 이름은 최대 255자로 제한한다.
- AWS는 영문자, 숫자, 하이픈, 밑줄을 허용하지만 공통 규칙에 따라 소문자와 하이픈만 사용한다.
- AWS가 revision을 `:1`, `:2`처럼 자동으로 붙이므로 이름에 revision을 넣지 않는다.
- Image tag, CPU, Memory, Architecture, `fargate`, `ec2` 같은 실행 설정은 이름에 넣지 않는다.
- 같은 애플리케이션 역할의 변경 버전은 새로운 이름 대신 같은 `family`의 revision으로 관리한다.

예시:

```text
ninework-production-web-task-definition
ninework-production-api-task-definition
ninework-production-worker-task-definition
payments-staging-batch-task-definition
auth-production-scheduler-task-definition
```

## 6. ECR Repository 이름 짓는 법

기본 템플릿:

```text
<system>/<component>
```

- system: `ninework`, `payments`, `auth`
- component: `web`, `api`, `worker`, `batch`
- Repository 이름은 2자 이상 256자 이하로 제한한다.
- `/`를 namespace 구분자로 사용하고 각 경로 조각은 `lowercase-kebab-case`로 작성한다.
- 같은 이미지를 여러 환경으로 승격하므로 `environment`는 이름에 넣지 않는다.
- ECR 문맥에서 이미 리소스 종류가 드러나므로 `ecr`, `repository` suffix를 넣지 않는다.
- Image tag, digest, 애플리케이션 버전은 Repository 이름에 넣지 않는다.
- 독립적으로 빌드하고 배포하는 이미지 단위로 Repository를 분리한다.

예시:

```text
ninework/api
ninework/web
ninework/worker
payments/batch
```
