@echo off
title TransferLAN+ Firewall Assistant
echo TransferLAN+ necesita permitir:
echo - TCP 5050 para transferencias
echo - UDP 5050 para descubrimiento automatico
echo.

net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: No estas ejecutando como administrador.
    echo Clic derecho sobre este archivo ^> Ejecutar como administrador.
    pause
    exit /b 1
)

powershell -ExecutionPolicy Bypass -Command "Remove-NetFirewallRule -DisplayName 'TransferLAN+ Server TCP 5050' -ErrorAction SilentlyContinue; Remove-NetFirewallRule -DisplayName 'TransferLAN+ Discovery UDP 5050' -ErrorAction SilentlyContinue; New-NetFirewallRule -DisplayName 'TransferLAN+ Server TCP 5050' -Direction Inbound -Protocol TCP -LocalPort 5050 -Action Allow -Profile Any; New-NetFirewallRule -DisplayName 'TransferLAN+ Discovery UDP 5050' -Direction Inbound -Protocol UDP -LocalPort 5050 -Action Allow -Profile Any"

echo.
echo Reglas de Firewall creadas.
pause
