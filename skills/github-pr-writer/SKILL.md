---
name: github-pr-writer
description: branch history, diff, validation result를 바탕으로 GitHub pull request 제목과 구조화된 PR body를 작성한다. PR 설명 작성, branch 요약, base branch 선택, `gh pr create` 준비, repository template을 따르는 PR 생성이 필요할 때 사용한다.
---

# GitHub PR Writer

정확한 제목, 구조화된 body, 검증 메모를 갖춘 review 가능한 pull request로 branch를 요약한다. 추측이 아니라 git 근거를 바탕으로 작성한다.

## 워크플로우

1. git에서 PR context를 구성한다.
   - `git status --short`로 uncommitted work를 확인한다.
   - 현재 branch를 식별한다.
   - repository convention, 기존 tracking branch, 또는 명시적인 사용자 지시에 따라 comparison target을 결정한다.
   - 작성 전에 commit history와 diff summary로 branch delta를 검토한다.

2. 정확성을 보호한다.
   - 가능하면 committed changes를 기준으로 PR을 작성한다.
   - 중요한 작업이 아직 uncommitted라면 PR 설명이 불완전할 수 있다고 말한다.
   - branch가 뒷받침하지 않는 동기, 구현 세부사항, validation result를 지어내지 않는다.

3. 간결한 PR title을 작성한다.
   - 확인 가능한 repository convention을 따른다.
   - convention이 없으면 사용자에게 보이는 결과 또는 reviewer에게 중요한 결과를 설명하는 짧고 구체적인 제목을 쓴다.
   - repository가 이미 쓰는 경우가 아니라면 불필요한 prefix를 피한다.

4. PR body는 다음 구조로 채운다.

```md
## 요약
- ...
- ...

## 변경 사항
- ...

## 검증
- ...

## 노트
- ...
```

5. 섹션 규칙을 적용한다.
   - `요약`: PR이 필요한 이유와 변경된 내용을 2-4개 bullet로 설명한다.
   - `변경 사항`: reviewer가 확인해야 할 구체적인 구현 변경을 나열한다.
   - `검증`: 실행한 command, 수행한 test, 완료한 manual check를 포함한다.
   - `노트`: rollout concern, follow-up, review hint, screenshot을 추가하거나, 유용한 내용이 없으면 `- 없음`을 쓴다.

6. reviewer에게 가치 있는 정보만 남긴다.
   - 외부에서 보이는 동작, migration, risk, 명확하지 않은 design choice를 강조한다.
   - 신호가 약한 편집 목록은 더 넓은 bullet로 묶는다.
   - 수정하지 않았지만 관련 있는 영역은 review 또는 rollout에 영향을 줄 때만 언급한다.

7. 요청받았을 때만 PR을 생성한다.
   - 사용자가 실제 PR 생성을 원하면 먼저 최종 title과 body를 준비한다.
   - `gh pr create --title ... --body-file ...` 같은 non-interactive GitHub CLI 사용을 우선한다.
   - authentication, base branch, remote target이 불명확하면 가정을 명시하지 않은 채 PR을 열거나 수정하지 않는다.

## 근거 수집

- commit 의도를 파악할 때는 `git log --oneline <base>..HEAD`를 우선한다.
- scope를 파악할 때는 `git diff --stat <base>...HEAD`와 필요한 diff 직접 확인을 우선한다.
- `검증`에는 현재 session에서 실행한 command 또는 최근 test output을 사용한다.
- validation을 실행하지 않았다면 `검증`에 명시적으로 적는다.

## 출력 규칙

- 사용자가 draft만 요청하면 title과 완성된 PR body를 Markdown으로 제공한다.
- 사용자가 PR 생성을 요청하면 사용할 title/body를 먼저 보여준 뒤 create command를 실행한다.
- branch에 관련 없는 작업이 섞여 있다면 PR을 열기 전에 이를 지적하고 분리를 권한다.

## 품질 기준

- 모든 bullet이 reviewer에게 유용해야 한다.
- body는 실제 branch delta와 맞아야 한다.
- `update logic`, `improve stuff` 같은 일반적 표현보다 구체적인 명사와 동사를 우선한다.
- 요약한 설명이 더 명확하다면 raw commit history를 그대로 붙여 넣지 않는다.
