@echo off
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: ejecutar como administrador.
    pause
    exit /b 1
)
powershell -ExecutionPolicy Bypass -Command "Remove-NetFirewallRule -DisplayName 'TransferLAN+ Server 5050' -ErrorAction SilentlyContinue"
echo Regla eliminada.
pause
