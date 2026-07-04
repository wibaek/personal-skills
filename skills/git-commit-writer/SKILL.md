---
name: git-commit-writer
description: Conventional Commit 메시지를 작성하고 staged 또는 changed 파일에서 안전한 git commit workflow를 실행한다. commit 생성, commit message 제안, commit 전 git diff 확인이 필요할 때 사용한다.
---

# Git Commit Writer

실제 git 상태를 기준으로 명확한 Conventional Commit 메시지를 작성한다. 안전한 staging, 정확한 scope, 최소한의 commit body를 우선한다.

## 워크플로우

1. 메시지를 제안하기 전에 repository 상태를 확인한다.
   - `git status --short`를 확인한다.
   - 먼저 `git diff --cached`로 staged diff를 검토한다.
   - staged 항목이 없다면 `git diff`를 확인하고 staging이 필요한지 판단한다.
   - 수정된 모든 파일이 하나의 commit에 속한다고 가정하지 않는다.

2. 관련 없는 작업을 보호한다.
   - 요청된 작업과 무관한 사용자 변경은 commit에 포함하지 않는다.
   - worktree에 여러 관심사가 섞여 있다면 commit 분리를 권한다.
   - 요청된 commit scope가 모호하다면 commit 전에 가정을 명시한다.

3. 파일명만 보지 말고 동작 기준으로 commit 의도를 추론한다.
   - `feat`: 보이는 기능 또는 동작 기능 추가
   - `fix`: bug 또는 regression 수정
   - `refactor`: 동작 변경 없는 내부 구조 개선
   - `docs`: 문서 업데이트
   - `test`: test 추가 또는 조정
   - `chore`: tooling, config, dependencies, CI, maintenance
   - `style`: formatting 또는 비기능적 스타일 정리
   - `design`: frontend design, styling, visual UX 변경
   - `perf`: performance 개선

4. 명확성이 좋아질 때만 scope를 선택한다.
   - `auth`, `payment`, `readme`, `ci`처럼 package, feature, domain, subsystem 이름을 우선한다.
   - 변경이 작거나 repository 전체에 걸쳐 있으면 scope를 생략한다.

5. commit message는 다음 형식으로 작성한다.
   - `<type>: <subject>`
   - `<type>(<scope>): <subject>`
   - subject는 간결하고 구체적으로 작성한다.
   - commit log에서 자연스럽게 읽히는 현재형 한국어 표현을 사용한다.
   - 마침표로 끝내지 말고, `update`, `fix bug`, `changes`처럼 모호한 subject를 피한다.

6. 실제로 도움이 될 때만 body를 추가한다.
   - 명확하지 않은 reasoning, tradeoff, migration, multi-step fix에는 body를 사용한다.
   - body는 짧게 유지하고 변경이 왜 중요한지 설명한다.
   - subject와 body는 빈 줄로 분리한다.

7. 가능하면 commit 전에 검증한다.
   - 건드린 영역과 가장 관련 있는 test 또는 lint check를 실행한다.
   - 검증하지 않았다면 확신을 암시하지 말고 명시적으로 말한다.

8. 사용자가 commit 생성을 원하면 non-interactive 방식으로 commit한다.
   - `git commit -m "<subject>"`를 우선한다.
   - body가 필요하면 editor를 열지 말고 여러 `-m` flag를 사용한다.
   - 사용자가 명시적으로 요청하지 않는 한 amend 또는 history rewrite를 하지 않는다.

## 출력 규칙

- 사용자가 제안만 요청하면 다음을 제공한다.
  - 추천 commit message
  - diff와 연결된 짧은 근거
  - 실제 모호성이 있을 때만 대안
- 사용자가 commit을 요청하면 사용할 정확한 메시지를 보여준 뒤 commit을 실행한다.
- diff에 독립적인 변경이 여러 개 있다면 진행 전에 split commit을 권한다.

## 메시지 품질 기준

- 구현 세부사항보다 동작 중심 subject를 우선한다.
- 이후 history를 훑는 데 도움이 되면 영향을 받은 domain을 언급한다.
- 메시지는 실제 diff와 staged files에 맞춘다.
- diff가 뒷받침하지 않는 fix, feature, refactor를 주장하지 않는다.

## 예시

- `feat(auth): 로그인 만료 토큰 재발급 API 추가`
- `fix(payment): 결제 승인 실패 시 중복 요청 방지`
- `refactor(user): 회원 조회 서비스 책임 분리`
- `docs(readme): 로컬 실행 방법 보완`
- `test(order): 주문 취소 케이스 테스트 추가`
- `chore(ci): PR 검사 워크플로우 정리`

필요할 때는 body를 사용한다.

```text
fix(auth): refresh token 만료 검증 오류 수정

기존에는 access token 기준으로 만료를 체크하고 있어
refresh 요청이 정상적으로 처리되지 않는 문제가 있었다.
refresh token 기준으로 검증하도록 수정했다.
```
