---
name: systematic-debugging
description: 어떤 버그, 테스트 실패, 예상치 못한 동작을 만나든 수정안을 제안하기 전에 사용
---

# 체계적 디버깅

## 개요

무작위 수정은 시간을 낭비하고 새 버그를 만든다. 빠른 patch는 근본 문제를 가린다.

**핵심 원칙:** 수정을 시도하기 전에 항상 root cause를 찾는다. 증상 수정은 실패다.

**이 프로세스의 글자를 어기는 것은 디버깅의 정신을 어기는 것이다.**

## 철칙

```text
root cause 조사 없이는 수정하지 않는다
```

Phase 1을 완료하지 않았다면 수정안을 제안할 수 없다.

## 언제 사용할지

모든 기술 문제에 사용한다:

- 테스트 실패
- production bug
- 예상치 못한 동작
- 성능 문제
- 빌드 실패
- 통합 문제

**특히 다음 상황에서 사용한다:**

- "빠른 수정 하나"가 명백해 보일 때
- 이미 여러 수정안을 시도했을 때
- 이전 수정이 작동하지 않았을 때
- 문제를 완전히 이해하지 못했을 때

## 네 단계

다음 단계로 진행하기 전에 각 phase를 완료해야 한다.

### Phase 1: Root Cause 조사

**어떤 수정도 시도하기 전에:**

1. **오류 메시지를 주의 깊게 읽기**
   - 오류나 warning을 건너뛰지 않는다.
   - 정확한 해결책이 들어 있는 경우가 많다.
   - stack trace를 끝까지 읽는다.
   - line number, file path, error code를 기록한다.

2. **일관되게 재현하기**
   - 안정적으로 trigger할 수 있는가?
   - 정확한 단계는 무엇인가?
   - 매번 발생하는가?
   - 재현되지 않으면 더 많은 데이터를 모은다. 추측하지 않는다.

3. **최근 변경 확인**
   - 무엇이 바뀌어서 이 문제가 생길 수 있는가?
   - git diff, 최근 commit
   - 새 dependency, config 변경
   - 환경 차이

4. **여러 component 시스템에서 증거 수집**

   **시스템에 여러 component가 있으면(CI -> build -> signing, API -> service -> database):**

   **수정안을 제안하기 전에 diagnostic instrumentation을 추가한다:**

   ```text
   각 component boundary마다:
     - component에 들어오는 데이터 log
     - component에서 나가는 데이터 log
     - environment/config propagation 확인
     - 각 layer의 state 확인

   한 번 실행해 어디서 깨지는지 보여주는 증거를 모은다.
   그 다음 증거를 분석해 실패 component를 식별한다.
   그 다음 해당 component를 조사한다.
   ```

   **예시(multi-layer system):**

   ```bash
   # Layer 1: Workflow
   echo "=== Secrets available in workflow: ==="
   echo "IDENTITY: ${IDENTITY:+SET}${IDENTITY:-UNSET}"

   # Layer 2: Build script
   echo "=== Env vars in build script: ==="
   env | grep IDENTITY || echo "IDENTITY not in environment"

   # Layer 3: Signing script
   echo "=== Keychain state: ==="
   security list-keychains
   security find-identity -v

   # Layer 4: Actual signing
   codesign --sign "$IDENTITY" --verbose=4 "$APP"
   ```

   **이것이 드러내는 것:** 어떤 layer가 실패하는지(secrets -> workflow ✓, workflow -> build ✗)

5. **Data flow 추적**

   **오류가 call stack 깊은 곳에 있을 때:**

   완전한 backward tracing 기법은 이 디렉터리의 `root-cause-tracing.md`를 본다.

   **빠른 버전:**
   - 잘못된 값은 어디서 시작되는가?
   - 무엇이 이 잘못된 값으로 호출했는가?
   - source를 찾을 때까지 계속 위로 추적한다.
   - 증상이 아니라 source에서 고친다.

### Phase 2: Pattern 분석

**수정 전에 pattern을 찾는다:**

