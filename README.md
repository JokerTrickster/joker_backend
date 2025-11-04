# Claude 스킬 생성 가이드

## Claude 스킬이란?

Claude 스킬(Skills)은 특정 작업에 필요한 지침, 코드, 참고 자료 등을 하나의 폴더로 묶어, Claude가 자동으로 활용할 수 있게 해주는 기능입니다. 이를 통해 반복적인 프롬프트 입력 없이도 일관된 결과를 얻을 수 있습니다.

## 스킬 생성 방법

### 1. 기본 구조

```
my-skill/
├── SKILL.md          # 필수: 스킬 메타데이터 및 지침
├── references/       # 선택: 참고 자료
│   └── ...
├── scripts/          # 선택: 실행 가능한 코드
│   └── ...
└── examples/         # 선택: 예시 파일
    └── ...
```

### 2. SKILL.md 파일 작성

`SKILL.md` 파일은 반드시 포함해야 하며, 다음 형식으로 작성합니다:

```markdown
---
name: my-skill-name
description: 이 스킬의 기능과 사용 상황을 설명합니다.
---

# My Skill Name

여기에 Claude가 따라야 할 지침을 작성하세요.

## Workflow
- Step 1
- Step 2
```

### 3. 스킬 업로드 및 사용

1. 스킬 폴더를 ZIP 파일로 압축
2. Claude 인터페이스의 "Capabilities" 메뉴에서 업로드
3. 사용 시: "방금 만든 'my-skill-name' 스킬을 사용하여 이 작업을 수행해줘"

## 예시 스킬 구조

이 프로젝트에는 다양한 예시 스킬이 포함되어 있습니다. 각 스킬 폴더를 참고하세요.

