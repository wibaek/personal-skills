# Git Hooks 설정 가이드

사용자가 명시적으로 동의한 경우에만 이 설정을 사용한다.

## Dependencies 설치

### pnpm

```bash
pnpm add -D husky lint-staged
pnpm husky init
```

### npm

```bash
npm install -D husky lint-staged
npx husky init
```

### yarn

```bash
yarn add -D husky lint-staged
yarn husky init
```

### bun

```bash
bun add -d husky lint-staged
bunx husky init
```

## `package.json` 예시

이 section을 추가하거나 merge한다.

```json
{
  "lint-staged": {
    "*.{ts,tsx,js,jsx,mjs,cjs,json}": ["biome check --write --no-errors-on-unmatched"]
  }
}
```

## Pre-commit Hook Pattern

pre-commit에서 lint-staged와 typecheck를 실행한다.

```sh
<package-manager> lint-staged && <package-manager> typecheck
```

더 엄격한 project에는 test를 추가할 수 있다.

```sh
<package-manager> lint-staged && <package-manager> typecheck && <package-manager> test
```

package manager별 실행 형식을 사용한다.

- pnpm: `pnpm <script>`
- npm: `npm run <script>`
- yarn: `yarn <script>`
- bun: `bun run <script>`
