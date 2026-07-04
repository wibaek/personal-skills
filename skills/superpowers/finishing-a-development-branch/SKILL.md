---
name: finishing-a-development-branch
description: 구현이 완료되고 모든 테스트가 통과한 뒤, merge, PR, 정리 중 어떻게 통합할지 결정해야 할 때 사용. 구조화된 선택지를 제시해 개발 작업 마무리를 안내한다.
---

# 개발 브랜치 마무리

## 개요

명확한 선택지를 제시하고 선택된 워크플로우를 처리해 개발 작업을 마무리한다.

**핵심 원칙:** 테스트 검증 -> 환경 감지 -> 선택지 제시 -> 선택 실행 -> 정리.

**시작할 때 알림:** "이 작업을 완료하기 위해 finishing-a-development-branch 스킬을 사용합니다."

## 프로세스

### 1단계: 테스트 검증

**선택지를 제시하기 전에 테스트가 통과하는지 검증한다:**

```bash
# 프로젝트 테스트 스위트 실행
pnpm test / cargo test / pytest / go test ./...
```

**테스트가 실패하면:**

```text
테스트 실패(<N>개 실패). 완료 전에 반드시 수정해야 함:

[실패 내용 표시]

테스트가 통과하기 전에는 merge/PR을 진행할 수 없음.
```

멈춘다. 2단계로 진행하지 않는다.

**테스트가 통과하면:** 2단계로 계속한다.

### 2단계: 환경 감지

**선택지를 제시하기 전에 작업공간 상태를 확인한다:**

```bash
GIT_DIR=$(cd "$(git rev-parse --git-dir)" 2>/dev/null && pwd -P)
GIT_COMMON=$(cd "$(git rev-parse --git-common-dir)" 2>/dev/null && pwd -P)
```

이 값으로 어떤 메뉴를 보여줄지와 정리 방식을 결정한다:

| 상태                                      | 메뉴                          | 정리                  |
| ----------------------------------------- | ----------------------------- | --------------------- |
| `GIT_DIR == GIT_COMMON`(일반 repo)        | 표준 4개 선택지               | 정리할 worktree 없음  |
| `GIT_DIR != GIT_COMMON`, 이름 있는 브랜치 | 표준 4개 선택지               | 출처 기반(6단계 참고) |
| `GIT_DIR != GIT_COMMON`, detached HEAD    | 축소된 3개 선택지(merge 없음) | 정리 없음(외부 관리)  |

### 3단계: 기준 브랜치 결정

```bash
# 흔한 기준 브랜치 시도
git merge-base HEAD main 2>/dev/null || git merge-base HEAD master 2>/dev/null
```

또는 묻는다: "이 브랜치가 main에서 분기된 것이 맞나요?"

### 4단계: 선택지 제시

**일반 repo와 이름 있는 브랜치 worktree에서는 정확히 다음 4개 선택지를 제시한다:**

```text
구현 완료. 어떻게 진행할까요?

1. 로컬에서 <base-branch>로 merge
2. Push 후 Pull Request 생성
3. 브랜치를 그대로 유지(나중에 직접 처리)
4. 이 작업 폐기

어떤 선택지를 고를까요?
```

**Detached HEAD에서는 정확히 다음 3개 선택지를 제시한다:**

```text
구현 완료. 현재 detached HEAD 상태입니다(외부에서 관리하는 작업공간).

1. 새 브랜치로 push 후 Pull Request 생성
2. 그대로 유지(나중에 직접 처리)
3. 이 작업 폐기

어떤 선택지를 고를까요?
```

**설명을 덧붙이지 않는다.** 선택지는 간결하게 유지한다.

### 5단계: 선택 실행

#### 선택지 1: 로컬 merge

```bash
# CWD 안전을 위해 main repo 루트 확인
MAIN_ROOT=$(git -C "$(git rev-parse --git-common-dir)/.." rev-parse --show-toplevel)
cd "$MAIN_ROOT"

# 먼저 merge하고 성공을 확인한 뒤 어떤 것도 제거한다.
git checkout <base-branch>
git pull
git merge <feature-branch>

# merge 결과에서 테스트 검증
<test command>

# merge가 성공한 뒤에만 worktree 정리(6단계), 그 다음 브랜치 삭제
```

그 다음 worktree 정리(6단계) 후 브랜치를 삭제한다:

```bash
git branch -d <feature-branch>
```

#### 선택지 2: Push 및 PR 생성

```bash
# 브랜치 push
git push -u origin <feature-branch>

# PR 생성
gh pr create --title "<title>" --body "$(cat <<'EOF'
## 요약
<변경 사항 2-3개 bullet>

## 검증
- [ ] <검증 단계>
EOF
)"
```

**worktree를 정리하지 않는다.** 사용자가 PR 피드백을 반복 반영하려면 살아 있어야 한다.

#### 선택지 3: 그대로 유지

