# AWS CLI commands

값을 출력하지 말고 project에 맞는 non-secret 변수만 바꿔서 사용한다.

```bash
profile="personal"
region="ap-northeast-2"
parameter="/project-name/dev/env"
```

## Preflight

AWS identity를 확인한다.

```bash
aws sts get-caller-identity \
  --profile "$profile" \
  --region "$region" \
  --query '{Account:Account,Arn:Arn}'
```

`.env` 크기와 Git ignore 상태를 확인한다. 파일 내용은 출력하지 않는다.

```bash
wc -c < .env
git check-ignore -v .env
awk -F= '/^[A-Za-z_][A-Za-z0-9_]*=/{print $1}' .env
```

기존 parameter가 있으면 값 없이 type과 version만 확인한다.

```bash
aws ssm get-parameter \
  --profile "$profile" \
  --region "$region" \
  --name "$parameter" \
  --no-with-decryption \
  --query 'Parameter.{Type:Type,Version:Version}'
```

## Create

기존 parameter를 덮어쓰지 않고 `.env` 원문을 등록한다.

```bash
aws ssm put-parameter \
  --profile "$profile" \
  --region "$region" \
  --name "$parameter" \
  --type SecureString \
  --tier Standard \
  --value file://.env \
  --no-overwrite
```

## Update

기존 parameter 갱신을 사용자가 명시적으로 확인한 경우에만 실행한다.

```bash
aws ssm put-parameter \
  --profile "$profile" \
  --region "$region" \
  --name "$parameter" \
  --type SecureString \
  --tier Standard \
  --value file://.env \
  --overwrite
```

## Restore

기존 `.env`가 있으면 중단한다. AWS CLI가 실패해도 기존 파일이나 빈 `.env`를 남기지 않는다.

```bash
(
  set -eu
  if [ -e .env ] || [ -L .env ]; then
    echo ".env already exists" >&2
    exit 1
  fi

  umask 077
  tmp="$(mktemp ./.env.XXXXXX)"
  trap 'rm -f "$tmp"' EXIT HUP INT TERM

  aws ssm get-parameter \
    --profile "$profile" \
    --region "$region" \
    --name "$parameter" \
    --with-decryption \
    --query 'Parameter.Value' \
    --output text > "$tmp"

  chmod 600 "$tmp"
  mv "$tmp" .env
  trap - EXIT HUP INT TERM
)
```

확인 결과만 출력한다.

```bash
stat -f '%Sp %N' .env
wc -c < .env
git check-ignore -v .env
```

Linux에서는 file mode 확인에 `stat -c '%A %n' .env`를 사용한다.

## Git ignore

기존 규칙과 충돌하지 않는지 확인한 뒤 필요한 항목만 추가한다.

```gitignore
.env
.env.*
!.env.example
```

## Project 문서에 기록할 내용

- AWS account ID, profile, Region, parameter 경로
- AWS CLI 인증 방법
- 위 restore 명령
- `.env`를 Git에 commit하지 않는다는 원칙
- 기존 `.env`를 자동으로 덮어쓰지 않는다는 주의

SSO profile을 사용하면 인증 명령도 기록한다.

```bash
aws sso login --profile "$profile"
```
