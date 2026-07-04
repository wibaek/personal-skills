# TypeScript Config 예시

## 표준 `package.json` Scripts

이 scripts를 baseline으로 사용한다. 기존 project-specific script는 제거하지 말고 누락된 script만 추가한다.

```json
{
  "typecheck": "tsc --noEmit",
  "build": "tsc -p tsconfig.json",
  "lint": "biome check .",
  "lint:fix": "biome check --write .",
  "format": "biome format --write .",
  "test": "vitest run",
  "test:watch": "vitest"
}
```

## Strict+Safety `tsconfig.json` Baseline

Node와 frontend project의 기본값으로 사용한다. ESM default를 유지한다.

```json
{
  "$schema": "https://json.schemastore.org/tsconfig",
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "lib": ["ES2023"],
    "types": ["node", "vitest/globals"],
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitOverride": true,
    "noFallthroughCasesInSwitch": true,
    "useUnknownInCatchVariables": true,
    "resolveJsonModule": true,
    "verbatimModuleSyntax": true,
    "skipLibCheck": true,
    "declaration": true,
    "sourceMap": true,
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src", "test", "vitest.config.ts"],
  "exclude": ["dist", "node_modules"]
}
```

frontend project에서는 필요한 경우에만 `jsx` 같은 framework-specific option을 추가한다.

## Biome `biome.json` 예시

```json
{
  "$schema": "https://biomejs.dev/schemas/latest/schema.json",
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true
    }
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "double",
      "semicolons": "always"
    }
  }
}
```

## Vitest `vitest.config.ts` 예시

```ts
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    globals: true,
    environment: "node",
    include: ["src/**/*.test.ts", "test/**/*.test.ts"]
  }
});
```
