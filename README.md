# han2f12
[Korean-IME.nvim](https://github.com/kiyoon/Korean-IME.nvim)은 시스템 기본 IME 대신 Neovim 내부에서 영어 입력을 한글로 변환합니다. 그러나 F12와 F9가 각각 한영 키와 한자 키에 할당되어 있어 기존 습관대로 쓰기 불편합니다. han2f12는 윈도 터미널이 포커스를 받았을 때만 한영 키 입력과 한자 키 입력을 nvim상 대응하는 단축키인 F12와 F9으로 변환해서 기존 습관대로 자연스럽게 쓸 수 있도록 하면서도 nvim 이외의 다른 프로그램에는 영향을 주지 않습니다.

## 설치 및 사용법
[Releases](../../releases)에서 `han2f12.exe`를 다운로드하고 실행시켜 두기만 하면 윈도 터미널에 대한 입력이 발생할 때마다 한영 키를 F12로, 한자 키를 F9로 자동변환합니다. 종료하면 원래 상태로 복귀됩니다.
- 실행 프로그램을 시작 프로그램(`shell:startup`) 폴더에 넣어 두면 윈도 부팅시 자동 실행되도록 할 수 있습니다.
- 콘솔창을 숨기려면 `conhost.exe --headless han2f12.exe`와 같이 실행하는 방법이 있습니다. 이 경우 작업관리자 등으로 종료할 수 있습니다.

## 요구사항
- Windows Terminal
- [Korean-IME.nvim](https://github.com/kiyoon/Korean-IME.nvim) (한글 입력기 플러그인)

