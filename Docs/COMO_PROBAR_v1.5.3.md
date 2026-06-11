# Cómo probar TransferLAN+ v1.5.3-beta

## Prueba

1. Instalar APK.
2. Conectar con Windows.
3. Enviar archivo grande.
4. La notificación debe mostrar progreso.
5. La pantalla Android también debe actualizar progreso del servicio.
6. Salir y volver a la app durante el envío.
7. Probar botón `Cancelar transferencia`.

## Resultado esperado

- La Activity muestra estados reales del TransferService.
- Se ve progreso en notificación y pantalla.
- Cancelar corta la transferencia.
