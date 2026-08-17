# Observability Naming Convention

## 1. CloudWatch Log Group 이름 짓는 법

AWS 서비스가 자동 생성하는 Log Group은 해당 서비스의 기본 이름을 유지한다.

예시:

```text
/aws/lambda/<lambda-name>
```

직접 만드는 Log Group의 기본 템플릿:

```text
/<system>/<environment>/<component>
```

로그 종류를 구분해야 하면:

```text
/<system>/<environment>/<component>/<log-type>
```

- system: `ninework`, `payments`, `auth`
- environment: `development`, `staging`, `production`, `test`
- component: `api`, `web`, `worker`
- log-type: `access`, `application`, `audit`
- AWS 서비스가 자동 생성하는 `/aws/...` 이름은 변경하지 않는다.
- 직접 만드는 Log Group에는 `/aws/` namespace를 사용하지 않는다.
- `/`를 계층 구분자로 사용하고 각 경로 조각은 `lowercase-kebab-case`로 작성한다.
- CloudWatch Logs 문맥에서 이미 리소스 종류가 드러나므로 `cloudwatch`, `log`, `log-group`을 붙이지 않는다.
- 이름은 1자 이상 512자 이하로 제한한다.
- 이름은 계정과 Region 안에서 중복할 수 없다.

예시:

```text
/aws/lambda/ninework-production-api-lambda
/ninework/production/api
/ninework/production/worker
/ninework/production/web/access
/payments/staging/api/audit
```
