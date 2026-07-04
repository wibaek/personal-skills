## Ruff (formatter + linter)

`assets/pyproject.toml.template`와 같은 Ruff config를 사용한다.

```toml
[tool.ruff]
line-length = 88
target-version = "py312"

[tool.ruff.lint]
select = [
  "F",              # Pyflakes
  "E4", "E7", "E9", # pycodestyle errors(Black과 호환되는 subset)
  # 아래 항목은 Ruff default rule selection에 포함되지 않는다.
  "I",              # isort
]
ignore = []

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
```

## ty (type checker)

type checking이 필요하면 이 section을 추가한다. Ruff config는 그대로 유지한다.

```toml
[tool.ty]
pythonVersion = "3.12"
typeCheckingMode = "basic"

reportMissingImports = true
reportMissingTypeStubs = false

reportUnusedImport = true
reportUnusedClass = true
reportUnusedFunction = true
reportUnusedVariable = true
```

## uv (PEP 621 project example)

이 예시는 다음을 보여준다.
- uv-style dependencies (`[project]`, `[project.optional-dependencies]`)
- Ruff config(template과 동일)
- Optional ty config

```toml
[project]
name = "project-name"
version = "0.1.0"
description = ""
readme = "README.md"
requires-python = ">=3.12"
dependencies = []

[project.optional-dependencies]
dev = [
  "ruff>=0.14.14",
  "ty>=0.0.14",
]

# Ruff (`assets/pyproject.toml.template` 기준)
[tool.ruff]
line-length = 88
target-version = "py312"

[tool.ruff.lint]
select = [
  "F",              # Pyflakes
  "E4", "E7", "E9", # pycodestyle errors(Black과 호환되는 subset)
  # 아래 항목은 Ruff default rule selection에 포함되지 않는다.
  "I",              # isort
]
ignore = []

[tool.ruff.format]
quote-style = "double"
indent-style = "space"

# ty(optional)
[tool.ty]
pythonVersion = "3.12"
typeCheckingMode = "basic"
reportMissingImports = true
reportMissingTypeStubs = false
reportUnusedImport = true
reportUnusedClass = true
reportUnusedFunction = true
reportUnusedVariable = true
```

## Poetry (project example)

이 예시는 다음을 보여준다.
- Poetry dependency blocks (`[tool.poetry.*]`)
- Ruff config(template과 동일)
- Optional ty config

```toml
[tool.poetry]
name = "project-name"
version = "0.1.0"
description = ""
authors = ["Your Name <you@example.com>"]

[tool.poetry.dependencies]
python = "^3.12"

[tool.poetry.group.dev.dependencies]
ruff = "^0.14.14"
ty = "^0.0.14"

# Ruff (`assets/pyproject.toml.template` 기준)
[tool.ruff]
line-length = 88
target-version = "py312"

[tool.ruff.lint]
select = [
  "F",              # Pyflakes
  "E4", "E7", "E9", # pycodestyle errors(Black과 호환되는 subset)
  # 아래 항목은 Ruff default rule selection에 포함되지 않는다.
  "I",              # isort
]
ignore = []

[tool.ruff.format]
quote-style = "double"
indent-style = "space"

# ty(optional)
[tool.ty]
pythonVersion = "3.12"
typeCheckingMode = "basic"
reportMissingImports = true
reportMissingTypeStubs = false
reportUnusedImport = true
reportUnusedClass = true
reportUnusedFunction = true
reportUnusedVariable = true
```
