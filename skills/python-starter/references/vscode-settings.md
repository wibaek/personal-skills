# VSCode Workspace Settings 가이드

## .vscode/settings.json 예시

### Ruff 설정만

```json
{
  "[python]": {
    "editor.defaultFormatter": "charliermarsh.ruff",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.fixAll": "explicit",
      "source.organizeImports": "explicit"
    }
  }
}
```

### ty type checking 설정만

```json
{
  "python.analysis.typeCheckingMode": "basic",
  "python.analysis.diagnosticMode": "workspace"
}
```

## 필요한 VSCode extensions

사용자에게 다음 extension을 제안한다.

1. **Ruff** (`charliermarsh.ruff`) - 필수
2. **Python** (`ms-python.python`) - 필수(ty와 함께 쓰는 Pylance 포함)

## Configuration file 생성

project root에 `.vscode/settings.json` 파일 생성을 제안한다.
