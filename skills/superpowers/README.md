# Superpowers Skills

`obra/superpowers`에서 `skills` 내용만 가져와 한국어로 옮긴 개인용 슈퍼파워 모음이다. 플러그인 메타데이터, 에이전트 설정, README, LICENSE, assets는 포함하지 않았다.

`using-superpowers`는 모든 대화에서 강제 호출되는 성격이 강해 기본 설치 대상에서 제외했다.

## 권장 사용 순서

일반적인 개발 작업에서는 아래 흐름을 기본 순서로 본다. 모든 스킬을 항상 쓰는 것은 아니고, 현재 상황에 맞는 단계만 선택한다.

1. `brainstorming` - 창의적 작업, 기능 설계, 동작 변경 전에 의도와 요구사항 정리
2. `writing-plans` - 여러 단계 작업을 코드 변경 전에 실행 가능한 계획으로 분해
3. `executing-plans` - 이미 작성된 구현 계획을 체크포인트 기반으로 실행
4. `subagent-driven-development` - 독립 작업이 여러 개 있는 구현 계획을 현재 세션에서 병렬적으로 진행
5. `dispatching-parallel-agents` - 공유 상태나 순차 의존성이 없는 2개 이상 작업을 병렬 처리
6. `test-driven-development` - 기능/버그 수정 구현 전에 실패 테스트부터 작성
7. `systematic-debugging` - 버그, 테스트 실패, 예상치 못한 동작의 원인을 추적
8. `requesting-code-review` - 작업 완료 후 요구사항 충족 여부를 리뷰 요청
9. `receiving-code-review` - 리뷰 피드백을 검토하고 구현 여부를 판단
10. `verification-before-completion` - 완료, 수정, 통과를 말하기 전에 검증 명령 실행
11. `finishing-a-development-branch` - 테스트 통과 후 merge, PR, 정리 방식 결정

## 분류별 색인

| 분류 | 스킬 | 용도 |
| --- | --- | --- |
| 설계 | `brainstorming` | 구현 전에 의도, 요구사항, 설계 선택지 정리 |
| 계획 | `writing-plans` | 큰 작업을 검증 가능한 작은 단계로 분해 |
| 계획 실행 | `executing-plans` | 작성된 계획을 체크포인트 단위로 실행 |
| 병렬 실행 | `subagent-driven-development` | 독립 구현 작업을 서브에이전트 흐름으로 진행 |
| 병렬 실행 | `dispatching-parallel-agents` | 독립 조사/수정/검증 작업을 동시에 처리 |
| 구현 | `test-driven-development` | 실패 테스트, 최소 구현, 리팩터 순서 유지 |
| 디버깅 | `systematic-debugging` | 원인 조사, 패턴 확인, 수정 전 근본 원인 확인 |
| 리뷰 | `requesting-code-review` | 완료 작업에 대한 코드 리뷰 요청 |
| 리뷰 | `receiving-code-review` | 리뷰 피드백 검토, 반박, 구현, 검증 |
| 검증 | `verification-before-completion` | 성공 주장 전 실제 명령 출력 확인 |
| 마무리 | `finishing-a-development-branch` | merge, PR 생성, worktree 정리 결정 |

## 참고 규칙

- 다른 슈퍼파워 스킬을 문서에서 언급할 때는 `superpowers:<skill-name>` 형식을 우선 사용한다.
- 파일 경로가 필요할 때는 `skills/superpowers/<skill-name>/...` 구조를 기준으로 적는다.
- 새 개인 스킬을 추가할 때는 이 README의 순서와 색인을 함께 갱신한다.
