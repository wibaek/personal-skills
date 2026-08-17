# Network Naming Convention

- [VPC](#1-vpc-이름-짓는-법)
- [Subnet](#2-subnet-이름-짓는-법)
- [Security Group](#3-security-group-이름-짓는-법)
- [ALB/NLB](#4-albnlb-이름-짓는-법)
- [Target Group](#5-target-group-이름-짓는-법)

## 1. VPC 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-vpc
```

같은 환경과 시스템에서 용도가 다른 VPC를 구분해야 하면:

```text
<system>-<environment>-<purpose>-vpc
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- purpose: 선택 항목. `shared`, `inspection`, `egress`
- VPC는 별도의 이름 필드가 없으므로 이 이름을 `Name` 태그 값으로 사용한다.
- `Name` 태그 값은 최대 256자로 제한한다.
- 구분할 별도 용도가 없으면 `purpose`를 생략한다. 이름에 `null`을 넣지 않는다.
- CIDR은 변경되거나 추가될 수 있으므로 이름에 넣지 않는다.

예시:

```text
ninework-production-vpc
ninework-production-shared-vpc
ninework-production-inspection-vpc
```

## 2. Subnet 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<scope>-<az>-subnet
```

용도를 구분해야 하면:

```text
<system>-<environment>-<scope>-<purpose>-<az>-subnet
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- scope: `public`, `private`, `isolated`
- purpose: 선택 항목. `app`, `db`, `cache`, `ingress`
- az: Availability Zone ID. 예: `apne2-az1`, `apne2-az2`
- Subnet은 별도의 이름 필드가 없으므로 이 이름을 `Name` 태그 값으로 사용한다.
- `public`은 인터넷 게이트웨이로 향하는 경로가 있는 Subnet에 사용한다.
- `private`은 인터넷 게이트웨이로 직접 향하는 경로가 없는 Subnet에 사용한다.
- `isolated`는 인터넷 게이트웨이나 NAT Gateway를 통한 외부 경로가 없는 Subnet에 사용한다.
- 구분할 별도 용도가 없으면 `purpose`를 생략한다. 이름에 `null`을 넣지 않는다.
- CIDR은 변경되거나 추가될 수 있으므로 이름에 넣지 않는다.
- 여러 AWS 계정에서 같은 물리적 Availability Zone을 일관되게 식별할 수 있도록 AZ 이름보다 AZ ID를 사용한다.

예시:

```text
ninework-production-public-apne2-az1-subnet
ninework-production-private-app-apne2-az1-subnet
ninework-production-private-db-apne2-az2-subnet
ninework-production-isolated-cache-apne2-az2-subnet
```

## 3. Security Group 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<component>-sg
```

같은 구성요소에서 역할을 구분해야 하면:

```text
<system>-<environment>-<component>-<purpose>-sg
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- component: `alb`, `api`, `web`, `worker`, `db`, `cache`, `bastion`, `vpc-endpoint`
- purpose: 선택 항목. `ingress`, `egress`, `client`, `management`
- Security Group 이름은 최대 255자로 제한한다.
- Security Group 이름은 같은 VPC 안에서 중복할 수 없으며 `sg-`로 시작할 수 없다.
- 연결 규칙의 포트, 프로토콜, CIDR은 변경될 수 있으므로 이름에 넣지 않는다.
- 이름과 설명은 생성 후 변경할 수 없으므로 변하지 않는 역할만 이름에 넣는다.
- 구분할 별도 역할이 없으면 `purpose`를 생략한다. 이름에 `null`을 넣지 않는다.

예시:

```text
ninework-production-alb-sg
ninework-production-api-sg
ninework-production-db-sg
ninework-production-db-client-sg
ninework-production-bastion-sg
```

## 4. ALB/NLB 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<purpose>-<type>
```

인터넷에 공개된 Load Balancer는 `public`을 추가한다.

```text
<system>-<environment>-public-<purpose>-<type>
```

- system: `ninework`, `payments`, `auth`
- environment: `dev`, `stg`, `prod`, `test`
- purpose: 선택 항목. `web`, `api`
- type: `alb`, `nlb`
- 32자 제한을 고려해 이 리소스에서는 `dev`, `stg`, `prod`, `test` 축약형을 사용한다.
- `public`은 AWS Scheme이 `internet-facing`인 Load Balancer에 사용한다.
- AWS Scheme이 `internal`이면 이름에 별도 scope를 넣지 않는다.
- 이름은 최대 32자로 제한한다.
- 이름은 계정과 Region 안에서 중복할 수 없다.
- 이름에는 영문자, 숫자, 하이픈만 사용하며 하이픈으로 시작하거나 끝낼 수 없다.
- 이름은 `internal-`로 시작할 수 없다.
- 이름은 생성 후 변경할 수 없으므로 변하지 않는 역할만 이름에 넣는다.
- 구분할 별도 용도가 없으면 `purpose`를 생략한다. 이름에 `null`을 넣지 않는다.
- 32자 제한을 고려해 `application-load-balancer` 대신 `alb`, `network-load-balancer` 대신 `nlb`를 사용한다.

예시:

```text
ninework-prod-public-alb
auth-prod-public-web-alb
auth-prod-api-alb
ninework-prod-public-nlb
auth-prod-api-nlb
```

## 5. Target Group 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<component>-tg
```

용도를 구분해야 하면:

```text
<system>-<environment>-<component>-<purpose>-tg
```

- system: `ninework`, `payments`, `auth`
- environment: `dev`, `stg`, `prod`, `test`
- component: `web`, `api`, `worker`, `admin`
- purpose: 선택 항목. `migration`
- 32자 제한을 고려해 이 리소스에서는 `dev`, `stg`, `prod`, `test` 축약형을 사용한다.
- 연결된 ALB/NLB보다 실제 트래픽을 받는 구성요소를 기준으로 짓는다.
- Target type, 프로토콜, 포트, 헬스 체크 설정은 이름에 넣지 않는다.
- 이름은 최대 32자로 제한한다.
- 이름은 계정과 Region 안에서 중복할 수 없다.
- 이름에는 영문자, 숫자, 하이픈만 사용하며 하이픈으로 시작하거나 끝낼 수 없다.
- 구분할 별도 용도가 없으면 `purpose`를 생략한다. 이름에 `null`을 넣지 않는다.
- 32자 제한을 고려해 `target-group` 대신 `tg`를 suffix로 사용한다.

예시:

```text
ninework-prod-web-tg
ninework-prod-api-tg
auth-prod-api-migration-tg
payments-prod-worker-tg
auth-prod-admin-tg
```