1. **작동하는 예시 찾기**
   - 같은 코드베이스에서 비슷하게 작동하는 코드를 찾는다.
   - 깨진 것과 비슷하지만 작동하는 것은 무엇인가?

2. **Reference와 비교**
   - pattern을 구현 중이라면 reference implementation을 완전히 읽는다.
   - 훑지 않는다. 모든 line을 읽는다.
   - 적용하기 전에 pattern을 완전히 이해한다.

3. **차이 식별**
   - 작동하는 것과 깨진 것의 차이는 무엇인가?
   - 아무리 작아 보여도 모든 차이를 나열한다.
   - "그건 중요하지 않을 것"이라고 가정하지 않는다.

4. **Dependency 이해**
   - 어떤 다른 component가 필요한가?
   - 어떤 설정, config, environment가 필요한가?
   - 어떤 가정을 하는가?

### Phase 3: 가설과 테스트

**과학적 방법:**

1. **단일 가설 형성**
   - 명확히 말한다: "Y 때문에 X가 root cause라고 생각한다"
   - 적어 둔다.
   - 모호하지 않게 구체적으로 말한다.

2. **최소 테스트**
   - 가설을 테스트하는 가장 작은 변경을 한다.
   - 한 번에 변수 하나만.
   - 여러 것을 동시에 고치지 않는다.

3. **계속하기 전에 검증**
   - 작동했는가? Yes -> Phase 4
   - 작동하지 않았는가? 새 가설을 만든다.
   - 위에 수정을 더 쌓지 않는다.

4. **모를 때**
   - "X를 이해하지 못했다"라고 말한다.
   - 아는 척하지 않는다.
   - 도움을 요청한다.
   - 더 조사한다.

### Phase 4: 구현

**증상이 아니라 root cause를 고친다:**

1. **실패 테스트 케이스 만들기**
   - 가능한 가장 단순한 재현
   - 가능하면 자동화 테스트
   - framework가 없으면 one-off test script
   - 수정 전에 반드시 있어야 함
   - 올바른 실패 테스트 작성에는 `superpowers:test-driven-development` 스킬 사용

2. **단일 수정 구현**
   - 식별한 root cause를 처리한다.
   - 한 번에 변경 하나.
   - "하는 김에" 개선하지 않는다.
   - 리팩터링을 함께 묶지 않는다.

3. **수정 검증**
   - 테스트가 이제 통과하는가?
   - 다른 테스트가 깨지지 않았는가?
   - 문제가 실제로 해결됐는가?

4. **수정이 작동하지 않으면**
   - 멈춘다.
   - 개수 확인: 몇 개의 수정을 시도했는가?
   - 3개 미만이면 Phase 1로 돌아가 새 정보로 다시 분석한다.
   - **3개 이상이면 멈추고 architecture를 질문한다(아래 5단계).**
   - architecture 논의 없이 Fix #4를 시도하지 않는다.

5. **3개 이상 수정 실패: Architecture 질문**

   **architecture 문제를 가리키는 pattern:**
   - 각 수정이 서로 다른 곳의 새 shared state/coupling/problem을 드러냄
   - 수정을 구현하려면 "massive refactoring"이 필요함
   - 각 수정이 다른 곳에서 새 증상을 만듦

   **멈추고 근본을 질문한다:**
   - 이 pattern은 근본적으로 타당한가?
   - 우리가 "관성만으로 계속 붙잡고 있는" 상태인가?
   - 증상 수정을 계속할지, architecture를 refactor할지?

   **추가 수정을 시도하기 전에 사람 파트너와 논의한다.**

   이것은 실패한 가설이 아니다. 잘못된 architecture다.

## 위험 신호 - 멈추고 프로세스를 따름

다음 생각이 들면:

