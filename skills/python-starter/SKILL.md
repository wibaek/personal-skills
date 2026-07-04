---
name: python-starter
description: Python project를 처음 시작하거나 formatting, linting, testing 등 기본 개발 도구를 설정할 때 사용한다.
---

# Python Starter

Python project를 초기화할 때 formatter와 linter를 자동으로 구성한다.

## 워크플로우

1. **환경 감지와 선택**
   - `uv.lock` 파일이 있거나 `uv` command를 사용할 수 있으면 uv를 사용한다.
   - `pyproject.toml`에 `[tool.poetry]` section이 있으면 Poetry를 사용한다.
   - 환경이 설정되어 있지 않으면 uv를 추천한다.
   - 그 외에는 venv를 사용한다.
   - 확실하지 않으면 사용자에게 묻는다.

2. **ruff 설치**
   - 선택한 환경에 맞춰 ruff를 dev dependency로 설치한다.
   - 직접 코드를 쓰기보다 command를 사용한다.
   - 참고: [environment-setup.md](references/environment-setup.md)

3. **pyproject.toml 설정**
   - `pyproject.toml`이 없으면 생성한다.
   - ruff 설정을 추가한다. 기본 설정은 [pyproject-toml-examples.md](references/pyproject-toml-examples.md)를 참고한다.
   - 템플릿: [assets/pyproject.toml.template](assets/pyproject.toml.template)

4. **pytest 설치**
   - 선택한 환경에 맞춰 pytest를 dev dependency로 설치한다.
   - 직접 코드를 쓰기보다 command를 사용한다.

5. **Type Checking 설정(Optional)**
   - ty 설정이 필요한지 사용자에게 묻는다.
   - 필요하면 ty를 dev dependency로 설치한다.
   - `pyproject.toml`에 ty 설정을 추가한다.
   - 참고: [pyproject-toml-examples.md](references/pyproject-toml-examples.md)

6. **Pre-commit 설정(Optional)**
   - pre-commit 설정이 필요한지 사용자에게 묻는다.
   - 필요하면 pre-commit을 설치하고 `.pre-commit-config.yaml`을 생성한다.
   - 참고: [pre-commit-setup.md](references/pre-commit-setup.md)
   - 템플릿: [assets/.pre-commit-config.yaml.template](assets/.pre-commit-config.yaml.template)

7. **VSCode Settings 제안**
   - VSCode를 사용한다면 `.vscode/settings.json` 생성을 제안한다.
   - ruff와, 설정된 경우 ty 설정을 포함한다.
   - 참고: [vscode-settings.md](references/vscode-settings.md)
   - 템플릿: [assets/.vscode-settings.json.template](assets/.vscode-settings.json.template)

8. **.gitignore 생성**
   - `.gitignore`가 없으면 생성한다.
   - virtual environment, cache file, build artifact, IDE file, OS file에 대한 일반적인 Python ignore를 포함한다.
   - 템플릿: [assets/.gitignore.template](assets/.gitignore.template)

## 원칙

- **Command-first**: 가능하면 직접 코드를 쓰기보다 command를 사용한다.
- **pyproject.toml 사용**: 모든 설정은 `pyproject.toml`에 저장한다.
- **Dev dependencies**: ruff, pytest, ty, pre-commit은 dev dependencies로 설치한다.
- **User confirmation**: 환경 선택, ty, pre-commit은 진행 전에 사용자에게 확인한다.

## 참고 자료

- **환경 설정**: [references/environment-setup.md](references/environment-setup.md)
- **pyproject.toml examples**: [references/pyproject-toml-examples.md](references/pyproject-toml-examples.md)
- **Pre-commit 설정**: [references/pre-commit-setup.md](references/pre-commit-setup.md)
- **VSCode settings**: [references/vscode-settings.md](references/vscode-settings.md)
