@echo off
title TransferLAN+ v1.1.0-beta
cd /d "%~dp0..\core\server"
echo Iniciando TransferLAN+ v1.1.0-beta...
echo.
echo Si Android no detecta esta PC:
echo - cerrar esta ventana
echo - ejecutar Windows\PERMITIR_FIREWALL_ADMIN.bat como administrador
echo - volver a iniciar TransferLAN+
echo.
TransferLAN+.exe
pause