- "지금은 quick fix하고 나중에 조사하자"
- "X를 바꿔보고 작동하는지 보자"
- "여러 변경을 추가하고 테스트 돌리자"
- "테스트는 건너뛰고 수동으로 검증하자"
- "아마 X일 테니 고치자"
- "완전히 이해하진 못했지만 이게 될 수도 있어"
- "pattern은 X라고 하는데 다르게 adapt하자"
- "주요 문제는 이것들입니다: [조사 없이 수정 목록]"
- data flow 추적 전에 solution 제안
- **"수정 한 번만 더" (이미 2개 이상 시도했을 때)**
- **각 수정이 다른 곳에서 새 문제를 드러냄**

**모두 멈추라는 뜻이다. Phase 1로 돌아간다.**

**3개 이상 수정이 실패했다면:** architecture를 질문한다(Phase 4.5 참고).

## 흔한 합리화

| 핑계                                            | 현실                                                                      |
| ----------------------------------------------- | ------------------------------------------------------------------------- |
| "문제가 단순해서 프로세스가 필요 없어"          | 단순한 문제에도 root cause가 있다. 단순 bug에는 프로세스가 빠르다.        |
| "긴급이라 프로세스 할 시간이 없어"              | 체계적 디버깅은 guess-and-check보다 빠르다.                               |
| "이것부터 해보고 그 다음 조사하자"              | 첫 수정이 pattern을 정한다. 처음부터 제대로 한다.                         |
| "수정이 되는지 확인한 뒤 test 쓰겠다"           | test 없는 수정은 유지되지 않는다. test first가 증명한다.                  |
| "여러 수정을 동시에 하면 시간이 절약된다"       | 무엇이 작동했는지 격리할 수 없다. 새 bug를 만든다.                        |
| "reference가 너무 길어서 pattern만 adapt하겠다" | 부분 이해는 bug를 보장한다. 완전히 읽는다.                                |
| "문제를 봤으니 고치겠다"                        | 증상을 보는 것은 root cause 이해가 아니다.                                |
| "한 번만 더 수정 시도"(2회 이상 실패 후)        | 3회 이상 실패는 architecture 문제다. pattern을 질문하고 더 고치지 않는다. |

## 빠른 참조

| Phase                 | 핵심 활동                              | 성공 기준             |
| --------------------- | -------------------------------------- | --------------------- |
| **1. Root Cause**     | error 읽기, 재현, 변경 확인, 증거 수집 | WHAT과 WHY 이해       |
| **2. Pattern**        | 작동 예시 찾기, 비교                   | 차이 식별             |
| **3. Hypothesis**     | 이론 만들기, 최소 테스트               | 확인됨 또는 새 가설   |
| **4. Implementation** | 테스트 작성, 수정, 검증                | bug 해결, 테스트 통과 |

## 프로세스가 "Root Cause 없음"을 드러낼 때

체계적 조사가 문제가 정말 환경적, timing-dependent, external이라고 드러내면:

1. 프로세스를 완료했다.
2. 무엇을 조사했는지 문서화한다.
3. 적절한 handling을 구현한다(retry, timeout, error message).
4. 향후 조사를 위한 monitoring/logging을 추가한다.

**하지만:** "root cause 없음" 사례의 95%는 불완전한 조사다.

## 지원 기법

다음 기법은 체계적 디버깅의 일부이며 이 디렉터리에 있다:

- **`root-cause-tracing.md`** - call stack을 거슬러 bug의 최초 trigger를 찾는다.
- **`defense-in-depth.md`** - root cause를 찾은 뒤 여러 layer에 validation을 추가한다.
- **`condition-based-waiting.md`** - 임의 timeout을 condition polling으로 교체한다.

**관련 스킬:**

- **superpowers:test-driven-development** - 실패 테스트 케이스 작성(Phase 4, Step 1)
- **superpowers:verification-before-completion** - 성공을 주장하기 전에 수정이 작동했는지 검증

## 실제 영향

디버깅 세션 기준:

- 체계적 접근: 수정까지 15-30분
- 무작위 수정 접근: 2-3시간 허둥댐
- 첫 수정 성공률: 95% vs 40%
- 새 bug 유입: 거의 없음 vs 흔함
