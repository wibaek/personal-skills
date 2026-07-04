---
name: requesting-code-review
description: 작업 완료, 주요 기능 구현, merge 전 요구사항 충족 여부를 검증해야 할 때 사용
---

# 코드 리뷰 요청

문제가 커지기 전에 코드 리뷰어 서브에이전트를 파견해 잡아낸다. 리뷰어는 평가에 필요한 컨텍스트만 정확히 받으며, 현재 세션 기록을 받지 않는다. 이렇게 하면 리뷰어는 사고 과정이 아니라 작업 산출물에 집중하고, 자신의 컨텍스트는 이후 작업을 위해 보존된다.

**핵심 원칙:** 일찍 리뷰하고, 자주 리뷰한다.

## 리뷰를 요청할 때

**필수:**

- Subagent-Driven Development의 각 작업 후
- 주요 기능 완료 후
- main으로 merge하기 전

**선택이지만 유용함:**

- 막혔을 때(새 관점)
- 리팩터링 전(기준선 확인)
- 복잡한 버그 수정 후

## 요청 방법

**1. git SHA 가져오기:**

```bash
BASE_SHA=$(git rev-parse HEAD~1)  # 또는 origin/main
HEAD_SHA=$(git rev-parse HEAD)
```

**2. 코드 리뷰어 서브에이전트 파견:**

`general-purpose` 유형의 Task 도구를 사용하고, `code-reviewer.md` 템플릿을 채운다.

**플레이스홀더:**

- `{DESCRIPTION}` - 무엇을 만들었는지 짧은 요약
- `{PLAN_OR_REQUIREMENTS}` - 무엇을 해야 하는지
- `{BASE_SHA}` - 시작 commit
- `{HEAD_SHA}` - 종료 commit

**3. 피드백 처리:**

- Critical 이슈는 즉시 수정한다.
- Important 이슈는 진행 전에 수정한다.
- Minor 이슈는 나중을 위해 메모한다.
- 리뷰어가 틀렸다면 근거를 들어 반박한다.

## 예시

```text
[방금 Task 2: verification function 추가 완료]

You: 계속하기 전에 코드 리뷰를 요청하겠습니다.

BASE_SHA=$(git log --oneline | grep "Task 1" | head -1 | awk '{print $1}')
HEAD_SHA=$(git rev-parse HEAD)

[코드 리뷰어 서브에이전트 파견]
  DESCRIPTION: issue type 4개를 포함해 verifyIndex()와 repairIndex() 추가
  PLAN_OR_REQUIREMENTS: docs/superpowers/plans/deployment-plan.md의 Task 2
  BASE_SHA: a7981ec
  HEAD_SHA: 3df7661

[서브에이전트 반환]:
  강점: 깔끔한 architecture, 실제 테스트
  이슈:
    Important: progress indicator 누락
    Minor: reporting interval의 magic number (100)
  평가: 계속 진행 가능

You: [progress indicator 수정]
[Task 3 계속]
```

## 워크플로우 통합

**Subagent-Driven Development:**

- 각 작업 후 리뷰
- 문제가 누적되기 전에 포착
- 다음 작업으로 이동하기 전에 수정

**Executing Plans:**

- 각 작업 후 또는 자연스러운 체크포인트에서 리뷰
- 피드백을 받고 적용한 뒤 계속

**Ad-Hoc Development:**

- merge 전 리뷰
- 막혔을 때 리뷰

## 위험 신호

**절대 하지 말 것:**

- "간단하니까" 리뷰 건너뛰기
- Critical 이슈 무시
- Important 이슈를 고치지 않고 진행
- 타당한 기술 피드백과 논쟁

**리뷰어가 틀렸다면:**

- 기술적 근거로 반박한다.
- 작동을 증명하는 코드/테스트를 보여준다.
- 명확화를 요청한다.

템플릿 위치: `requesting-code-review/code-reviewer.md`
