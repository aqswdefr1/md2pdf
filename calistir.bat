@echo off
title md2pdf Interactive Interface
echo =================================================
echo            md2pdf Baslatiliyor...
echo =================================================
echo.

where go >nul 2>nul
if %errorlevel% neq 0 (
    if exist "C:\Program Files\Go\bin\go.exe" (
        set "PATH=%PATH%;C:\Program Files\Go\bin"
    ) else (
        echo Hata: Go kurulumu bulunamadi!
        echo Lutfen Go'nun yuklu oldugundan emin olun.
        echo.
        pause
        exit /b
    )
)

go run .

echo.
echo =================================================
echo Program sonlandirildi. Kapatmak icin bir tusa basin.
echo =================================================
pause > nul
