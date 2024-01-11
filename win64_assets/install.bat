@echo off

set p=%USERPROFILE%\.mkvtool\deps

rd /s/q %p%

reg query HKCU\Environment /v Path | findstr %p:\=\\% >Nul || reg add HKCU\Environment /f /v Path /t REG_EXPAND_SZ /d "%Path%;%p%;"
xcopy /y/s/i * %p%
rd /s/q %p%\fonttools
del %p%\install.bat