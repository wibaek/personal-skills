---
name: ts-starter
description: 기존 Node 또는 frontend project에 app scaffolding 없이 TypeScript project tooling을 구성한다. TS starter setup, TS lint/format/test setup, Node TypeScript starter, Vite TypeScript config 요청에 사용한다.
---

# TypeScript Starter

TypeScript project의 linting, formatting, testing, compiler default를 자동으로 구성한다.

## 워크플로우

1. **package manager 감지**
   - lockfile이 없으면 `pnpm`을 우선한다.
   - lockfile이 있으면 기존 manager를 유지한다.
   - 감지 순서:
     1. `pnpm-lock.yaml` -> `pnpm`
     2. `package-lock.json` -> `npm`
     3. `yarn.lock` -> `yarn`
     4. `bun.lock` or `bun.lockb` -> `bun`
   - 감지가 모호할 때만 사용자에게 묻는다.

2. **project kind 감지**
   - `frontend`, `node-app`, `library` 중 하나로 분류한다.
   - 기존 파일(`vite.config.*`, framework deps, publish fields, entry points)을 감지에 사용한다.
   - project kind가 불분명하면 사용자에게 묻는다.

3. **baseline dependencies 설치**
   - `typescript`, `@types/node`, `@biomejs/biome`, `vitest`를 dev dependencies로 설치한다.
   - [references/environment-setup.md](references/environment-setup.md)의 package-manager-specific command를 사용한다.

4. **TypeScript configuration 적용**
   - strict-and-safety default로 `tsconfig.json`을 생성하거나 업데이트한다.
   - ESM default를 유지한다.
   - 가능하면 기존 option을 보존하고, 필요한 누락 field만 patch한다.
   - 참고: [references/config-examples.md](references/config-examples.md)
   - 템플릿: [assets/tsconfig.json.template](assets/tsconfig.json.template)

5. **Biome configuration 적용**
   - `biome.json`을 생성하거나 업데이트한다.
   - `package.json`에 `lint`, `lint:fix`, `format` script가 없으면 추가한다.
   - 기존 script는 유지하고 안전하게 merge한다.
   - 참고: [references/config-examples.md](references/config-examples.md)
   - 템플릿: [assets/biome.json.template](assets/biome.json.template)

6. **Vitest configuration 적용**
   - `vitest.config.ts`와 test script를 생성하거나 업데이트한다.
   - 이 skill version에서는 모든 project kind에 Vitest를 기본 runner로 사용한다.
   - 참고: [references/config-examples.md](references/config-examples.md)
   - 템플릿: [assets/vitest.config.ts.template](assets/vitest.config.ts.template)

7. **VSCode settings(Optional)**
   - `.vscode/settings.json`을 추가할지 묻는다.
   - 참고: [references/vscode-settings.md](references/vscode-settings.md)
   - 템플릿: [assets/.vscode-settings.json.template](assets/.vscode-settings.json.template)

8. **Git hooks(Optional)**
   - Husky + lint-staged 설정 여부를 묻는다.
   - lint와 typecheck용 pre-commit을 구성한다. test는 optional이다.
   - 참고: [references/git-hooks-setup.md](references/git-hooks-setup.md)
   - 템플릿: [assets/.husky-pre-commit.template](assets/.husky-pre-commit.template)

9. **ignore rule 업데이트**
   - 일반적인 TypeScript/Node artifact용 `.gitignore`를 생성하거나 patch한다.
   - 템플릿: [assets/.gitignore.template](assets/.gitignore.template)

10. **smoke validation 실행**
   - setup 후 `typecheck`, `lint`, `test`를 한 번 실행한다.
   - command는 package manager에 맞춰 유지한다.

## 원칙

- **Config-only**: 기본적으로 framework application을 scaffold하지 않는다.
- **Command-first**: 가능하면 manual editing보다 command를 우선한다.
- **Idempotent updates**: setup을 다시 실행해도 script를 중복하거나 config를 망가뜨리면 안 된다.
- **Preserve existing intent**: 현재 project configuration과 merge하고 필요한 부분만 patch한다.

## 내부 contract

- `project_kind`: `frontend | node-app | library`
- `package_manager`: `pnpm | npm | yarn | bun`

## 참고 자료

- **환경 설정**: [references/environment-setup.md](references/environment-setup.md)
- **Configuration 예시**: [references/config-examples.md](references/config-examples.md)
- **Git hooks 설정**: [references/git-hooks-setup.md](references/git-hooks-setup.md)
- **VSCode settings**: [references/vscode-settings.md](references/vscode-settings.md)