보고: "브랜치 <name>을 그대로 둡니다. Worktree는 <path>에 보존했습니다."

**worktree를 정리하지 않는다.**

#### 선택지 4: 폐기

**먼저 확인한다:**

```text
다음 항목이 영구 삭제됩니다:
- 브랜치 <name>
- 모든 commit: <commit-list>
- <path>의 worktree

확인하려면 'discard'를 입력하세요.
```

정확한 확인 입력을 기다린다.

확인되면:

```bash
MAIN_ROOT=$(git -C "$(git rev-parse --git-common-dir)/.." rev-parse --show-toplevel)
cd "$MAIN_ROOT"
```

그 다음 worktree 정리(6단계) 후 브랜치를 강제 삭제한다:

```bash
git branch -D <feature-branch>
```

### 6단계: 작업공간 정리

**선택지 1과 4에서만 실행한다.** 선택지 2와 3은 항상 worktree를 보존한다.

```bash
GIT_DIR=$(cd "$(git rev-parse --git-dir)" 2>/dev/null && pwd -P)
GIT_COMMON=$(cd "$(git rev-parse --git-common-dir)" 2>/dev/null && pwd -P)
WORKTREE_PATH=$(git rev-parse --show-toplevel)
```

**`GIT_DIR == GIT_COMMON`이면:** 일반 repo다. 정리할 worktree가 없다. 완료.

**worktree 경로가 `.worktrees/`, `worktrees/` 아래라면:** Superpowers가 만든 worktree다. 정리 책임이 있다.

```bash
MAIN_ROOT=$(git -C "$(git rev-parse --git-common-dir)/.." rev-parse --show-toplevel)
cd "$MAIN_ROOT"
git worktree remove "$WORKTREE_PATH"
git worktree prune  # 자가 복구: 오래된 등록 정리
```

**그 외:** 호스트 환경(harness)이 이 작업공간을 소유한다. 제거하지 않는다. 플랫폼이 작업공간 종료 도구를 제공하면 사용한다. 없으면 작업공간을 그대로 둔다.

## 빠른 참조

| 선택지         | Merge | Push | Worktree 유지 | 브랜치 정리 |
| -------------- | ----- | ---- | ------------- | ----------- |
| 1. 로컬 merge  | 예    | -    | -             | 예          |
| 2. PR 생성     | -     | 예   | 예            | -           |
| 3. 그대로 유지 | -     | -    | 예            | -           |
| 4. 폐기        | -     | -    | -             | 예(force)   |

## 흔한 실수

**테스트 검증 건너뛰기**

- **문제:** 깨진 코드를 merge하거나 실패하는 PR을 만든다.
- **수정:** 선택지를 제시하기 전에 항상 테스트를 검증한다.

**열린 질문**

- **문제:** "다음에 뭘 할까요?"는 모호하다.
- **수정:** 정확히 4개의 구조화된 선택지(detached HEAD는 3개)를 제시한다.

**선택지 2에서 worktree 정리**

- **문제:** 사용자가 PR 반복 작업에 필요한 worktree를 제거한다.
- **수정:** 선택지 1과 4에서만 정리한다.

**worktree 제거 전에 브랜치 삭제**

- **문제:** worktree가 아직 브랜치를 참조해 `git branch -d`가 실패한다.
- **수정:** 먼저 merge하고, worktree를 제거한 다음, 브랜치를 삭제한다.

**worktree 안에서 `git worktree remove` 실행**

- **문제:** 제거 대상 worktree 내부가 CWD이면 명령이 조용히 실패할 수 있다.
- **수정:** `git worktree remove` 전에 항상 main repo 루트로 `cd`한다.

**harness가 소유한 worktree 정리**

- **문제:** harness가 만든 worktree를 제거하면 유령 상태가 생긴다.
- **수정:** `.worktrees/`, `worktrees/` 아래 worktree만 정리한다.

**폐기에 확인 없음**

- **문제:** 작업을 실수로 삭제한다.
- **수정:** 사용자가 직접 `"discard"`를 입력해 확인하도록 요구한다.

## 위험 신호

**절대 하지 말 것:**

- 실패한 테스트를 두고 진행하기
- 결과 테스트를 검증하지 않고 merge하기
- 확인 없이 작업 삭제하기
- 명시적 요청 없이 force-push하기
- merge 성공을 확인하기 전에 worktree 제거하기
- 내가 만들지 않은 worktree 정리하기(출처 확인)
- worktree 안에서 `git worktree remove` 실행하기

**항상 할 것:**

- 선택지를 제시하기 전에 테스트 검증
- 메뉴를 제시하기 전에 환경 감지
- 정확히 4개 선택지(detached HEAD는 3개) 제시
- 선택지 4에는 입력 확인 받기
- 선택지 1과 4에서만 worktree 정리
- worktree 제거 전에 main repo 루트로 `cd`
- 제거 후 `git worktree prune` 실행
