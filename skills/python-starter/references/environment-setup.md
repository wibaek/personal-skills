# Python 환경 설정 가이드

## 환경 선택 기준

- **venv**: standard library 기반. 단순한 project에 적합하다.
- **uv**: 빠른 속도가 필요한 project에 적합하다.
- **poetry**: dependency management와 packaging이 필요한 project에 적합하다.

## venv 사용

```bash
# virtual environment 생성
python -m venv .venv

# virtual environment 활성화(macOS/Linux)
source .venv/bin/activate

# virtual environment 활성화(Windows)
.venv\Scripts\activate

# ruff 설치(dev)
pip install --upgrade ruff

# ty 설치(dev, optional)
pip install --upgrade ty
```

## uv 사용

```bash
# uv project 초기화
uv init

# 또는 기존 project에 uv 설정
uv pip install --system ruff ty

# 또는 pyproject.toml에 dev dependencies를 추가한 뒤
uv sync --dev
```

## Poetry 사용

```bash
# Poetry project 초기화(기존 directory)
poetry init --no-interaction

# 또는 pyproject.toml이 이미 있는 경우
poetry install

# ruff를 dev dependency로 추가
poetry add --group dev ruff

# ty를 dev dependency로 추가(optional)
poetry add --group dev ty

# virtual environment 활성화
poetry shell
```

## 환경 감지 방법

1. **uv 감지**: `uv.lock` 파일이 있거나 `uv` command를 사용할 수 있는지 확인한다.
2. **Poetry 감지**: `pyproject.toml`에 `[tool.poetry]` section이 있는지 확인한다.
3. **기본값**: 위 조건에 해당하지 않으면 venv를 사용한다.
