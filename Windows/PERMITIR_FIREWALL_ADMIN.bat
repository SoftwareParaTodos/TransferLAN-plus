@echo off
title TransferLAN+ Firewall
netsh advfirewall firewall delete rule name="TransferLAN+ Server TCP 5050" >nul 2>&1
netsh advfirewall firewall delete rule name="TransferLAN+ Discovery UDP 5050" >nul 2>&1
netsh advfirewall firewall add rule name="TransferLAN+ Server TCP 5050" dir=in action=allow protocol=TCP localport=5050 profile=any
netsh advfirewall firewall add rule name="TransferLAN+ Discovery UDP 5050" dir=in action=allow protocol=UDP localport=5050 profile=any
echo Listo.
pause
