# Cómo probar TransferLAN+ v1.5.1-beta

## Prueba principal

1. Instalar APK.
2. Abrir TransferLAN+.
3. Conectar con la PC.
4. Enviar un archivo.
5. Debe aparecer notificación persistente de TransferLAN+.
6. La transferencia sigue usando el flujo estable actual.
7. Al terminar, debe cerrarse/liberarse el servicio preparado.

## Qué valida esta versión

- Manifest acepta ForegroundService.
- La app compila con `TransferService`.
- Android permite iniciar el servicio.
- WakeLock está preparado.
- Acción cancelar está preparada.

## Próxima fase

v1.5.2-beta moverá el upload completo al servicio.
