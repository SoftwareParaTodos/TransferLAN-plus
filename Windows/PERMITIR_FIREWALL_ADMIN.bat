@echo off
title TransferLAN+ Firewall Assistant
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: ejecutar como administrador.
    pause
    exit /b 1
)
powershell -ExecutionPolicy Bypass -Command "Remove-NetFirewallRule -DisplayName 'TransferLAN+ Server 5050' -ErrorAction SilentlyContinue; New-NetFirewallRule -DisplayName 'TransferLAN+ Server 5050' -Direction Inbound -Protocol TCP -LocalPort 5050 -Action Allow -Profile Any"
echo Regla de Firewall creada para TransferLAN+ puerto 5050.
pause
