# han2f12

Windows Terminal에서 한/영 키를 F12로 변환해주는 백그라운드 유틸리티입니다.

Neovim + [Korean-IME.nvim](https://github.com/kiyoon/Korean-IME.nvim) 사용 시 한/영 키로 자연스럽게 한영 전환을 할 수 있게 해줍니다.

## 배경

Korean-IME.nvim은 시스템 IME 대신 Neovim 내부에서 한글 변환을 처리합니다. 플러그인 내에서 한영 전환은 F12로 동작하는데, han2f12는 Windows Terminal이 포커스일 때 한/영 키 입력을 F12로 변환하여 기존 습관대로 한/영 키를 쓸 수 있게 합니다.

## 설치

[Releases](../../releases)에서 `han2f12.exe`를 다운로드하세요.

## 사용법

```
han2f12.exe
```

실행 후 Windows Terminal에서 한/영 키를 누르면 F12로 변환됩니다. Enter 또는 Ctrl+C로 종료합니다.

시작 프로그램에 등록하려면 `shell:startup` 폴더에 exe를 넣으세요.

## 빌드

```
go build -ldflags "-s -w" -o han2f12.exe .
```

## 요구사항

- Windows Terminal
- [Korean-IME.nvim](https://github.com/kiyoon/Korean-IME.nvim) (F12 한영 전환 설정)

## 라이선스

MIT
