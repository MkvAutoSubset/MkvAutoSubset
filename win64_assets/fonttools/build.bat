@echo off

set p=..\_internal

rd /s/q %p%
del ..\ttx.exe ..\pyftsubset.exe

pyinstaller pyftsubset.py
pyinstaller ttx.py

set _p=dist\pyftsubset\_internal
xcopy /s/i %_p%\fontTools %p%\fontTools

move %_p%\_socket.pyd %p%
move %_p%\pyexpat.pyd %p%
move %_p%\unicodedata.pyd %p%
move %_p%\select.pyd %p%
move %_p%\base_library.zip %p%
move %_p%\python311.dll %p%

move dist\pyftsubset\pyftsubset.exe ..
move dist\ttx\ttx.exe ..

del *.spec
rd /s/q dist build