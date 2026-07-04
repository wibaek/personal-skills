# Pre-commit 설정 가이드

## Pre-commit 설치와 설정

```bash
# pre-commit 설치(dev dependency)
# venv 사용 시
pip install --upgrade pre-commit

# poetry 사용 시
poetry add --group dev pre-commit

# uv 사용 시
uv pip install --system pre-commit
# 또는 pyproject.toml에 추가한 뒤 uv sync --dev
```

## .pre-commit-config.yaml 예시

### 기본(ruff only)

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-json
      - id: check-toml
      - id: check-merge-conflict
      - id: debug-statements

  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix, --exit-non-zero-on-fix]
      - id: ruff-format
```

### Ruff + ty

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-json
      - id: check-toml
      - id: check-merge-conflict
      - id: debug-statements

  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix, --exit-non-zero-on-fix]
      - id: ruff-format

  - repo: https://github.com/RobertCraigie/ty-python
    rev: v1.1.0
    hooks:
      - id: ty
```

## Pre-commit 설치 command

configuration file을 만든 뒤:

```bash
# pre-commit hooks 설치
pre-commit install

# 모든 파일에 한 번 실행(optional)
pre-commit run --all-files
```
