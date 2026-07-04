# TypeScript 환경 설정 가이드

## Package Manager 감지 규칙

다음 순서로 확인하고 첫 match에서 멈춘다.

1. `pnpm-lock.yaml` -> `pnpm`
2. `package-lock.json` -> `npm`
3. `yarn.lock` -> `yarn`
4. `bun.lock` or `bun.lockb` -> `bun`
5. lockfile 없음 -> `pnpm`(기본값)

lockfile이 여러 개 있으면 dependencies를 변경하기 전에 어떤 manager가 기준인지 사용자에게 묻는다.

## Baseline dependencies

다음을 dev dependencies로 설치한다.

- `typescript`
- `@types/node`
- `@biomejs/biome`
- `vitest`

## Package manager commands

### pnpm

```bash
pnpm add -D typescript @types/node @biomejs/biome vitest
```

### npm

```bash
npm install -D typescript @types/node @biomejs/biome vitest
```

### yarn

```bash
yarn add -D typescript @types/node @biomejs/biome vitest
```

### bun

```bash
bun add -d typescript @types/node @biomejs/biome vitest
```

## 기존 Manager 존중 정책

- 기존 lockfile이 있으면 해당 package manager를 유지한다.
- lockfile이 없으면 `pnpm`으로 초기화하고 진행한다.
- lockfile ecosystem을 조용히 바꾸지 않는다.
