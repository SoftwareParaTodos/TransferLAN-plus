@echo off
title TransferLAN+ - Liberar puerto 5050
echo Este script intenta cerrar procesos que esten usando el puerto 5050.
echo Ejecutar como administrador si hace falta.
echo.
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :5050') do (
    echo Cerrando PID %%a
    taskkill /PID %%a /F
)
echo.
echo Listo. Volve a iniciar TransferLAN+.
pause
