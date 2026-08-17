# Storage and Database Naming Convention

- [S3 Bucket](#1-s3-bucket-이름-짓는-법)
- [RDS DB Instance](#2-rds-db-instance-이름-짓는-법)
- [RDS DB Cluster](#3-rds-db-cluster-이름-짓는-법)

## 1. S3 Bucket 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<purpose>
```

전역 이름 충돌을 피하기 위해 조직 식별자가 필요하면:

```text
<org>-<system>-<environment>-<purpose>
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- purpose: `assets`, `uploads`, `logs`, `backups`, `artifacts`
- 이름은 3자 이상 63자 이하로 제한한다.
- 소문자, 숫자, 하이픈만 사용하고 문자나 숫자로 시작하고 끝낸다.
- 이름은 AWS partition 전체에서 중복할 수 없다.
- 마침표는 허용되지만 HTTPS 호환성을 위해 사용하지 않는다.
- IP 주소 형식과 AWS 예약 prefix 및 suffix는 사용하지 않는다.
- 이름과 Region은 생성 후 변경할 수 없다.
- 이름이 URL에 노출되므로 민감한 정보는 넣지 않는다.
- 같은 용도의 Bucket을 여러 Region에 만들 때만 Region 코드를 추가한다.
- S3 문맥에서 이미 리소스 종류가 드러나므로 `s3`, `bucket` suffix를 넣지 않는다.

예시:

```text
ninework-production-assets
ninework-production-uploads
ninework-production-logs
ninework-production-backups
payments-staging-artifacts
```

전역 이름 충돌이 있을 때만:

```text
acme-ninework-production-assets
```

## 2. RDS DB Instance 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<component>-rds
```

같은 역할의 Instance를 구분해야 하면:

```text
<system>-<environment>-<component>-<index>-rds
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- component: `app`, `core`, `reporting`, `analytics`, `legacy`
- index: 선택 항목. `01`, `02`
- 이름은 1자 이상 63자 이하로 제한한다.
- 영문자로 시작하고 소문자, 숫자, 하이픈만 사용한다.
- 하이픈으로 끝내거나 연속된 하이픈을 사용할 수 없다.
- 이름은 같은 AWS 계정과 Region 안에서 중복할 수 없다.
- RDS가 아닌 리소스와 구분할 수 있도록 `rds` suffix를 사용한다.
- Engine, Engine version, Instance class, 용량, AZ는 이름에 넣지 않는다.
- 장애조치로 역할이 바뀔 수 있으므로 `primary`, `writer`, `reader`, `replica`는 이름에 넣지 않는다.

예시:

```text
ninework-production-app-rds
ninework-production-reporting-rds
ninework-production-app-01-rds
ninework-production-app-02-rds
payments-staging-core-rds
```

## 3. RDS DB Cluster 이름 짓는 법

기본 템플릿:

```text
<system>-<environment>-<component>-rds-cluster
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- component: `app`, `core`, `reporting`, `analytics`
- Aurora DB Cluster와 Multi-AZ DB Cluster에 공통으로 적용한다.
- 두 Cluster 유형의 제한을 모두 만족하도록 이름은 최대 52자로 제한한다.
- 영문자로 시작하고 소문자, 숫자, 하이픈만 사용한다.
- 하이픈으로 끝내거나 연속된 하이픈을 사용할 수 없다.
- `rds`는 다른 종류의 Cluster와 구분하고 `cluster`는 RDS DB Instance와 구분하므로 둘 다 사용한다.
- Engine, Engine version, Instance class, AZ는 이름에 넣지 않는다.
- 장애조치로 역할이 바뀔 수 있으므로 `primary`, `writer`, `reader`는 이름에 넣지 않는다.

예시:

```text
ninework-production-app-rds-cluster
ninework-production-reporting-rds-cluster
payments-staging-core-rds-cluster
```
