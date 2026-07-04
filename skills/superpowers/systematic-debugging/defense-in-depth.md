# Defense-in-Depth 검증

## 개요

잘못된 데이터 때문에 생긴 bug를 고칠 때 한 곳에 validation을 추가하면 충분해 보인다. 하지만 단일 check는 다른 code path, refactoring, mock으로 우회될 수 있다.

**핵심 원칙:** 데이터가 통과하는 모든 layer에서 validate한다. bug가 구조적으로 불가능하게 만든다.

## 왜 여러 Layer인가

단일 validation: "버그를 고쳤다"
여러 layer: "버그가 불가능하게 만들었다"

서로 다른 layer는 서로 다른 case를 잡는다:

- Entry validation은 대부분의 bug를 잡는다.
- Business logic은 edge case를 잡는다.
- Environment guard는 context-specific danger를 막는다.
- Debug logging은 다른 layer가 실패할 때 도움이 된다.

## 네 Layer

### Layer 1: Entry Point Validation

**목적:** API boundary에서 명백히 잘못된 input 거부

```typescript
function createProject(name: string, workingDirectory: string) {
  if (!workingDirectory || workingDirectory.trim() === '') {
    throw new Error('workingDirectory cannot be empty');
  }
  if (!existsSync(workingDirectory)) {
    throw new Error(`workingDirectory does not exist: ${workingDirectory}`);
  }
  if (!statSync(workingDirectory).isDirectory()) {
    throw new Error(`workingDirectory is not a directory: ${workingDirectory}`);
  }
  // ... proceed
}
```

### Layer 2: Business Logic Validation

**목적:** 이 operation에 대해 data가 말이 되는지 보장

```typescript
function initializeWorkspace(projectDir: string, sessionId: string) {
  if (!projectDir) {
    throw new Error('projectDir required for workspace initialization');
  }
  // ... proceed
}
```

### Layer 3: Environment Guard

**목적:** 특정 context에서 위험한 operation 방지

```typescript
async function gitInit(directory: string) {
  // 테스트에서는 temp directory 밖 git init 거부
  if (process.env.NODE_ENV === 'test') {
    const normalized = normalize(resolve(directory));
    const tmpDir = normalize(resolve(tmpdir()));

    if (!normalized.startsWith(tmpDir)) {
      throw new Error(
        `Refusing git init outside temp dir during tests: ${directory}`
      );
    }
  }
  // ... proceed
}
```

### Layer 4: Debug Instrumentation

**목적:** forensic을 위한 context 캡처

```typescript
async function gitInit(directory: string) {
  const stack = new Error().stack;
  logger.debug('About to git init', {
    directory,
    cwd: process.cwd(),
    stack,
  });
  // ... proceed
}
```

## Pattern 적용

bug를 찾으면:

1. **data flow 추적** - 잘못된 값은 어디서 시작되는가? 어디서 사용되는가?
2. **모든 checkpoint 매핑** - data가 통과하는 모든 지점 나열
3. **각 layer에 validation 추가** - entry, business, environment, debug
4. **각 layer 테스트** - layer 1을 우회해 보고 layer 2가 잡는지 검증

## 세션 예시

Bug: 빈 `projectDir` 때문에 source code에서 `git init`이 실행됨

**Data flow:**

1. Test setup -> empty string
2. `Project.create(name, '')`
3. `WorkspaceManager.createWorkspace('')`
4. `git init` runs in `process.cwd()`

**추가된 네 layer:**

- Layer 1: `Project.create()`가 empty/exists/writable 검증
- Layer 2: `WorkspaceManager`가 projectDir empty 검증
- Layer 3: `WorktreeManager`가 test에서 tmpdir 밖 `git init` 거부
- Layer 4: `git init` 전 stack trace logging

**결과:** 1847개 test 모두 통과, bug 재현 불가능

## 핵심 통찰

네 layer가 모두 필요했다. 테스트 중 각 layer는 다른 layer가 놓친 bug를 잡았다:

- 다른 code path가 entry validation을 우회했다.
- mock이 business logic check를 우회했다.
- 다른 platform의 edge case에는 environment guard가 필요했다.
- debug logging이 구조적 오용을 식별했다.

**validation point 하나에서 멈추지 않는다.** 모든 layer에 check를 추가한다.
