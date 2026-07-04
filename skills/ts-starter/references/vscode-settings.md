# VSCode Workspace Settings 가이드

사용자가 VSCode integration을 원할 때만 이 설정을 제안한다.

## `.vscode/settings.json` 예시

```json
{
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "biomejs.biome",
  "editor.codeActionsOnSave": {
    "quickfix.biome": "explicit",
    "source.organizeImports.biome": "explicit"
  },
  "[typescript]": {
    "editor.defaultFormatter": "biomejs.biome"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "biomejs.biome"
  },
  "typescript.tsdk": "node_modules/typescript/lib"
}
```

## 추천 VSCode extensions

1. **Biome** (`biomejs.biome`)
2. **TypeScript and JavaScript Language Features** (built-in)
3. **Vitest** (`vitest.explorer`) - optional

## 파일 위치

project root에 `.vscode/settings.json`을 생성한다.
